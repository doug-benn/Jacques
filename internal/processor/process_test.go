package processor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func processDefault(sql string) string {
	return Process(sql, nil)
}

func processExperimental(sql string) string {
	return Process(sql, &Options{ExperimentalFolding: true})
}

func loadTestFixture(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "e2e", name)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(data)
}

func loadIntegrationFixture(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "integration", name)
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

func TestNormalizeSequenceName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no schema",
			input: "my_seq",
			want:  "my_seq",
		},
		{
			name:  "with schema",
			input: "public.my_seq",
			want:  "my_seq",
		},
		{
			name:  "nested schema",
			input: "app.public.my_seq",
			want:  "my_seq",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeSequenceName(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}
