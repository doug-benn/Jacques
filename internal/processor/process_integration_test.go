package processor

import (
	"testing"

	"github.com/doug-benn/Jacques/internal/model"
	"github.com/stretchr/testify/assert"
)

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

func TestIntegration_RemoveBlockComments(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple block comment",
			input: "/* note */ x;",
			want:  " x;",
		},
		{
			name:  "multiline block comment",
			input: "/* note\nline2 */ x;",
			want:  " x;",
		},
		{
			name:  "no block comment",
			input: "x;",
			want:  "x;",
		},
		{
			name:  "multiple block comments",
			input: "/* a */ x; /* b */",
			want:  " x; ",
		},
		{
			name:  "block comment in middle",
			input: "x /* note */ y;",
			want:  "x  y;",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeBlockComments(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestIntegration_RemoveLineComments(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "line comment at start",
			input: "-- note\nx;",
			want:  "x;\n",
		},
		{
			name:  "line comment at end",
			input: "x; -- note",
			want:  "x; \n",
		},
		{
			name:  "line comment in middle",
			input: "x; -- note\ny;",
			want:  "x; \ny;\n",
		},
		{
			name:  "no line comment",
			input: "x;",
			want:  "x;\n",
		},
		{
			name:  "multiple line comments",
			input: "-- a\n-- b\nx;",
			want:  "x;\n",
		},
		{
			name:  "empty string",
			input: "",
			want:  "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeLineComments(tt.input)
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
			want:  " x;\n",
		},
		{
			name:  "line comment removed",
			input: "-- note\nx;",
			want:  "x;\n",
		},
		{
			name:  "mixed comments removed",
			input: "/* block */ x; -- line",
			want:  " x; \n",
		},
		{
			name:  "comment in middle of line",
			input: "x; -- note",
			want:  "x; \n",
		},
		{
			name:  "multiline block comment",
			input: "/*\nmulti\nline\n*/ x;",
			want:  " x;\n",
		},
		{
			name:  "empty string",
			input: "",
			want:  "\n",
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
			result := countSequenceUsage(tt.tables, tt.tableOrder)
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
			applySerialConversion(tt.tables, tt.tableOrder, tt.usageCount)
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
			kept, converted := extractSequencesFromPassthroughs(tt.passThroughs, tt.usageCount, tt.tables)
			assert.Equal(t, tt.wantKept, len(kept), "kept sequences")
			assert.Equal(t, tt.wantConverted, len(converted), "converted sequences")
		})
	}
}
