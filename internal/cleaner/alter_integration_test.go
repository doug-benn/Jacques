package cleaner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/doug-benn/Jacques/internal/model"
)

func TestIntegration_RouteAlter(t *testing.T) {
	tests := []struct {
		name    string
		stmt    string
		tables  map[string]*model.TableDef
		wantNil bool
		check   func(*testing.T, *string, map[string]*model.TableDef)
	}{
		{
			name:    "OWNER TO discarded",
			stmt:    "ALTER TABLE foo OWNER TO bar;",
			tables:  make(map[string]*model.TableDef),
			wantNil: true,
		},
		{
			name:    "CLUSTER ON discarded",
			stmt:    "ALTER TABLE foo CLUSTER ON idx;",
			tables:  make(map[string]*model.TableDef),
			wantNil: true,
		},
		{
			name:    "SET WITHOUT CLUSTER discarded",
			stmt:    "ALTER TABLE foo SET WITHOUT CLUSTER;",
			tables:  make(map[string]*model.TableDef),
			wantNil: true,
		},
		{
			name:    "SET WITHOUT OIDS discarded",
			stmt:    "ALTER TABLE foo SET WITHOUT OIDS;",
			tables:  make(map[string]*model.TableDef),
			wantNil: true,
		},
		{
			name:    "ALTER SEQUENCE OWNED BY discarded",
			stmt:    "ALTER SEQUENCE foo_seq OWNED BY foo.id;",
			tables:  make(map[string]*model.TableDef),
			wantNil: true,
		},
		{
			name: "SET DEFAULT folded",
			stmt: "ALTER TABLE foo ALTER COLUMN id SET DEFAULT 1;",
			tables: map[string]*model.TableDef{
				"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}}),
			},
			wantNil: true,
			check: func(t *testing.T, _ *string, tables map[string]*model.TableDef) {
				assert.Equal(t, "1", tables["foo"].Columns[0].Default)
			},
		},
		{
			name: "SET NOT NULL folded",
			stmt: "ALTER TABLE foo ALTER COLUMN id SET NOT NULL;",
			tables: map[string]*model.TableDef{
				"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}}),
			},
			wantNil: true,
			check: func(t *testing.T, _ *string, tables map[string]*model.TableDef) {
				assert.Contains(t, tables["foo"].Columns[0].RawDef, "NOT NULL")
			},
		},
		{
			name: "SET TYPE folded",
			stmt: "ALTER TABLE foo ALTER COLUMN id TYPE int;",
			tables: map[string]*model.TableDef{
				"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}}),
			},
			wantNil: true,
			check: func(t *testing.T, _ *string, tables map[string]*model.TableDef) {
				assert.Contains(t, tables["foo"].Columns[0].RawDef, "int")
			},
		},
		{
			name: "DROP DEFAULT discarded",
			stmt: "ALTER TABLE foo ALTER COLUMN id DROP DEFAULT;",
			tables: map[string]*model.TableDef{
				"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint DEFAULT 1"}}),
			},
			wantNil: true,
		},
		{
			name: "DROP NOT NULL folded",
			stmt: "ALTER TABLE foo ALTER COLUMN id DROP NOT NULL;",
			tables: map[string]*model.TableDef{
				"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint NOT NULL"}}),
			},
			wantNil: true,
			check: func(t *testing.T, _ *string, tables map[string]*model.TableDef) {
				assert.NotContains(t, tables["foo"].Columns[0].RawDef, "NOT NULL")
			},
		},
		{
			name: "ADD PRIMARY KEY single folded",
			stmt: "ALTER TABLE foo ADD CONSTRAINT pk PRIMARY KEY (id);",
			tables: map[string]*model.TableDef{
				"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}}),
			},
			wantNil: true,
			check: func(t *testing.T, _ *string, tables map[string]*model.TableDef) {
				assert.True(t, tables["foo"].Columns[0].IsPrimaryKey)
			},
		},
		{
			name: "ADD PRIMARY KEY multi folded",
			stmt: "ALTER TABLE foo ADD CONSTRAINT pk PRIMARY KEY (id, other);",
			tables: map[string]*model.TableDef{
				"foo": newTable("foo", []*model.ColumnDef{
					{Name: "id", RawDef: "bigint"},
					{Name: "other", RawDef: "int"},
				}),
			},
			wantNil: true,
			check: func(t *testing.T, _ *string, tables map[string]*model.TableDef) {
				assert.Equal(t, "id, other", tables["foo"].TableLevelPK)
			},
		},
		{
			name: "ADD UNIQUE single folded",
			stmt: "ALTER TABLE foo ADD CONSTRAINT uq UNIQUE (email);",
			tables: map[string]*model.TableDef{
				"foo": newTable("foo", []*model.ColumnDef{{Name: "email", RawDef: "text"}}),
			},
			wantNil: true,
			check: func(t *testing.T, _ *string, tables map[string]*model.TableDef) {
				assert.True(t, tables["foo"].Columns[0].IsUnique)
			},
		},
		{
			name: "ADD UNIQUE multi folded",
			stmt: "ALTER TABLE foo ADD CONSTRAINT uq UNIQUE (a, b);",
			tables: map[string]*model.TableDef{
				"foo": newTable("foo", []*model.ColumnDef{
					{Name: "a", RawDef: "int"},
					{Name: "b", RawDef: "int"},
				}),
			},
			wantNil: true,
			check: func(t *testing.T, _ *string, tables map[string]*model.TableDef) {
				assert.Contains(t, tables["foo"].TableLevelUniques, "a, b")
			},
		},
		{
			name: "ADD CHECK folded",
			stmt: "ALTER TABLE foo ADD CONSTRAINT ck CHECK (amount > 0);",
			tables: map[string]*model.TableDef{
				"foo": newTableWithConstraints("foo", []*model.ColumnDef{{Name: "amount", RawDef: "int"}}, nil),
			},
			wantNil: true,
			check: func(t *testing.T, _ *string, tables map[string]*model.TableDef) {
				require.Len(t, tables["foo"].TableConstraints, 1)
				assert.Contains(t, tables["foo"].TableConstraints[0], "CHECK")
			},
		},
		{
			name: "ADD FOREIGN KEY inlined",
			stmt: "ALTER TABLE bar ADD CONSTRAINT fk FOREIGN KEY (foo_id) REFERENCES foo(id);",
			tables: map[string]*model.TableDef{
				"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}}),
				"bar": newTable("bar", []*model.ColumnDef{{Name: "foo_id", RawDef: "bigint"}}),
			},
			wantNil: true,
			check: func(t *testing.T, _ *string, tables map[string]*model.TableDef) {
				assert.Equal(t, "foo(id)", tables["bar"].Columns[0].References)
			},
		},
		{
			name: "ADD FOREIGN KEY with ON DELETE CASCADE",
			stmt: "ALTER TABLE bar ADD CONSTRAINT fk FOREIGN KEY (foo_id) REFERENCES foo(id) ON DELETE CASCADE;",
			tables: map[string]*model.TableDef{
				"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}}),
				"bar": newTable("bar", []*model.ColumnDef{{Name: "foo_id", RawDef: "bigint"}}),
			},
			wantNil: true,
			check: func(t *testing.T, _ *string, tables map[string]*model.TableDef) {
				assert.Equal(t, "foo(id)", tables["bar"].Columns[0].References)
				assert.Equal(t, "CASCADE", tables["bar"].Columns[0].OnDelete)
			},
		},
		{
			name: "ADD FOREIGN KEY with ON DELETE SET NULL",
			stmt: "ALTER TABLE bar ADD CONSTRAINT fk FOREIGN KEY (foo_id) REFERENCES foo(id) ON DELETE SET NULL;",
			tables: map[string]*model.TableDef{
				"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}}),
				"bar": newTable("bar", []*model.ColumnDef{{Name: "foo_id", RawDef: "bigint"}}),
			},
			wantNil: true,
			check: func(t *testing.T, _ *string, tables map[string]*model.TableDef) {
				assert.Equal(t, "foo(id)", tables["bar"].Columns[0].References)
				assert.Equal(t, "SET NULL", tables["bar"].Columns[0].OnDelete)
			},
		},
		{
			name: "ADD FOREIGN KEY with ON DELETE CASCADE ON UPDATE RESTRICT",
			stmt: "ALTER TABLE bar ADD CONSTRAINT fk FOREIGN KEY (foo_id) REFERENCES foo(id) ON DELETE CASCADE ON UPDATE RESTRICT;",
			tables: map[string]*model.TableDef{
				"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}}),
				"bar": newTable("bar", []*model.ColumnDef{{Name: "foo_id", RawDef: "bigint"}}),
			},
			wantNil: true,
			check: func(t *testing.T, _ *string, tables map[string]*model.TableDef) {
				assert.Equal(t, "foo(id)", tables["bar"].Columns[0].References)
				assert.Equal(t, "CASCADE", tables["bar"].Columns[0].OnDelete)
				assert.Equal(t, "RESTRICT", tables["bar"].Columns[0].OnUpdate)
			},
		},
		{
			name: "ADD FOREIGN KEY with MATCH FULL",
			stmt: "ALTER TABLE bar ADD CONSTRAINT fk FOREIGN KEY (foo_id) REFERENCES foo(id) MATCH FULL;",
			tables: map[string]*model.TableDef{
				"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}}),
				"bar": newTable("bar", []*model.ColumnDef{{Name: "foo_id", RawDef: "bigint"}}),
			},
			wantNil: true,
			check: func(t *testing.T, _ *string, tables map[string]*model.TableDef) {
				assert.Equal(t, "foo(id)", tables["bar"].Columns[0].References)
				assert.Equal(t, "FULL", tables["bar"].Columns[0].Match)
			},
		},
		{
			name: "ADD COLUMN inlined",
			stmt: "ALTER TABLE foo ADD COLUMN new_col int;",
			tables: map[string]*model.TableDef{
				"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}}),
			},
			wantNil: true,
			check: func(t *testing.T, result *string, tables map[string]*model.TableDef) {
				assert.Nil(t, result)
				assert.Len(t, tables["foo"].Columns, 2)
				assert.Equal(t, "new_col", tables["foo"].Columns[1].Name)
			},
		},
		{
			name:    "unknown ALTER passthrough",
			stmt:    "ALTER TABLE foo SET something = 1;",
			tables:  make(map[string]*model.TableDef),
			wantNil: false,
			check: func(t *testing.T, result *string, _ map[string]*model.TableDef) {
				assert.Contains(t, *result, "ALTER TABLE foo SET something = 1;")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RouteAlter(tt.stmt, tt.tables)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
			}
			if tt.check != nil {
				tt.check(t, result, tt.tables)
			}
		})
	}
}

