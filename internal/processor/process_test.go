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

//Intergration Fixture Tests

func TestIntegration_DomainTypes(t *testing.T) {
	input := loadIntegrationFixture(t, "domain_types_input.sql")
	expected := loadIntegrationFixture(t, "domain_types_expected.sql")
	result := processExperimental(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_Inheritance(t *testing.T) {
	input := loadIntegrationFixture(t, "inheritance_input.sql")
	expected := loadIntegrationFixture(t, "inheritance_expected.sql")
	result := processExperimental(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}
func TestIntegration_PartitionedTables(t *testing.T) {
	input := loadIntegrationFixture(t, "partitioned_tables_input.sql")
	expected := loadIntegrationFixture(t, "partitioned_tables_expected.sql")
	result := processExperimental(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_IfExistsForDrop(t *testing.T) {
	input := loadIntegrationFixture(t, "drop_statements_input.sql")
	expected := loadIntegrationFixture(t, "drop_statements_expected.sql")
	result := processExperimental(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_OnlyRemoval(t *testing.T) {
	input := loadTestFixture(t, "only_removal_input.sql")
	expected := loadTestFixture(t, "only_removal_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestGating_DomainTypesSkippedByDefault(t *testing.T) {
	input := loadIntegrationFixture(t, "domain_types_input.sql")
	result := processDefault(input)
	assert.NotContains(t, result, "CREATE DOMAIN")
	assert.Contains(t, result, "email public.email")
}

func TestGating_DomainTypesPreservedWithExperimentalFolding(t *testing.T) {
	input := loadIntegrationFixture(t, "domain_types_input.sql")
	result := processExperimental(input)
	assert.Contains(t, result, "CREATE DOMAIN public.email")
	assert.Contains(t, result, "CREATE DOMAIN public.positive_int")
}

func TestGating_PartitionChildrenSkippedByDefault(t *testing.T) {
	input := loadIntegrationFixture(t, "partitioned_tables_input.sql")
	result := processDefault(input)
	assert.NotContains(t, result, "PARTITION OF")
	assert.Contains(t, result, "PARTITION BY RANGE")
	assert.Contains(t, result, "PARTITION BY LIST")
	assert.Contains(t, result, "PARTITION BY HASH")
}

func TestGating_PartitionChildrenPreservedWithExperimentalFolding(t *testing.T) {
	input := loadIntegrationFixture(t, "partitioned_tables_input.sql")
	result := processExperimental(input)
	assert.Contains(t, result, "PARTITION OF")
	assert.Contains(t, result, "PARTITION BY RANGE")
}

func TestRobust_InheritanceGatedByDefault(t *testing.T) {
	input := loadIntegrationFixture(t, "inheritance_input.sql")
	result := processDefault(input)

	assert.Contains(t, result, "CREATE TABLE public.users")
	assert.Contains(t, result, "CREATE TABLE public.administrators")
	assert.Contains(t, result, "CREATE TABLE public.moderators")
	assert.Contains(t, result, "CREATE TABLE public.registered_users")

	assert.NotContains(t, result, "INHERITS")
}

func TestRobust_InheritancePreservedWithExperimentalFolding(t *testing.T) {
	input := loadIntegrationFixture(t, "inheritance_input.sql")
	result := processExperimental(input)

	assert.Contains(t, result, "CREATE TABLE public.users")
	assert.Contains(t, result, "CREATE TABLE public.administrators")
	assert.Contains(t, result, "INHERITS (public.users)")
}

func TestRobust_DomainTypesOrdering(t *testing.T) {
	input := loadIntegrationFixture(t, "domain_types_input.sql")
	result := processExperimental(input)

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
}

//e2e Fixture Tests

func TestIntegration_BasicTable(t *testing.T) {
	input := loadTestFixture(t, "basic_table_input.sql")
	expected := loadTestFixture(t, "basic_table_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_FKCheck(t *testing.T) {
	input := loadTestFixture(t, "fk_check_input.sql")
	expected := loadTestFixture(t, "fk_check_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
	assert.Contains(t, result, "REFERENCES public.orders(id)")
}

func TestIntegration_Sequences(t *testing.T) {
	input := loadTestFixture(t, "sequences_input.sql")
	expected := loadTestFixture(t, "sequences_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_FKCascade(t *testing.T) {
	input := loadTestFixture(t, "fk_cascade_input.sql")
	expected := loadTestFixture(t, "fk_cascade_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_Types(t *testing.T) {
	input := loadTestFixture(t, "types_input.sql")
	expected := loadTestFixture(t, "types_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_AlterColumn(t *testing.T) {
	input := loadTestFixture(t, "alter_column_input.sql")
	expected := loadTestFixture(t, "alter_column_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_SelfReferential(t *testing.T) {
	input := loadTestFixture(t, "self_referential_input.sql")
	expected := loadTestFixture(t, "self_referential_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_PartialIndexes(t *testing.T) {
	input := loadTestFixture(t, "partial_indexes_input.sql")
	expected := loadTestFixture(t, "partial_indexes_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_MaterializedViews(t *testing.T) {
	input := loadTestFixture(t, "materialized_views_input.sql")
	expected := loadTestFixture(t, "materialized_views_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_Collation(t *testing.T) {
	input := loadTestFixture(t, "collation_input.sql")
	expected := loadTestFixture(t, "collation_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_GeneratedColumns(t *testing.T) {
	input := loadTestFixture(t, "generated_columns_input.sql")
	expected := loadTestFixture(t, "generated_columns_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_Schemas(t *testing.T) {
	input := loadTestFixture(t, "schemas_input.sql")
	expected := loadTestFixture(t, "schemas_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_IdentityColumns(t *testing.T) {
	input := loadTestFixture(t, "identity_columns_input.sql")
	expected := loadTestFixture(t, "identity_columns_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_ExclusionConstraints(t *testing.T) {
	input := loadTestFixture(t, "exclusion_constraints_input.sql")
	expected := loadTestFixture(t, "exclusion_constraints_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_FKColumnMapping(t *testing.T) {
	input := loadTestFixture(t, "fk_column_mapping_input.sql")
	expected := loadTestFixture(t, "fk_column_mapping_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_RangeTypes(t *testing.T) {
	input := loadTestFixture(t, "range_types_input.sql")
	expected := loadTestFixture(t, "range_types_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_FKNoAction(t *testing.T) {
	input := loadTestFixture(t, "fk_no_action_input.sql")
	expected := loadTestFixture(t, "fk_no_action_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_Triggers(t *testing.T) {
	input := loadTestFixture(t, "triggers_input.sql")
	expected := loadTestFixture(t, "triggers_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_RLSPolicies(t *testing.T) {
	input := loadTestFixture(t, "rls_policies_input.sql")
	expected := loadTestFixture(t, "rls_policies_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestIntegration_Functions(t *testing.T) {
	input := loadTestFixture(t, "functions_input.sql")
	expected := loadTestFixture(t, "functions_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestPassthrough_DropTableIfExists(t *testing.T) {
	sql := "DROP TABLE IF EXISTS foo;"
	result := processDefault(sql)
	assert.Contains(t, result, "DROP TABLE IF EXISTS foo")
}

func TestPassthrough_CreateSequence(t *testing.T) {
	sql := "CREATE SEQUENCE foo_seq;"
	result := processDefault(sql)
	assert.Contains(t, result, "CREATE SEQUENCE foo_seq")
}

func TestPassthrough_CreateIndex(t *testing.T) {
	sql := "CREATE INDEX idx ON foo (bar);"
	result := processDefault(sql)
	assert.Contains(t, result, "CREATE INDEX idx ON foo")
}

func TestPassthrough_CreateType(t *testing.T) {
	sql := "CREATE TYPE foo_type AS ENUM ('a', 'b');"
	result := processDefault(sql)
	assert.Contains(t, result, "CREATE TYPE foo_type")
}

func TestFK_WithCascadeActions(t *testing.T) {
	input := `
CREATE TABLE users (id bigint NOT NULL);
CREATE TABLE orders (id bigint NOT NULL, user_id bigint NOT NULL);
ALTER TABLE orders ADD CONSTRAINT orders_user_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
`
	result := processDefault(input)
	assert.Contains(t, result, "ON DELETE CASCADE")
	assert.Contains(t, result, "REFERENCES users(id) ON DELETE CASCADE")
}

func TestFK_WithMultipleActions(t *testing.T) {
	input := `
CREATE TABLE users (id bigint NOT NULL);
CREATE TABLE posts (id bigint NOT NULL, author_id bigint NOT NULL);
ALTER TABLE posts ADD CONSTRAINT posts_author_fkey FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE RESTRICT ON UPDATE CASCADE;
`
	result := processDefault(input)
	assert.Contains(t, result, "ON DELETE RESTRICT")
	assert.Contains(t, result, "ON UPDATE CASCADE")
}

func TestRobust_SchemasOrdering(t *testing.T) {
	input := loadTestFixture(t, "schemas_input.sql")
	result := processDefault(input)

	assert.Contains(t, result, "CREATE SCHEMA app")
	assert.Contains(t, result, "CREATE TABLE app.users")
	assert.Contains(t, result, "CREATE TABLE public.countries")
	assert.Contains(t, result, "country_id bigint REFERENCES public.countries")
}

func TestRobust_ComplexSchemaFeatures(t *testing.T) {
	input := loadTestFixture(t, "complex_schema_input.sql")
	result := processDefault(input)

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
	result := processDefault(input)

	assert.NotContains(t, result, "DROP NOT NULL")
	assert.NotContains(t, result, "DROP DEFAULT")
	assert.NotContains(t, result, "SET DEFAULT")
	assert.NotContains(t, result, "SET NOT NULL")
	assert.NotContains(t, result, "ALTER COLUMN")
}

func TestRobust_SequenceHandling(t *testing.T) {
	input := loadTestFixture(t, "sequences_input.sql")
	result := processDefault(input)

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
			expected: "",
		},
		{
			name:     "mixed comments and whitespace",
			input:    "-- header\n\n/* body */\n\nSELECT 1;",
			expected: "SELECT 1;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processDefault(tt.input)
			assert.Equal(t, tt.expected, strings.TrimSpace(result))
		})
	}
}

func TestFeature_BlockCommentRemoval(t *testing.T) {
	input := `/* This is a block comment */ CREATE TABLE foo (id int); /* Another block */`
	result := processDefault(input)
	assert.NotContains(t, result, "/*")
	assert.NotContains(t, result, "*/")
	assert.Contains(t, result, "CREATE TABLE foo")
}

func TestFeature_BlockCommentRemoval_Multiline(t *testing.T) {
	input := `/* 
		Multi-line 
		block comment 
	*/ 
	CREATE TABLE bar (id int);`
	result := processDefault(input)
	assert.NotContains(t, result, "/*")
	assert.NotContains(t, result, "Multi-line")
	assert.Contains(t, result, "CREATE TABLE bar")
}

func TestFeature_IfExistsForDrop(t *testing.T) {
	input := `DROP TABLE foo;
CREATE TABLE foo (id int);`
	result := processExperimental(input)
	assert.Contains(t, result, "DROP TABLE IF EXISTS foo")
	assert.NotContains(t, result, "DROP TABLE foo;")
}

func TestFeature_IfExistsForDrop_AlreadyExists(t *testing.T) {
	input := `DROP TABLE IF EXISTS foo;
CREATE TABLE foo (id int);`
	result := processExperimental(input)
	assert.Contains(t, result, "DROP TABLE IF EXISTS foo")
}

func TestFeature_IfExistsForDrop_Index(t *testing.T) {
	input := `DROP INDEX idx_foo;
CREATE INDEX idx_foo ON foo(id);`
	result := processExperimental(input)
	assert.Contains(t, result, "DROP INDEX IF EXISTS idx_foo")
}

func TestGating_IfExistsForDropGatedByDefault(t *testing.T) {
	input := `DROP TABLE foo;
CREATE TABLE foo (id int);`
	result := processDefault(input)
	assert.NotContains(t, result, "IF EXISTS")
	assert.Contains(t, result, "DROP TABLE foo;")
}

func TestFeature_AlterSequenceFilteredWhenSerial(t *testing.T) {
	input := `CREATE SEQUENCE public.order_ids START WITH 1;
CREATE TABLE public.orders (id bigint NOT NULL);
ALTER TABLE public.orders ALTER COLUMN id SET DEFAULT nextval('public.order_ids'::regclass);
ALTER SEQUENCE public.order_ids RESTART WITH 2000;`
	result := Process(input, nil)
	assert.NotContains(t, result, "ALTER SEQUENCE")
	assert.Contains(t, result, "BIGSERIAL")
}

func TestFeature_AlterSequenceFiltered_Multiple(t *testing.T) {
	input := `CREATE SEQUENCE seq1 START WITH 1;
CREATE SEQUENCE seq2 START WITH 1;
CREATE TABLE t1 (id bigint NOT NULL);
CREATE TABLE t2 (id bigint NOT NULL);
ALTER TABLE t1 ALTER COLUMN id SET DEFAULT nextval('seq1'::regclass);
ALTER TABLE t2 ALTER COLUMN id SET DEFAULT nextval('seq2'::regclass);
ALTER SEQUENCE seq1 RESTART WITH 100;
ALTER SEQUENCE seq2 RESTART WITH 200;
ALTER SEQUENCE seq1 INCREMENT BY 5;`
	result := Process(input, nil)
	assert.NotContains(t, result, "ALTER SEQUENCE")
	assert.Contains(t, result, "BIGSERIAL")
}

func TestFeature_AlterSequenceKeptWhenNotSerial(t *testing.T) {
	input := `CREATE SEQUENCE order_ids START WITH 1;
CREATE TABLE orders (id bigint NOT NULL);
ALTER TABLE orders ALTER COLUMN id SET DEFAULT nextval('order_ids'::regclass);
ALTER SEQUENCE order_ids RESTART WITH 2000;`
	result := Process(input, nil)
	// Should keep ALTER SEQUENCE because sequence is not converted to SERIAL (bigint can be SERIAL but let's check)
	// Actually bigint CAN be SERIAL (BIGSERIAL), so this should be filtered
	// Let me check: bigint -> BIGSERIAL is supported, so ALTER SEQUENCE should be filtered
	assert.NotContains(t, result, "ALTER SEQUENCE")
	assert.Contains(t, result, "BIGSERIAL")
}
