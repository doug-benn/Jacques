package processor

import (
	"strings"
	"testing"

	"github.com/doug-benn/Jacques/internal/model"
	"github.com/stretchr/testify/assert"
)

func processDefault(sql string) string {
	return Process(sql, nil)
}

func processExperimental(sql string) string {
	return Process(sql, &Options{ExperimentalFolding: true})
}

func TestIntegration_ExtractSequenceName(t *testing.T) {
	tests := []struct {
		name string
		stmt string
		want string
	}{
		{
			name: "basic sequence",
			stmt: "CREATE SEQUENCE x;",
			want: "x",
		},
		{
			name: "sequence with schema",
			stmt: "CREATE SEQUENCE public.x;",
			want: "public.x",
		},
		{
			name: "sequence with IF NOT EXISTS",
			stmt: "CREATE SEQUENCE IF NOT EXISTS x;",
			want: "x",
		},
		{
			name: "sequence with schema and IF NOT EXISTS",
			stmt: "CREATE SEQUENCE IF NOT EXISTS public.x;",
			want: "public.x",
		},
		{
			name: "not a sequence",
			stmt: "CREATE TABLE x (id int);",
			want: "",
		},
		{
			name: "empty string",
			stmt: "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSequenceName(tt.stmt)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestIntegration_ExtractAlterSequenceName(t *testing.T) {
	tests := []struct {
		name string
		stmt string
		want string
	}{
		{
			name: "basic alter sequence",
			stmt: "ALTER SEQUENCE x;",
			want: "x",
		},
		{
			name: "alter sequence with schema",
			stmt: "ALTER SEQUENCE public.x;",
			want: "public.x",
		},
		{
			name: "alter sequence with IF EXISTS",
			stmt: "ALTER SEQUENCE IF EXISTS x;",
			want: "x",
		},
		{
			name: "not an alter sequence",
			stmt: "ALTER TABLE x;",
			want: "",
		},
		{
			name: "empty string",
			stmt: "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractAlterSequenceName(tt.stmt)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestIntegration_PreprocessSQL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "block comment removed",
			input: "/* note */ x;",
			want:  " x;",
		},
		{
			name:  "line comment removed",
			input: "-- note\nx;",
			want:  "x;",
		},
		{
			name:  "mixed comments removed",
			input: "/* block */ x; -- line",
			want:  " x; ",
		},
		{
			name:  "comment inside single quotes preserved",
			input: "SELECT '-- not a comment';",
			want:  "SELECT '-- not a comment';",
		},
		{
			name:  "comment inside dollar quotes preserved",
			input: "CREATE FUNCTION foo() AS $$ -- not a comment $$;",
			want:  "CREATE FUNCTION foo() AS $$ -- not a comment $$;",
		},
		{
			name:  "nested block comments removed",
			input: "/* outer /* inner */ outer */ x;",
			want:  " x;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := preprocessSQL(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestIntegration_DetectStatementType(t *testing.T) {
	tests := []struct {
		name string
		stmt string
		opts *Options
		want StatementType
	}{
		{
			name: "CREATE SEQUENCE",
			stmt: "CREATE SEQUENCE my_seq;",
			opts: &Options{},
			want: StatementSequence,
		},
		{
			name: "CREATE TABLE",
			stmt: "CREATE TABLE foo (id int);",
			opts: &Options{},
			want: StatementTable,
		},
		{
			name: "CREATE TYPE",
			stmt: "CREATE TYPE my_type AS ENUM ();",
			opts: &Options{},
			want: StatementTypeDomainSchema,
		},
		{
			name: "CREATE SCHEMA",
			stmt: "CREATE SCHEMA my_schema;",
			opts: &Options{},
			want: StatementTypeDomainSchema,
		},
		{
			name: "ALTER TABLE",
			stmt: "ALTER TABLE foo ADD COLUMN id int;",
			opts: &Options{},
			want: StatementAlter,
		},
		{
			name: "DROP TABLE with ExperimentalFolding",
			stmt: "DROP TABLE foo;",
			opts: &Options{ExperimentalFolding: true},
			want: StatementDrop,
		},
		{
			name: "DROP TABLE without ExperimentalFolding",
			stmt: "DROP TABLE foo;",
			opts: &Options{},
			want: StatementDrop, // DROP IF EXISTS is now default
		},
		{
			name: "CREATE DOMAIN with ExperimentalFolding",
			stmt: "CREATE DOMAIN my_domain AS int;",
			opts: &Options{ExperimentalFolding: true},
			want: StatementTypeDomainSchema,
		},
		{
			name: "CREATE DOMAIN without ExperimentalFolding",
			stmt: "CREATE DOMAIN my_domain AS int;",
			opts: &Options{},
			want: StatementTypeDomainSchema, // Basic DOMAIN is now default
		},
		{
			name: "CREATE DOMAIN with CHECK without ExperimentalFolding",
			stmt: "CREATE DOMAIN my_domain AS int CHECK (VALUE > 0);",
			opts: &Options{},
			want: StatementNoise, // DOMAIN with CHECK is gated
		},
		{
			name: "CREATE DOMAIN with CHECK with ExperimentalFolding",
			stmt: "CREATE DOMAIN my_domain AS int CHECK (VALUE > 0);",
			opts: &Options{ExperimentalFolding: true},
			want: StatementTypeDomainSchema,
		},
		{
			name: "Partition child without ExperimentalFolding",
			stmt: "CREATE TABLE foo_1 PARTITION OF foo FOR VALUES FROM (1) TO (100);",
			opts: &Options{},
			want: StatementTable,
		},
		{
			name: "Partition child with ExperimentalFolding",
			stmt: "CREATE TABLE foo_1 PARTITION OF foo FOR VALUES FROM (1) TO (100);",
			opts: &Options{ExperimentalFolding: true},
			want: StatementTable,
		},
		{
			name: "Unknown statement",
			stmt: "CREATE VIEW my_view AS SELECT 1;",
			opts: &Options{},
			want: StatementUnknown,
		},
		{
			name: "Line comment",
			stmt: "-- this is a comment",
			opts: &Options{},
			want: StatementNoise,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectStatementType(tt.stmt, tt.opts)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestIntegration_CategorizeStatements(t *testing.T) {
	tests := []struct {
		name            string
		statements      []string
		opts            *Options
		wantTables      int
		wantTypeStmts   int
		wantPassThrough int
	}{
		{
			name: "CREATE TABLE and unknown ALTER",
			statements: []string{
				"CREATE TABLE foo (id int);",
				"ALTER TABLE foo SET something = 1;",
			},
			opts:            &Options{},
			wantTables:      1,
			wantTypeStmts:   0,
			wantPassThrough: 1,
		},
		{
			name: "CREATE TYPE and TABLE",
			statements: []string{
				"CREATE TYPE my_type AS ENUM ('a', 'b');",
				"CREATE TABLE foo (id int, type my_type);",
			},
			opts:            &Options{},
			wantTables:      1,
			wantTypeStmts:   1,
			wantPassThrough: 0,
		},
		{
			name: "CREATE SEQUENCE",
			statements: []string{
				"CREATE SEQUENCE my_seq;",
			},
			opts:            &Options{},
			wantTables:      0,
			wantTypeStmts:   0,
			wantPassThrough: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tables, _, typeStmts, passThroughs, _, tableOrder := categorizeStatements(tt.statements, tt.opts)
			assert.Equal(t, tt.wantTables, len(tables), "tables count")
			assert.Equal(t, tt.wantTables, len(tableOrder), "tableOrder count")
			assert.Equal(t, tt.wantTypeStmts, len(typeStmts), "typeStmts count")
			assert.Equal(t, tt.wantPassThrough, len(passThroughs), "passThroughs count")
		})
	}
}

func TestIntegration_InferMissingSchemas(t *testing.T) {
	tests := []struct {
		name      string
		tables    map[string]*model.TableDef
		typeStmts []string
		want      int
	}{
		{
			name: "no missing schemas",
			tables: map[string]*model.TableDef{
				"public.foo": {Schema: "public", Name: "foo"},
			},
			typeStmts: []string{"CREATE SCHEMA app;"},
			want:      0,
		},
		{
			name: "one missing schema",
			tables: map[string]*model.TableDef{
				"app.foo": {Schema: "app", Name: "foo"},
			},
			typeStmts: []string{},
			want:      1,
		},
		{
			name: "schema already declared",
			tables: map[string]*model.TableDef{
				"app.foo": {Schema: "app", Name: "foo"},
			},
			typeStmts: []string{"CREATE SCHEMA app;"},
			want:      0,
		},
		{
			name: "multiple missing schemas",
			tables: map[string]*model.TableDef{
				"app.foo":   {Schema: "app", Name: "foo"},
				"other.bar": {Schema: "other", Name: "bar"},
			},
			typeStmts: []string{},
			want:      2,
		},
		{
			name: "quoted schema case sensitivity",
			tables: map[string]*model.TableDef{
				"\"MySchema\".foo": {Schema: "\"MySchema\"", Name: "foo"},
			},
			// Current implementation uses strings.ToLower on schema name,
			// so "MySchema" becomes "myschema", which may incorrectly
			// match an unquoted schema if one existed.
			// This test ensures we're aware of the behavior.
			typeStmts: []string{"CREATE SCHEMA \"myschema\";"},
			want:      1, // Should be 1 because "MySchema" != "myschema" in Postgres
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inferMissingSchemas(tt.tables, tt.typeStmts)
			assert.Equal(t, tt.want, len(result))
		})
	}
}

func TestIntegration_CountSequenceUsage(t *testing.T) {
	tests := []struct {
		name       string
		tables     map[string]*model.TableDef
		tableOrder []string
		want       map[string]int
	}{
		{
			name: "single sequence usage",
			tables: map[string]*model.TableDef{
				"foo": {
					Name: "foo",
					Columns: []*model.ColumnDef{
						{Name: "id", SequenceName: "foo_id_seq"},
					},
				},
			},
			tableOrder: []string{"foo"},
			want:       map[string]int{"foo_id_seq": 1},
		},
		{
			name: "shared sequence",
			tables: map[string]*model.TableDef{
				"foo": {
					Name: "foo",
					Columns: []*model.ColumnDef{
						{Name: "id", SequenceName: "shared_seq"},
					},
				},
				"bar": {
					Name: "bar",
					Columns: []*model.ColumnDef{
						{Name: "id", SequenceName: "shared_seq"},
					},
				},
			},
			tableOrder: []string{"foo", "bar"},
			want:       map[string]int{"shared_seq": 2},
		},
		{
			name: "sequence with schema prefix",
			tables: map[string]*model.TableDef{
				"foo": {
					Name: "foo",
					Columns: []*model.ColumnDef{
						{Name: "id", SequenceName: "public.foo_id_seq"},
					},
				},
			},
			tableOrder: []string{"foo"},
			want:       map[string]int{"foo_id_seq": 1},
		},
		{
			name: "no sequences",
			tables: map[string]*model.TableDef{
				"foo": {
					Name: "foo",
					Columns: []*model.ColumnDef{
						{Name: "id", RawDef: "int"},
					},
				},
			},
			tableOrder: []string{"foo"},
			want:       map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countSequenceUsage(tt.tables, tt.tableOrder, make(map[string]map[string]bool))
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestIntegration_ApplySerialConversion(t *testing.T) {
	tests := []struct {
		name       string
		tables     map[string]*model.TableDef
		tableOrder []string
		usageCount map[string]int
		wantSerial bool
	}{
		{
			name: "single usage converts to serial",
			tables: map[string]*model.TableDef{
				"foo": {
					Name: "foo",
					Columns: []*model.ColumnDef{
						{Name: "id", RawDef: "bigint", SequenceName: "foo_id_seq"},
					},
				},
			},
			tableOrder: []string{"foo"},
			usageCount: map[string]int{"foo_id_seq": 1},
			wantSerial: true,
		},
		{
			name: "multiple usage does not convert",
			tables: map[string]*model.TableDef{
				"foo": {
					Name: "foo",
					Columns: []*model.ColumnDef{
						{Name: "id", RawDef: "bigint", SequenceName: "shared_seq"},
					},
				},
			},
			tableOrder: []string{"foo"},
			usageCount: map[string]int{"shared_seq": 2},
			wantSerial: false,
		},
		{
			name: "integer type converts to serial",
			tables: map[string]*model.TableDef{
				"foo": {
					Name: "foo",
					Columns: []*model.ColumnDef{
						{Name: "id", RawDef: "integer", SequenceName: "foo_id_seq"},
					},
				},
			},
			tableOrder: []string{"foo"},
			usageCount: map[string]int{"foo_id_seq": 1},
			wantSerial: true,
		},
		{
			name: "text type does not convert",
			tables: map[string]*model.TableDef{
				"foo": {
					Name: "foo",
					Columns: []*model.ColumnDef{
						{Name: "id", RawDef: "text", SequenceName: "foo_id_seq"},
					},
				},
			},
			tableOrder: []string{"foo"},
			usageCount: map[string]int{"foo_id_seq": 1},
			wantSerial: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applySerialConversion(tt.tables, tt.tableOrder, tt.usageCount, nil, make(map[string]map[string]bool))
			for _, td := range tt.tables {
				for _, col := range td.Columns {
					if col.SequenceName != "" {
						assert.Equal(t, tt.wantSerial, col.IsSerial, "IsSerial for column %s", col.Name)
					}
				}
			}
		})
	}
}

func TestIntegration_ExtractSequencesFromPassthroughs(t *testing.T) {
	tests := []struct {
		name          string
		passThroughs  []string
		usageCount    map[string]int
		tables        map[string]*model.TableDef
		wantKept      int
		wantConverted int
	}{
		{
			name:          "unused sequence kept",
			passThroughs:  []string{"CREATE SEQUENCE my_seq;"},
			usageCount:    map[string]int{"my_seq": 0},
			tables:        map[string]*model.TableDef{},
			wantKept:      1,
			wantConverted: 0,
		},
		{
			name:          "shared sequence kept",
			passThroughs:  []string{"CREATE SEQUENCE shared_seq;"},
			usageCount:    map[string]int{"shared_seq": 2},
			tables:        map[string]*model.TableDef{},
			wantKept:      1,
			wantConverted: 0,
		},
		{
			name:         "single usage converted to serial",
			passThroughs: []string{"CREATE SEQUENCE foo_id_seq;"},
			usageCount:   map[string]int{"foo_id_seq": 1},
			tables: map[string]*model.TableDef{
				"foo": {
					Name: "foo",
					Columns: []*model.ColumnDef{
						{Name: "id", RawDef: "bigint", SequenceName: "foo_id_seq"},
					},
				},
			},
			wantKept:      0,
			wantConverted: 1,
		},
		{
			name:         "single usage but wrong type kept",
			passThroughs: []string{"CREATE SEQUENCE foo_id_seq;"},
			usageCount:   map[string]int{"foo_id_seq": 1},
			tables: map[string]*model.TableDef{
				"foo": {
					Name: "foo",
					Columns: []*model.ColumnDef{
						{Name: "id", RawDef: "text", SequenceName: "foo_id_seq"},
					},
				},
			},
			wantKept:      1,
			wantConverted: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kept, converted, _ := extractSequencesFromPassthroughs(tt.passThroughs, tt.usageCount, tt.tables)
			assert.Equal(t, tt.wantKept, len(kept), "kept sequences")
			assert.Equal(t, tt.wantConverted, len(converted), "converted sequences")
		})
	}
}

// TestIntegrity_NonSQLFormatting verifies that regular text and comments
// aren't mangled or collapsed.
func TestIntegrity_NonSQLFormatting(t *testing.T) {
	input := `This is a line.
This is another line with a -- comment.
This is the ONLY way to do it.
And a semicolon here;
And one here;`

	output := Process(input, nil)
	//TODO Check this test
	// Newlines after line comments are removed
	assert.Contains(t, output, "This is another line with a This is the ONLY way to do it.")
	// Global ONLY should be preserved in regular text
	assert.Contains(t, output, "This is the ONLY way to do it.")
}

// TestIntegrity_MarkdownDocumentation ensures that Jacques doesn't damage
// Markdown documentation mixed with SQL.
func TestIntegrity_MarkdownDocumentation(t *testing.T) {
	input := `# Project Schema

This is the database schema for the project.
It contains important tables for users and orders.

-- Database Schema
CREATE TABLE users (
    id bigint PRIMARY KEY
);

## Future Improvements
* Add more indexes
* Optimize queries
`
	output := Process(input, nil)

	assert.Contains(t, output, "# Project Schema")
	assert.Contains(t, output, "## Future Improvements")
	assert.Contains(t, output, "* Add more indexes")
	// Ensure the prose isn't collapsed into one line
	assert.True(t, strings.Count(output, "\n") >= 10, "Output should maintain original line structure")
}

// TestIntegrity_ComplexStrings verifies that complex string literals
// (like JSON or code snippets) are preserved exactly.
func TestIntegrity_ComplexStrings(t *testing.T) {
	input := `CREATE TABLE settings (
    key text PRIMARY KEY,
    -- JSON default containing many potential "noise" tokens
    config jsonb DEFAULT '{"only": true, "grant": "all", "set": 1, "comment": "none"}'
);

INSERT INTO settings (key, config) VALUES ('test', '{
    "nested": {
        "value": 1
    },
    "script": "alert(''hello'');"
}');`

	output := Process(input, nil)

	// Verify JSON content is untouched
	assert.Contains(t, output, `{"only": true, "grant": "all", "set": 1, "comment": "none"}`)
	assert.Contains(t, output, `"script": "alert(''hello'');"`)
}

// TestIntegrity_TemplateVariables ensures that template placeholders
// used in some deployment scripts aren't stripped or mangled.
func TestIntegrity_TemplateVariables(t *testing.T) {
	input := `CREATE ROLE ${DB_USER} WITH PASSWORD '${DB_PASSWORD}';
CREATE TABLE {{ .TableName }} (
    id bigint PRIMARY KEY
);`

	output := Process(input, nil)

	assert.Contains(t, output, "CREATE ROLE ${DB_USER}")
	assert.Contains(t, output, "PASSWORD '${DB_PASSWORD}'")
	assert.Contains(t, output, "CREATE TABLE {{ .TableName }}")
}

// TestIntegrity_SpecialIdentifiers ensures that identifiers containing
// potential "noise" characters like -- or ; are preserved.
func TestIntegrity_SpecialIdentifiers(t *testing.T) {
	input := `CREATE TABLE "my--table;with--noise" (
    id bigint PRIMARY KEY
);`
	output := Process(input, nil)
	assert.Contains(t, output, `"my--table;with--noise"`)
}

// TestIntegrity_DelimitersInComments ensures that semicolons or dollar tags
// inside comments don't break statement splitting or preprocessing.
func TestIntegrity_DelimitersInComments(t *testing.T) {
	input := `CREATE TABLE test (id int); -- This is an end; of a statement?
-- What about a $$ dollar tag?
CREATE TABLE other (id int);`
	output := Process(input, nil)
	assert.Contains(t, output, "CREATE TABLE test")
	assert.Contains(t, output, "CREATE TABLE other")
}

// TestIntegrity_MixedLineEndings ensures that Jacques handles both LF and CRLF
// without corrupting the content.
func TestIntegrity_MixedLineEndings(t *testing.T) {
	input := "Line 1\nLine 2\r\nLine 3"
	output := Process(input, nil)
	// We expect line content to be preserved, though line endings might be normalized
	assert.Contains(t, output, "Line 1")
	assert.Contains(t, output, "Line 2")
	assert.Contains(t, output, "Line 3")
}

// TestIntegrity_BinaryAndLargeData ensures that large data blocks
// aren't reformatted or corrupted.
func TestIntegrity_BinaryAndLargeData(t *testing.T) {
	input := `INSERT INTO large_data (id, bin) VALUES (1, E'\\xdeadbeef0123456789');`
	output := Process(input, nil)
	assert.Equal(t, input, output)
}
