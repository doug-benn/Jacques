package processor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixture_DomainTypes(t *testing.T) {
	input := loadIntegrationFixture(t, "domain_types_input.sql")
	expected := loadIntegrationFixture(t, "domain_types_expected.sql")
	result := processExperimental(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Integration_Inheritance(t *testing.T) {
	input := loadIntegrationFixture(t, "inheritance_input.sql")
	expected := loadIntegrationFixture(t, "inheritance_expected.sql")
	result := processExperimental(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Integration_PartitionedTables(t *testing.T) {
	input := loadIntegrationFixture(t, "partitioned_tables_input.sql")
	expected := loadIntegrationFixture(t, "partitioned_tables_expected.sql")
	result := processExperimental(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Integration_IfExistsForDrop(t *testing.T) {
	input := loadIntegrationFixture(t, "drop_statements_input.sql")
	expected := loadIntegrationFixture(t, "drop_statements_expected.sql")
	result := processExperimental(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Integration_OnlyRemoval(t *testing.T) {
	input := loadTestFixture(t, "only_removal_input.sql")
	expected := loadTestFixture(t, "only_removal_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Gating_DomainTypesSkippedByDefault(t *testing.T) {
	input := loadIntegrationFixture(t, "domain_types_input.sql")
	result := processDefault(input)
	assert.NotContains(t, result, "CREATE DOMAIN")
	assert.Contains(t, result, "email public.email")
}

func TestFixture_Gating_DomainTypesPreservedWithExperimentalFolding(t *testing.T) {
	input := loadIntegrationFixture(t, "domain_types_input.sql")
	result := processExperimental(input)
	assert.Contains(t, result, "CREATE DOMAIN public.email")
	assert.Contains(t, result, "CREATE DOMAIN public.positive_int")
}

func TestFixture_Gating_PartitionChildrenSkippedByDefault(t *testing.T) {
	input := loadIntegrationFixture(t, "partitioned_tables_input.sql")
	result := processDefault(input)
	assert.NotContains(t, result, "PARTITION OF")
	assert.Contains(t, result, "PARTITION BY RANGE")
	assert.Contains(t, result, "PARTITION BY LIST")
	assert.Contains(t, result, "PARTITION BY HASH")
}

func TestFixture_Gating_PartitionChildrenPreservedWithExperimentalFolding(t *testing.T) {
	input := loadIntegrationFixture(t, "partitioned_tables_input.sql")
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
	input := loadIntegrationFixture(t, "inheritance_input.sql")
	result := processDefault(input)

	assert.Contains(t, result, "CREATE TABLE public.users")
	assert.Contains(t, result, "CREATE TABLE public.administrators")
	assert.Contains(t, result, "CREATE TABLE public.moderators")
	assert.Contains(t, result, "CREATE TABLE public.registered_users")

	assert.NotContains(t, result, "INHERITS")
}

func TestFixture_Robust_InheritancePreservedWithExperimentalFolding(t *testing.T) {
	input := loadIntegrationFixture(t, "inheritance_input.sql")
	result := processExperimental(input)

	assert.Contains(t, result, "CREATE TABLE public.users")
	assert.Contains(t, result, "CREATE TABLE public.administrators")
	assert.Contains(t, result, "INHERITS (public.users)")
}

func TestFixture_Robust_DomainTypesOrdering(t *testing.T) {
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

func TestFixture_Integration_BasicTable(t *testing.T) {
	input := loadTestFixture(t, "basic_table_input.sql")
	expected := loadTestFixture(t, "basic_table_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Integration_FKCheck(t *testing.T) {
	input := loadTestFixture(t, "fk_check_input.sql")
	expected := loadTestFixture(t, "fk_check_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
	assert.Contains(t, result, "REFERENCES public.orders(id)")
}

func TestFixture_Integration_Sequences(t *testing.T) {
	input := loadTestFixture(t, "sequences_input.sql")
	expected := loadTestFixture(t, "sequences_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Integration_FKCascade(t *testing.T) {
	input := loadTestFixture(t, "fk_cascade_input.sql")
	expected := loadTestFixture(t, "fk_cascade_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Integration_Types(t *testing.T) {
	input := loadTestFixture(t, "types_input.sql")
	expected := loadTestFixture(t, "types_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Integration_AlterColumn(t *testing.T) {
	input := loadTestFixture(t, "alter_column_input.sql")
	expected := loadTestFixture(t, "alter_column_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Integration_SelfReferential(t *testing.T) {
	input := loadTestFixture(t, "self_referential_input.sql")
	expected := loadTestFixture(t, "self_referential_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Integration_PartialIndexes(t *testing.T) {
	input := loadTestFixture(t, "partial_indexes_input.sql")
	expected := loadTestFixture(t, "partial_indexes_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Integration_MaterializedViews(t *testing.T) {
	input := loadTestFixture(t, "materialized_views_input.sql")
	expected := loadTestFixture(t, "materialized_views_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Integration_Collation(t *testing.T) {
	input := loadTestFixture(t, "collation_input.sql")
	expected := loadTestFixture(t, "collation_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Integration_GeneratedColumns(t *testing.T) {
	input := loadTestFixture(t, "generated_columns_input.sql")
	expected := loadTestFixture(t, "generated_columns_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Integration_Schemas(t *testing.T) {
	input := loadTestFixture(t, "schemas_input.sql")
	expected := loadTestFixture(t, "schemas_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Integration_IdentityColumns(t *testing.T) {
	input := loadTestFixture(t, "identity_columns_input.sql")
	expected := loadTestFixture(t, "identity_columns_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Integration_ExclusionConstraints(t *testing.T) {
	input := loadTestFixture(t, "exclusion_constraints_input.sql")
	expected := loadTestFixture(t, "exclusion_constraints_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Integration_FKColumnMapping(t *testing.T) {
	input := loadTestFixture(t, "fk_column_mapping_input.sql")
	expected := loadTestFixture(t, "fk_column_mapping_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Integration_RangeTypes(t *testing.T) {
	input := loadTestFixture(t, "range_types_input.sql")
	expected := loadTestFixture(t, "range_types_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Integration_FKNoAction(t *testing.T) {
	input := loadTestFixture(t, "fk_no_action_input.sql")
	expected := loadTestFixture(t, "fk_no_action_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Integration_Triggers(t *testing.T) {
	input := loadTestFixture(t, "triggers_input.sql")
	expected := loadTestFixture(t, "triggers_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Integration_RLSPolicies(t *testing.T) {
	input := loadTestFixture(t, "rls_policies_input.sql")
	expected := loadTestFixture(t, "rls_policies_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Integration_Functions(t *testing.T) {
	input := loadTestFixture(t, "functions_input.sql")
	expected := loadTestFixture(t, "functions_expected.sql")
	result := processDefault(input)
	assert.Equal(t, normalizeSQL(expected), normalizeSQL(result))
}

func TestFixture_Robust_SchemasOrdering(t *testing.T) {
	input := loadTestFixture(t, "schemas_input.sql")
	result := processDefault(input)

	assert.Contains(t, result, "CREATE SCHEMA app")
	assert.Contains(t, result, "CREATE TABLE app.users")
	assert.Contains(t, result, "CREATE TABLE public.countries")
	assert.Contains(t, result, "country_id bigint REFERENCES public.countries")
}

func TestFixture_Robust_ComplexSchemaFeatures(t *testing.T) {
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

func TestFixture_Robust_AlterColumnChanges(t *testing.T) {
	input := loadTestFixture(t, "alter_column_input.sql")
	result := processDefault(input)

	assert.NotContains(t, result, "DROP NOT NULL")
	assert.NotContains(t, result, "DROP DEFAULT")
	assert.NotContains(t, result, "SET DEFAULT")
	assert.NotContains(t, result, "SET NOT NULL")
	assert.NotContains(t, result, "ALTER COLUMN")
}

func TestFixture_Robust_SequenceHandling(t *testing.T) {
	input := loadTestFixture(t, "sequences_input.sql")
	result := processDefault(input)

	assert.Contains(t, result, "BIGSERIAL")
	assertStatementContains(t, result, "CREATE SEQUENCE global_id_seq", "Shared sequence preserved")
	// Note: tags_id_seq is NOT preserved because smallint can be converted to SMALLSERIAL

	assertTypeBeforeTable(t, result, "global_id_seq", "orders")
}