func TestIntegration_ConstraintPriority(t *testing.T) {
	tables := map[string]*model.TableDef{
		"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}}),
	}
	RouteAlter("ALTER TABLE foo ADD CONSTRAINT pk PRIMARY KEY (id);", tables)
	RouteAlter("ALTER TABLE foo ADD CONSTRAINT uq UNIQUE (id);", tables)

	assert.True(t, tables["foo"].Columns[0].IsPrimaryKey)
	assert.False(t, tables["foo"].Columns[0].IsUnique)
}

func TestIntegration_FindTable(t *testing.T) {
	tests := []struct {
		name      string
		tables    map[string]*model.TableDef
		schema    string
		tableName string
		wantName  string
		wantNil   bool
	}{
		{
			name: "qualified match",
			tables: map[string]*model.TableDef{
				"public.foo": {Schema: "public", Name: "foo"},
			},
			schema:    "public",
			tableName: "foo",
			wantName:  "foo",
		},
		{
			name: "unqualified match",
			tables: map[string]*model.TableDef{
				"public.foo": {Schema: "public", Name: "foo"},
				"foo":        {Schema: "public", Name: "foo"},
			},
			schema:    "",
			tableName: "foo",
			wantName:  "foo",
		},
		{
			name: "case insensitive",
			tables: map[string]*model.TableDef{
				"public.foo": {Schema: "public", Name: "foo"},
			},
			schema:    "PUBLIC",
			tableName: "FOO",
			wantName:  "foo",
		},
		{
			name: "not found",
			tables: map[string]*model.TableDef{
				"public.foo": {Schema: "public", Name: "foo"},
			},
			schema:    "",
			tableName: "bar",
			wantNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindTable(tt.tables, tt.schema, tt.tableName)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.wantName, result.Name)
			}
		})
	}
}
