package processor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadE2EFixture(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "e2e", name)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(data)
}

func loadFixtureFile(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "fixtures", name)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(data)
}

func normalizeSQL(sql string) string {
	sql = strings.ReplaceAll(sql, "\r\n", "\n")
	return strings.TrimSpace(sql)
}

func extractStatements(sql string) []string {
	var statements []string
	var current strings.Builder
	lines := strings.Split(sql, "\n")
	inBlockComment := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "/*") {
			inBlockComment = true
		}
		if strings.HasSuffix(trimmed, "*/") {
			inBlockComment = false
			continue
		}
		if inBlockComment || trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}

		current.WriteString(line)
		current.WriteString("\n")

		if strings.HasSuffix(strings.TrimSpace(line), ";") {
			stmt := strings.TrimSpace(current.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
		}
	}

	return statements
}

func assertStatementContains(t *testing.T, sql, keyword, context string) {
	t.Helper()
	statements := extractStatements(sql)
	found := false
	for _, stmt := range statements {
		if strings.Contains(stmt, keyword) {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected to find statement containing %s for %s", keyword, context)
}

func assertStatementOrder(t *testing.T, sql, firstKeyword, secondKeyword string) {
	t.Helper()
	statements := extractStatements(sql)
	firstIdx := -1
	secondIdx := -1

	for i, stmt := range statements {
		if strings.Contains(stmt, firstKeyword) && firstIdx == -1 {
			firstIdx = i
		}
		if strings.Contains(stmt, secondKeyword) && secondIdx == -1 {
			secondIdx = i
		}
	}

	if firstIdx >= 0 && secondIdx >= 0 {
		assert.Less(t, firstIdx, secondIdx, "Expected %s to come before %s", firstKeyword, secondKeyword)
	}
}

func assertTypeBeforeTable(t *testing.T, sql, typeName, tableName string) {
	t.Helper()
	assertStatementOrder(t, sql, "CREATE TYPE "+typeName, "CREATE TABLE "+tableName)
}

func assertViewPreserved(t *testing.T, sql, viewName string) {
	t.Helper()
	assertStatementContains(t, sql, "CREATE VIEW "+viewName, "View "+viewName)
}

func assertMaterializedViewPreserved(t *testing.T, sql, viewName string) {
	t.Helper()
	assertStatementContains(t, sql, "CREATE MATERIALIZED VIEW "+viewName, "Materialized view "+viewName)
}

func assertIndexPreserved(t *testing.T, sql, indexName string) {
	t.Helper()
	assertStatementContains(t, sql, "CREATE INDEX "+indexName, "Index "+indexName)
}

func TestFixture_Gated_DomainTypes(t *testing.T) {
	input := loadFixtureFile(t, "domain_types_input.sql")
	expected := loadFixtureFile(t, "domain_types_expected.sql")
	result := processExperimental(input)
	assert.NotContains(t, result, "CREATE DOMAIN")
	assert.Contains(t, result, "email public.email")
	assert.Contains(t, result, "CREATE DOMAIN public.email")
	assert.Contains(t, result, "CREATE DOMAIN public.positive_int")
	assert.Contains(t, result, "email", "users")
	assertTypeBeforeTable(t, result, "positive_int", "orders")
	assertTypeBeforeTable(t, result, "order_status", "orders")
	assert.Contains(t, result, "CREATE DOMAIN public.email")
	assert.Contains(t, result, "CREATE DOMAIN public.positive_int")
	assert.Contains(t, result, "CREATE DOMAIN public.phone_number")
	assert.Contains(t, result, "CREATE DOMAIN public.order_status")
	assert.Contains(t, result, "email public.email NOT NULL UNIQUE")
	assert.Contains(t, result, "quantity public.positive_int NOT NULL")
	assert.Contains(t, result, "status public.order_status NOT NULL DEFAULT 'pending'")

	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Gated_Inheritance(t *testing.T) {
	input := loadFixtureFile(t, "inheritance_input.sql")
	expected := loadFixtureFile(t, "inheritance_expected.sql")
	result := processExperimental(input)
	assert.Contains(t, result, "CREATE TABLE public.users")
	assert.Contains(t, result, "CREATE TABLE public.administrators")
	assert.Contains(t, result, "CREATE TABLE public.moderators")
	assert.Contains(t, result, "CREATE TABLE public.registered_users")
	assert.NotContains(t, result, "INHERITS")
	assert.Contains(t, result, "CREATE TABLE public.users")
	assert.Contains(t, result, "CREATE TABLE public.administrators")
	assert.Contains(t, result, "INHERITS (public.users)")

	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Gated_PartitionedTables(t *testing.T) {
	input := loadFixtureFile(t, "partitioned_tables_input.sql")
	expected := loadFixtureFile(t, "partitioned_tables_expected.sql")
	result := processExperimental(input)
	assert.NotContains(t, result, "PARTITION OF")
	assert.Contains(t, result, "PARTITION BY RANGE")
	assert.Contains(t, result, "PARTITION BY LIST")
	assert.Contains(t, result, "PARTITION BY HASH")
	assert.Contains(t, result, "PARTITION OF")
	assert.Contains(t, result, "PARTITION BY RANGE")

	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Gated_IfExistsForDrop(t *testing.T) {
	input := loadFixtureFile(t, "drop_statements_input.sql")
	expected := loadFixtureFile(t, "drop_statements_expected.sql")
	result := processExperimental(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}
