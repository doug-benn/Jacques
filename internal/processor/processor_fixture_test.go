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

func TestFixture_DomainTypes(t *testing.T) {
	input := loadFixtureFile(t, "domain_types_input.sql")
	expected := loadFixtureFile(t, "domain_types_expected.sql")
	result := processExperimental(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_E2E_Inheritance(t *testing.T) {
	input := loadFixtureFile(t, "inheritance_input.sql")
	expected := loadFixtureFile(t, "inheritance_expected.sql")
	result := processExperimental(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_E2E_PartitionedTables(t *testing.T) {
	input := loadFixtureFile(t, "partitioned_tables_input.sql")
	expected := loadFixtureFile(t, "partitioned_tables_expected.sql")
	result := processExperimental(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_E2E_IfExistsForDrop(t *testing.T) {
	input := loadFixtureFile(t, "drop_statements_input.sql")
	expected := loadFixtureFile(t, "drop_statements_expected.sql")
	result := processExperimental(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_E2E_OnlyRemoval(t *testing.T) {
	input := loadE2EFixture(t, "only_removal_input.sql")
	expected := loadE2EFixture(t, "only_removal_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Gating_DomainTypesSkippedByDefault(t *testing.T) {
	input := loadFixtureFile(t, "domain_types_input.sql")
	result := processDefault(input)
	assert.NotContains(t, result, "CREATE DOMAIN")
	assert.Contains(t, result, "email public.email")
}

func TestFixture_Gating_DomainTypesPreservedWithExperimentalFolding(t *testing.T) {
	input := loadFixtureFile(t, "domain_types_input.sql")
	result := processExperimental(input)
	assert.Contains(t, result, "CREATE DOMAIN public.email")
	assert.Contains(t, result, "CREATE DOMAIN public.positive_int")
}

func TestFixture_Gating_PartitionChildrenSkippedByDefault(t *testing.T) {
	input := loadFixtureFile(t, "partitioned_tables_input.sql")
	result := processDefault(input)
	assert.NotContains(t, result, "PARTITION OF")
	assert.Contains(t, result, "PARTITION BY RANGE")
	assert.Contains(t, result, "PARTITION BY LIST")
	assert.Contains(t, result, "PARTITION BY HASH")
}

func TestFixture_Gating_PartitionChildrenPreservedWithExperimentalFolding(t *testing.T) {
	input := loadFixtureFile(t, "partitioned_tables_input.sql")
	result := processExperimental(input)
	assert.Contains(t, result, "PARTITION OF")
	assert.Contains(t, result, "PARTITION BY RANGE")
}

func TestFixture_Gating_IfExistsForDropGatedByDefault(t *testing.T) {
	input := `DROP TABLE foo;
CREATE TABLE foo (id int);`
	result := processDefault(input)
	assert.NotContains(t, result, "IF EXISTS")
	assert.Contains(t, result, "DROP TABLE foo;")
}

func TestFixture_Robust_InheritanceGatedByDefault(t *testing.T) {
	input := loadFixtureFile(t, "inheritance_input.sql")
	result := processDefault(input)

	assert.Contains(t, result, "CREATE TABLE public.users")
	assert.Contains(t, result, "CREATE TABLE public.administrators")
	assert.Contains(t, result, "CREATE TABLE public.moderators")
	assert.Contains(t, result, "CREATE TABLE public.registered_users")

	assert.NotContains(t, result, "INHERITS")
}

func TestFixture_Robust_InheritancePreservedWithExperimentalFolding(t *testing.T) {
	input := loadFixtureFile(t, "inheritance_input.sql")
	result := processExperimental(input)

	assert.Contains(t, result, "CREATE TABLE public.users")
	assert.Contains(t, result, "CREATE TABLE public.administrators")
	assert.Contains(t, result, "INHERITS (public.users)")
}

func TestFixture_Robust_DomainTypesOrdering(t *testing.T) {
	input := loadFixtureFile(t, "domain_types_input.sql")
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

func TestFixture_E2E_FKCascade(t *testing.T) {
	input := loadE2EFixture(t, "fk_cascade_input.sql")
	expected := loadE2EFixture(t, "fk_cascade_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_E2E_Types(t *testing.T) {
	input := loadE2EFixture(t, "types_input.sql")
	expected := loadE2EFixture(t, "types_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_E2E_AlterColumn(t *testing.T) {
	input := loadE2EFixture(t, "alter_column_input.sql")
	expected := loadE2EFixture(t, "alter_column_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_E2E_SelfReferential(t *testing.T) {
	input := loadE2EFixture(t, "self_referential_input.sql")
	expected := loadE2EFixture(t, "self_referential_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_E2E_PartialIndexes(t *testing.T) {
	input := loadE2EFixture(t, "partial_indexes_input.sql")
	expected := loadE2EFixture(t, "partial_indexes_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_E2E_MaterializedViews(t *testing.T) {
	input := loadE2EFixture(t, "materialized_views_input.sql")
	expected := loadE2EFixture(t, "materialized_views_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_E2E_Collation(t *testing.T) {
	input := loadE2EFixture(t, "collation_input.sql")
	expected := loadE2EFixture(t, "collation_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_E2E_GeneratedColumns(t *testing.T) {
	input := loadE2EFixture(t, "generated_columns_input.sql")
	expected := loadE2EFixture(t, "generated_columns_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_E2E_Schemas(t *testing.T) {
	input := loadE2EFixture(t, "schemas_input.sql")
	expected := loadE2EFixture(t, "schemas_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_E2E_IdentityColumns(t *testing.T) {
	input := loadE2EFixture(t, "identity_columns_input.sql")
	expected := loadE2EFixture(t, "identity_columns_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_E2E_ExclusionConstraints(t *testing.T) {
	input := loadE2EFixture(t, "exclusion_constraints_input.sql")
	expected := loadE2EFixture(t, "exclusion_constraints_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_E2E_FKColumnMapping(t *testing.T) {
	input := loadE2EFixture(t, "fk_column_mapping_input.sql")
	expected := loadE2EFixture(t, "fk_column_mapping_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_E2E_RangeTypes(t *testing.T) {
	input := loadE2EFixture(t, "range_types_input.sql")
	expected := loadE2EFixture(t, "range_types_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_E2E_FKNoAction(t *testing.T) {
	input := loadE2EFixture(t, "fk_no_action_input.sql")
	expected := loadE2EFixture(t, "fk_no_action_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_E2E_Triggers(t *testing.T) {
	input := loadE2EFixture(t, "triggers_input.sql")
	expected := loadE2EFixture(t, "triggers_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_E2E_RLSPolicies(t *testing.T) {
	input := loadE2EFixture(t, "rls_policies_input.sql")
	expected := loadE2EFixture(t, "rls_policies_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_E2E_Functions(t *testing.T) {
	input := loadE2EFixture(t, "functions_input.sql")
	expected := loadE2EFixture(t, "functions_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Robust_SchemasOrdering(t *testing.T) {
	input := loadE2EFixture(t, "schemas_input.sql")
	result := processDefault(input)

	assert.Contains(t, result, "CREATE SCHEMA app")
	assert.Contains(t, result, "CREATE TABLE app.users")
	assert.Contains(t, result, "CREATE TABLE public.countries")
	assert.Contains(t, result, "country_id bigint REFERENCES public.countries")
}

func TestFixture_Robust_ComplexSchemaFeatures(t *testing.T) {
	input := loadE2EFixture(t, "complex_schema_input.sql")
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

func TestFixture_Robust_AlterColumnChanges(t *testing.T) {
	input := loadE2EFixture(t, "alter_column_input.sql")
	result := processDefault(input)

	assert.NotContains(t, result, "DROP NOT NULL")
	assert.NotContains(t, result, "DROP DEFAULT")
	assert.NotContains(t, result, "SET DEFAULT")
	assert.NotContains(t, result, "SET NOT NULL")
	assert.NotContains(t, result, "ALTER COLUMN")
}

func TestFixture_Robust_SequenceHandling(t *testing.T) {
	input := loadFixtureFile(t, "sequences_input.sql")
	result := processDefault(input)

	assert.Contains(t, result, "BIGSERIAL")
	assertStatementContains(t, result, "CREATE SEQUENCE global_id_seq", "Shared sequence preserved")
	// Note: tags_id_seq is NOT preserved because smallint can be converted to SMALLSERIAL

	assertTypeBeforeTable(t, result, "global_id_seq", "orders")
}

func TestFixture_E2E_NoiseRemoval(t *testing.T) {
	input := loadE2EFixture(t, "noise_removal_input.sql")
	expected := loadE2EFixture(t, "noise_removal_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_E2E_SequenceToSerial(t *testing.T) {
	input := loadE2EFixture(t, "sequence_to_serial_input.sql")
	expected := loadE2EFixture(t, "sequence_to_serial_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_E2E_AlterFolding(t *testing.T) {
	input := loadE2EFixture(t, "alter_folding_input.sql")
	expected := loadE2EFixture(t, "alter_folding_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_E2E_FkInlining(t *testing.T) {
	input := loadE2EFixture(t, "fk_inlining_input.sql")
	expected := loadE2EFixture(t, "fk_inlining_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_E2E_AlterSequenceOptions(t *testing.T) {
	input := loadE2EFixture(t, "alter_sequence_options_input.sql")
	expected := loadE2EFixture(t, "alter_sequence_options_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_E2E_FkRemainingActions(t *testing.T) {
	input := loadE2EFixture(t, "fk_remaining_actions_input.sql")
	expected := loadE2EFixture(t, "fk_remaining_actions_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_E2E_QuotedIdentifiers(t *testing.T) {
	input := loadE2EFixture(t, "quoted_identifiers_input.sql")
	expected := loadE2EFixture(t, "quoted_identifiers_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_E2E_XmlType(t *testing.T) {
	input := loadE2EFixture(t, "xml_type_input.sql")
	expected := loadE2EFixture(t, "xml_type_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}
