package processor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestIntegration_BasicTable(t *testing.T) {
	input := loadTestFixture(t, "basic_table_input.sql")
	expected := loadTestFixture(t, "basic_table_expected.sql")
	result := Process(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_FKCheck(t *testing.T) {
	input := loadTestFixture(t, "fk_check_input.sql")
	expected := loadTestFixture(t, "fk_check_expected.sql")
	result := Process(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
	assert.Contains(t, result, "REFERENCES public.orders(id)")
}

func TestIntegration_Sequences(t *testing.T) {
	input := loadTestFixture(t, "sequences_input.sql")
	expected := loadTestFixture(t, "sequences_expected.sql")
	result := Process(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_FKCascade(t *testing.T) {
	input := loadTestFixture(t, "fk_cascade_input.sql")
	expected := loadTestFixture(t, "fk_cascade_expected.sql")
	result := Process(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_Types(t *testing.T) {
	input := loadTestFixture(t, "types_input.sql")
	expected := loadTestFixture(t, "types_expected.sql")
	result := Process(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_AlterColumn(t *testing.T) {
	input := loadTestFixture(t, "alter_column_input.sql")
	expected := loadTestFixture(t, "alter_column_expected.sql")
	result := Process(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_SelfReferential(t *testing.T) {
	input := loadTestFixture(t, "self_referential_input.sql")
	expected := loadTestFixture(t, "self_referential_expected.sql")
	result := Process(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_PartialIndexes(t *testing.T) {
	input := loadTestFixture(t, "partial_indexes_input.sql")
	expected := loadTestFixture(t, "partial_indexes_expected.sql")
	result := Process(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_MaterializedViews(t *testing.T) {
	input := loadTestFixture(t, "materialized_views_input.sql")
	expected := loadTestFixture(t, "materialized_views_expected.sql")
	result := Process(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_Collation(t *testing.T) {
	input := loadTestFixture(t, "collation_input.sql")
	expected := loadTestFixture(t, "collation_expected.sql")
	result := Process(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_GeneratedColumns(t *testing.T) {
	input := loadTestFixture(t, "generated_columns_input.sql")
	expected := loadTestFixture(t, "generated_columns_expected.sql")
	result := Process(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_DomainTypes(t *testing.T) {
	input := loadIntegrationFixture(t, "domain_types_input.sql")
	expected := loadIntegrationFixture(t, "domain_types_expected.sql")
	result := Process(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_Schemas(t *testing.T) {
	input := loadIntegrationFixture(t, "schemas_input.sql")
	expected := loadIntegrationFixture(t, "schemas_expected.sql")
	result := Process(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_Inheritance(t *testing.T) {
	input := loadIntegrationFixture(t, "inheritance_input.sql")
	expected := loadIntegrationFixture(t, "inheritance_expected.sql")
	result := Process(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_IdentityColumns(t *testing.T) {
	input := loadTestFixture(t, "identity_columns_input.sql")
	expected := loadTestFixture(t, "identity_columns_expected.sql")
	result := Process(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_ExclusionConstraints(t *testing.T) {
	input := loadTestFixture(t, "exclusion_constraints_input.sql")
	expected := loadTestFixture(t, "exclusion_constraints_expected.sql")
	result := Process(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_PartitionedTables(t *testing.T) {
	input := loadTestFixture(t, "partitioned_tables_input.sql")
	expected := loadTestFixture(t, "partitioned_tables_expected.sql")
	result := Process(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_FKColumnMapping(t *testing.T) {
	input := loadTestFixture(t, "fk_column_mapping_input.sql")
	expected := loadTestFixture(t, "fk_column_mapping_expected.sql")
	result := Process(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_RangeTypes(t *testing.T) {
	input := loadTestFixture(t, "range_types_input.sql")
	expected := loadTestFixture(t, "range_types_expected.sql")
	result := Process(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_FKNoAction(t *testing.T) {
	input := loadTestFixture(t, "fk_no_action_input.sql")
	expected := loadTestFixture(t, "fk_no_action_expected.sql")
	result := Process(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_Triggers(t *testing.T) {
	input := loadTestFixture(t, "triggers_input.sql")
	expected := loadTestFixture(t, "triggers_expected.sql")
	result := Process(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_RLSPolicies(t *testing.T) {
	input := loadTestFixture(t, "rls_policies_input.sql")
	expected := loadTestFixture(t, "rls_policies_expected.sql")
	result := Process(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_Functions(t *testing.T) {
	input := loadTestFixture(t, "functions_input.sql")
	expected := loadTestFixture(t, "functions_expected.sql")
	result := Process(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestPassthrough_DropTableIfExists(t *testing.T) {
	sql := "DROP TABLE IF EXISTS foo;"
	result := Process(sql)
	assert.Contains(t, result, "DROP TABLE IF EXISTS foo")
}

func TestPassthrough_CreateSequence(t *testing.T) {
	sql := "CREATE SEQUENCE foo_seq;"
	result := Process(sql)
	assert.Contains(t, result, "CREATE SEQUENCE foo_seq")
}

func TestPassthrough_CreateIndex(t *testing.T) {
	sql := "CREATE INDEX idx ON foo (bar);"
	result := Process(sql)
	assert.Contains(t, result, "CREATE INDEX idx ON foo")
}

func TestPassthrough_CreateType(t *testing.T) {
	sql := "CREATE TYPE foo_type AS ENUM ('a', 'b');"
	result := Process(sql)
	assert.Contains(t, result, "CREATE TYPE foo_type")
}

func TestFK_WithCascadeActions(t *testing.T) {
	input := `
CREATE TABLE users (id bigint NOT NULL);
CREATE TABLE orders (id bigint NOT NULL, user_id bigint NOT NULL);
ALTER TABLE orders ADD CONSTRAINT orders_user_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
`
	result := Process(input)
	assert.Contains(t, result, "ON DELETE CASCADE")
	assert.Contains(t, result, "REFERENCES users(id) ON DELETE CASCADE")
}

func TestFK_WithMultipleActions(t *testing.T) {
	input := `
CREATE TABLE users (id bigint NOT NULL);
CREATE TABLE posts (id bigint NOT NULL, author_id bigint NOT NULL);
ALTER TABLE posts ADD CONSTRAINT posts_author_fkey FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE RESTRICT ON UPDATE CASCADE;
`
	result := Process(input)
	assert.Contains(t, result, "ON DELETE RESTRICT")
	assert.Contains(t, result, "ON UPDATE CASCADE")
}

func TestRobust_DomainTypesOrdering(t *testing.T) {
	input := loadIntegrationFixture(t, "domain_types_input.sql")
	result := Process(input)

	assertTypeBeforeTable(t, result, "email", "users")
	assertTypeBeforeTable(t, result, "positive_int", "orders")
	assertTypeBeforeTable(t, result, "order_status", "orders")

	assert.Contains(t, result, "CREATE DOMAIN public.email")
	assert.Contains(t, result, "CREATE DOMAIN public.positive_int")
	assert.Contains(t, result, "CREATE DOMAIN public.phone_number")
	assert.Contains(t, result, "CREATE DOMAIN public.order_status")

	assert.Contains(t, result, "email public.email NOT NULL UNIQUE")
	assert.Contains(t, result, "quantity public.positive_int NOT NULL")
	assert.Contains(t, result, "status public.order_status NOT NULL DEFAULT 'pending'")
}

func TestRobust_SchemasOrdering(t *testing.T) {
	input := loadIntegrationFixture(t, "schemas_input.sql")
	result := Process(input)

	assert.Contains(t, result, "CREATE TABLE app.users")
	assert.Contains(t, result, "CREATE TABLE app.orders")
	assert.Contains(t, result, "CREATE TABLE audit.audit_logs")
	assert.Contains(t, result, "CREATE TABLE public.countries")
	assert.Contains(t, result, "CREATE TABLE finance.invoices")

	assert.Contains(t, result, "user_id bigint REFERENCES app.users")
	assert.Contains(t, result, "user_id bigint REFERENCES app.users")
	assert.Contains(t, result, "order_id bigint REFERENCES app.orders")
}

func TestRobust_InheritancePassthrough(t *testing.T) {
	input := loadIntegrationFixture(t, "inheritance_input.sql")
	result := Process(input)

	assert.Contains(t, result, "CREATE TABLE public.users")
	assert.Contains(t, result, "CREATE TABLE public.administrators")
	assert.Contains(t, result, "CREATE TABLE public.moderators")
	assert.Contains(t, result, "CREATE TABLE public.registered_users")

	assert.Contains(t, result, "INHERITS (public.users)")
}

func TestRobust_ComplexSchemaFeatures(t *testing.T) {
	input := loadTestFixture(t, "complex_schema_input.sql")
	result := Process(input)

	assertStatementContains(t, result, "CREATE TYPE", "ENUM types")
	assertViewPreserved(t, result, "active_users")
	assertViewPreserved(t, result, "order_summary")
	assertViewPreserved(t, result, "product_inventory_summary")
	assertMaterializedViewPreserved(t, result, "order_stats")
	assertMaterializedViewPreserved(t, result, "user_order_summary")
	assertIndexPreserved(t, result, "idx_users_organization_id")
	assertIndexPreserved(t, result, "idx_orders_user_id")
	assertStatementContains(t, result, "CREATE TRIGGER", "Triggers")
	assertStatementContains(t, result, "ENABLE ROW LEVEL SECURITY", "RLS")
}

func TestRobust_AlterColumnChanges(t *testing.T) {
	input := loadTestFixture(t, "alter_column_input.sql")
	result := Process(input)

	assertStatementContains(t, result, "ALTER TABLE", "ALTER TABLE statements")
	assert.Contains(t, result, "ALTER TABLE public.users ALTER COLUMN email DROP NOT NULL")
	assert.Contains(t, result, "ALTER TABLE public.logs ALTER COLUMN severity DROP NOT NULL")
}

func TestRobust_SequenceHandling(t *testing.T) {
	input := loadTestFixture(t, "sequences_input.sql")
	result := Process(input)

	assert.Contains(t, result, "BIGSERIAL")
	assertStatementContains(t, result, "CREATE SEQUENCE global_id_seq", "Shared sequence preserved")
	assertStatementContains(t, result, "CREATE SEQUENCE public.tags_id_seq", "Standalone sequence preserved")

	assertTypeBeforeTable(t, result, "global_id_seq", "orders")
}

func TestEdgeCases_EmptyAndComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   \n\t\n   ",
			expected: "",
		},
		{
			name:     "only line comments",
			input:    "-- comment\n-- another comment",
			expected: "",
		},
		{
			name:     "only block comment",
			input:    "/* this is a block comment */",
			expected: "/* this is a block comment */",
		},
		{
			name:     "mixed comments and whitespace",
			input:    "-- header\n\n/* body */\n\nSELECT 1;",
			expected: "/* body */\n\nSELECT 1;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Process(tt.input)
			assert.Equal(t, tt.expected, strings.TrimSpace(result))
		})
	}
}
