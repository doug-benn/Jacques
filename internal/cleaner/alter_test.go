package cleaner

import (
	"testing"

	"github.com/doug-benn/Jacques/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestIsAlterDiscardPattern(t *testing.T) {
	tests := []struct {
		name   string
		stmt   string
		expect bool
	}{
		{
			name:   "OWNER TO",
			stmt:   "ALTER TABLE foo OWNER TO bar;",
			expect: true,
		},
		{
			name:   "CLUSTER ON",
			stmt:   "ALTER TABLE foo CLUSTER ON idx;",
			expect: true,
		},
		{
			name:   "SET WITHOUT CLUSTER",
			stmt:   "ALTER TABLE foo SET WITHOUT CLUSTER;",
			expect: true,
		},
		{
			name:   "SET WITHOUT OIDS",
			stmt:   "ALTER TABLE foo SET WITHOUT OIDS;",
			expect: true,
		},
		{
			name:   "not a discard pattern",
			stmt:   "ALTER TABLE foo ADD COLUMN x int;",
			expect: false,
		},
		{
			name:   "empty string",
			stmt:   "",
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAlterDiscardPattern(tt.stmt)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestMatchAlterSequenceOwnedBy(t *testing.T) {
	tests := []struct {
		name   string
		stmt   string
		expect bool
	}{
		{
			name:   "ALTER SEQUENCE OWNED BY",
			stmt:   "ALTER SEQUENCE foo_seq OWNED BY foo.id;",
			expect: true,
		},
		{
			name:   "lowercase",
			stmt:   "alter sequence foo_seq owned by foo.id;",
			expect: true,
		},
		{
			name:   "not ALTER SEQUENCE",
			stmt:   "ALTER TABLE foo OWNER TO bar;",
			expect: false,
		},
		{
			name:   "empty string",
			stmt:   "",
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchAlterSequenceOwnedBy(tt.stmt)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestMatchSetDefault(t *testing.T) {
	tests := []struct {
		name    string
		stmt    string
		wantNil bool
	}{
		{
			name:    "SET DEFAULT",
			stmt:    "ALTER TABLE foo ALTER COLUMN id SET DEFAULT 1;",
			wantNil: false,
		},
		{
			name:    "not SET DEFAULT",
			stmt:    "ALTER TABLE foo ADD COLUMN x int;",
			wantNil: true,
		},
		{
			name:    "empty string",
			stmt:    "",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchSetDefault(tt.stmt)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestMatchSetNotNull(t *testing.T) {
	tests := []struct {
		name    string
		stmt    string
		wantNil bool
	}{
		{
			name:    "SET NOT NULL",
			stmt:    "ALTER TABLE foo ALTER COLUMN id SET NOT NULL;",
			wantNil: false,
		},
		{
			name:    "not SET NOT NULL",
			stmt:    "ALTER TABLE foo ADD COLUMN x int;",
			wantNil: true,
		},
		{
			name:    "empty string",
			stmt:    "",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchSetNotNull(tt.stmt)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestMatchSetType(t *testing.T) {
	tests := []struct {
		name    string
		stmt    string
		wantNil bool
	}{
		{
			name:    "SET TYPE",
			stmt:    "ALTER TABLE foo ALTER COLUMN id TYPE int;",
			wantNil: false,
		},
		{
			name:    "SET DATA TYPE",
			stmt:    "ALTER TABLE foo ALTER COLUMN id SET DATA TYPE text;",
			wantNil: false,
		},
		{
			name:    "not SET TYPE",
			stmt:    "ALTER TABLE foo ADD COLUMN x int;",
			wantNil: true,
		},
		{
			name:    "empty string",
			stmt:    "",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchSetType(tt.stmt)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestMatchDropNotNull(t *testing.T) {
	tests := []struct {
		name    string
		stmt    string
		wantNil bool
	}{
		{
			name:    "DROP NOT NULL",
			stmt:    "ALTER TABLE foo ALTER COLUMN id DROP NOT NULL;",
			wantNil: false,
		},
		{
			name:    "not DROP NOT NULL",
			stmt:    "ALTER TABLE foo ADD COLUMN x int;",
			wantNil: true,
		},
		{
			name:    "empty string",
			stmt:    "",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchDropNotNull(tt.stmt)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestMatchAddPK(t *testing.T) {
	tests := []struct {
		name    string
		stmt    string
		wantNil bool
	}{
		{
			name:    "ADD PRIMARY KEY",
			stmt:    "ALTER TABLE foo ADD CONSTRAINT pk PRIMARY KEY (id);",
			wantNil: false,
		},
		{
			name:    "not ADD PRIMARY KEY",
			stmt:    "ALTER TABLE foo ADD COLUMN x int;",
			wantNil: true,
		},
		{
			name:    "empty string",
			stmt:    "",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchAddPK(tt.stmt)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestMatchAddUnique(t *testing.T) {
	tests := []struct {
		name    string
		stmt    string
		wantNil bool
	}{
		{
			name:    "ADD UNIQUE",
			stmt:    "ALTER TABLE foo ADD CONSTRAINT uq UNIQUE (email);",
			wantNil: false,
		},
		{
			name:    "not ADD UNIQUE",
			stmt:    "ALTER TABLE foo ADD COLUMN x int;",
			wantNil: true,
		},
		{
			name:    "empty string",
			stmt:    "",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchAddUnique(tt.stmt)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestMatchAddCheck(t *testing.T) {
	tests := []struct {
		name    string
		stmt    string
		wantNil bool
	}{
		{
			name:    "ADD CHECK",
			stmt:    "ALTER TABLE foo ADD CONSTRAINT ck CHECK (amount > 0);",
			wantNil: false,
		},
		{
			name:    "not ADD CHECK",
			stmt:    "ALTER TABLE foo ADD COLUMN x int;",
			wantNil: true,
		},
		{
			name:    "empty string",
			stmt:    "",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchAddCheck(tt.stmt)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestMatchAddFK(t *testing.T) {
	tests := []struct {
		name    string
		stmt    string
		wantNil bool
	}{
		{
			name:    "ADD FOREIGN KEY",
			stmt:    "ALTER TABLE bar ADD CONSTRAINT fk FOREIGN KEY (foo_id) REFERENCES foo(id);",
			wantNil: false,
		},
		{
			name:    "not ADD FOREIGN KEY",
			stmt:    "ALTER TABLE foo ADD COLUMN x int;",
			wantNil: true,
		},
		{
			name:    "empty string",
			stmt:    "",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchAddFK(tt.stmt)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestMatchAddColumn(t *testing.T) {
	tests := []struct {
		name    string
		stmt    string
		wantNil bool
	}{
		{
			name:    "ADD COLUMN",
			stmt:    "ALTER TABLE foo ADD COLUMN new_col int;",
			wantNil: false,
		},
		{
			name:    "ADD COLUMN IF NOT EXISTS",
			stmt:    "ALTER TABLE foo ADD COLUMN IF NOT EXISTS new_col int;",
			wantNil: false,
		},
		{
			name:    "not ADD COLUMN",
			stmt:    "ALTER TABLE foo SET something = 1;",
			wantNil: true,
		},
		{
			name:    "empty string",
			stmt:    "",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchAddColumn(tt.stmt)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestMatchDropDefault(t *testing.T) {
	tests := []struct {
		name    string
		stmt    string
		wantNil bool
	}{
		{
			name:    "DROP DEFAULT",
			stmt:    "ALTER TABLE foo ALTER COLUMN id DROP DEFAULT;",
			wantNil: false,
		},
		{
			name:    "not DROP DEFAULT",
			stmt:    "ALTER TABLE foo ADD COLUMN x int;",
			wantNil: true,
		},
		{
			name:    "empty string",
			stmt:    "",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchDropDefault(tt.stmt)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestMatchAddExclude(t *testing.T) {
	tests := []struct {
		name    string
		stmt    string
		wantNil bool
	}{
		{
			name:    "ADD EXCLUDE",
			stmt:    "ALTER TABLE foo ADD CONSTRAINT ex EXCLUDE USING gist (x WITH &&);",
			wantNil: false,
		},
		{
			name:    "not ADD EXCLUDE",
			stmt:    "ALTER TABLE foo ADD COLUMN x int;",
			wantNil: true,
		},
		{
			name:    "empty string",
			stmt:    "",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchAddExclude(tt.stmt)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestHandleSetDefault(t *testing.T) {
	tests := []struct {
		name         string
		stmt         string
		tables       map[string]*model.TableDef
		expectResult AlterResult
		checkColumn  func(*testing.T, *model.ColumnDef)
	}{
		{
			name:         "SET DEFAULT matched and handled",
			stmt:         "ALTER TABLE foo ALTER COLUMN id SET DEFAULT 1;",
			tables:       map[string]*model.TableDef{"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}})},
			expectResult: AlterHandled,
			checkColumn: func(t *testing.T, c *model.ColumnDef) {
				assert.Equal(t, "1", c.Default)
			},
		},
		{
			name:         "SET DEFAULT with nextval converts to SERIAL",
			stmt:         "ALTER TABLE foo ALTER COLUMN id SET DEFAULT nextval('foo_id_seq');",
			tables:       map[string]*model.TableDef{"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}})},
			expectResult: AlterHandled,
			checkColumn: func(t *testing.T, c *model.ColumnDef) {
				assert.Equal(t, "foo_id_seq", c.SequenceName)
				assert.True(t, c.IsSerial)
			},
		},
		{
			name:         "no match returns NotMatched",
			stmt:         "ALTER TABLE foo ADD COLUMN x int;",
			tables:       map[string]*model.TableDef{"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}})},
			expectResult: AlterNotMatched,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handleSetDefault(tt.stmt, tt.tables)
			assert.Equal(t, tt.expectResult, result)
			if tt.checkColumn != nil && len(tt.tables["foo"].Columns) > 0 {
				tt.checkColumn(t, tt.tables["foo"].Columns[0])
			}
		})
	}
}

func TestHandleSetNotNull(t *testing.T) {
	tests := []struct {
		name         string
		stmt         string
		tables       map[string]*model.TableDef
		expectResult AlterResult
	}{
		{
			name:         "SET NOT NULL matched and handled",
			stmt:         "ALTER TABLE foo ALTER COLUMN id SET NOT NULL;",
			tables:       map[string]*model.TableDef{"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}})},
			expectResult: AlterHandled,
		},
		{
			name:         "no match returns NotMatched",
			stmt:         "ALTER TABLE foo ADD COLUMN x int;",
			tables:       map[string]*model.TableDef{"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}})},
			expectResult: AlterNotMatched,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handleSetNotNull(tt.stmt, tt.tables)
			assert.Equal(t, tt.expectResult, result)
		})
	}
}

func TestHandleSetType(t *testing.T) {
	tests := []struct {
		name         string
		stmt         string
		tables       map[string]*model.TableDef
		expectResult AlterResult
	}{
		{
			name:         "SET TYPE matched and handled",
			stmt:         "ALTER TABLE foo ALTER COLUMN id TYPE int;",
			tables:       map[string]*model.TableDef{"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}})},
			expectResult: AlterHandled,
		},
		{
			name:         "no match returns NotMatched",
			stmt:         "ALTER TABLE foo ADD COLUMN x int;",
			tables:       map[string]*model.TableDef{"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}})},
			expectResult: AlterNotMatched,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handleSetType(tt.stmt, tt.tables)
			assert.Equal(t, tt.expectResult, result)
		})
	}
}

func TestHandleDropDefault(t *testing.T) {
	tests := []struct {
		name         string
		stmt         string
		tables       map[string]*model.TableDef
		expectResult AlterResult
	}{
		{
			name:         "DROP DEFAULT matched and handled",
			stmt:         "ALTER TABLE foo ALTER COLUMN id DROP DEFAULT;",
			tables:       map[string]*model.TableDef{"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}})},
			expectResult: AlterHandled,
		},
		{
			name:         "no match returns NotMatched",
			stmt:         "ALTER TABLE foo ADD COLUMN x int;",
			tables:       map[string]*model.TableDef{"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}})},
			expectResult: AlterNotMatched,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handleDropDefault(tt.stmt, tt.tables)
			assert.Equal(t, tt.expectResult, result)
		})
	}
}

func TestHandleDropNotNull(t *testing.T) {
	tests := []struct {
		name         string
		stmt         string
		tables       map[string]*model.TableDef
		expectResult AlterResult
	}{
		{
			name:         "DROP NOT NULL matched and handled",
			stmt:         "ALTER TABLE foo ALTER COLUMN id DROP NOT NULL;",
			tables:       map[string]*model.TableDef{"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint NOT NULL"}})},
			expectResult: AlterHandled,
		},
		{
			name:         "no match returns NotMatched",
			stmt:         "ALTER TABLE foo ADD COLUMN x int;",
			tables:       map[string]*model.TableDef{"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}})},
			expectResult: AlterNotMatched,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handleDropNotNull(tt.stmt, tt.tables)
			assert.Equal(t, tt.expectResult, result)
		})
	}
}

func TestHandleAddPK(t *testing.T) {
	tests := []struct {
		name         string
		stmt         string
		tables       map[string]*model.TableDef
		expectResult AlterResult
	}{
		{
			name:         "ADD PRIMARY KEY matched and handled",
			stmt:         "ALTER TABLE foo ADD CONSTRAINT pk PRIMARY KEY (id);",
			tables:       map[string]*model.TableDef{"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}})},
			expectResult: AlterHandled,
		},
		{
			name:         "no match returns NotMatched",
			stmt:         "ALTER TABLE foo ADD COLUMN x int;",
			tables:       map[string]*model.TableDef{"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}})},
			expectResult: AlterNotMatched,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handleAddPK(tt.stmt, tt.tables)
			assert.Equal(t, tt.expectResult, result)
		})
	}
}

func TestHandleAddUnique(t *testing.T) {
	tests := []struct {
		name         string
		stmt         string
		tables       map[string]*model.TableDef
		expectResult AlterResult
	}{
		{
			name:         "ADD UNIQUE matched and handled",
			stmt:         "ALTER TABLE foo ADD CONSTRAINT uq UNIQUE (id);",
			tables:       map[string]*model.TableDef{"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}})},
			expectResult: AlterHandled,
		},
		{
			name:         "no match returns NotMatched",
			stmt:         "ALTER TABLE foo ADD COLUMN x int;",
			tables:       map[string]*model.TableDef{"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}})},
			expectResult: AlterNotMatched,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handleAddUnique(tt.stmt, tt.tables)
			assert.Equal(t, tt.expectResult, result)
		})
	}
}

func TestHandleAddCheck(t *testing.T) {
	tests := []struct {
		name         string
		stmt         string
		tables       map[string]*model.TableDef
		expectResult AlterResult
	}{
		{
			name:         "ADD CHECK matched and handled",
			stmt:         "ALTER TABLE foo ADD CONSTRAINT ck CHECK (amount > 0);",
			tables:       map[string]*model.TableDef{"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}})},
			expectResult: AlterHandled,
		},
		{
			name:         "no match returns NotMatched",
			stmt:         "ALTER TABLE foo ADD COLUMN x int;",
			tables:       map[string]*model.TableDef{"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}})},
			expectResult: AlterNotMatched,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handleAddCheck(tt.stmt, tt.tables)
			assert.Equal(t, tt.expectResult, result)
		})
	}
}

func TestHandleAddFK(t *testing.T) {
	tests := []struct {
		name         string
		stmt         string
		tables       map[string]*model.TableDef
		expectResult AlterResult
	}{
		{
			name:         "ADD FOREIGN KEY matched and handled",
			stmt:         "ALTER TABLE bar ADD CONSTRAINT fk FOREIGN KEY (foo_id) REFERENCES foo(id);",
			tables:       map[string]*model.TableDef{"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}}), "bar": newTable("bar", []*model.ColumnDef{{Name: "foo_id", RawDef: "bigint"}})},
			expectResult: AlterHandled,
		},
		{
			name:         "self-referential FK returns PassThrough",
			stmt:         "ALTER TABLE foo ADD CONSTRAINT fk FOREIGN KEY (parent_id) REFERENCES foo(id);",
			tables:       map[string]*model.TableDef{"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}, {Name: "parent_id", RawDef: "bigint"}})},
			expectResult: AlterPassThrough,
		},
		{
			name:         "no match returns NotMatched",
			stmt:         "ALTER TABLE foo ADD COLUMN x int;",
			tables:       map[string]*model.TableDef{"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}})},
			expectResult: AlterNotMatched,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handleAddFK(tt.stmt, tt.tables)
			assert.Equal(t, tt.expectResult, result)
		})
	}
}

func TestHandleAddColumn(t *testing.T) {
	tests := []struct {
		name         string
		stmt         string
		tables       map[string]*model.TableDef
		expectResult AlterResult
		checkColumn  func(*testing.T, *model.TableDef)
	}{
		{
			name:         "ADD COLUMN matched and handled",
			stmt:         "ALTER TABLE foo ADD COLUMN new_col int;",
			tables:       map[string]*model.TableDef{"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}})},
			expectResult: AlterHandled,
			checkColumn: func(t *testing.T, td *model.TableDef) {
				assert.Len(t, td.Columns, 2)
				assert.Equal(t, "new_col", td.Columns[1].Name)
			},
		},
		{
			name:         "FK statement returns NotMatched (not a real column add)",
			stmt:         "ALTER TABLE bar ADD CONSTRAINT fk FOREIGN KEY (foo_id) REFERENCES foo(id);",
			tables:       map[string]*model.TableDef{"bar": newTable("bar", []*model.ColumnDef{{Name: "foo_id", RawDef: "bigint"}})},
			expectResult: AlterNotMatched,
		},
		{
			name:         "no match returns NotMatched",
			stmt:         "ALTER TABLE foo SET something = 1;",
			tables:       map[string]*model.TableDef{"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}})},
			expectResult: AlterNotMatched,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handleAddColumn(tt.stmt, tt.tables)
			assert.Equal(t, tt.expectResult, result)
			if tt.checkColumn != nil {
				tt.checkColumn(t, tt.tables["foo"])
			}
		})
	}
}

func TestHandleAddExclude(t *testing.T) {
	tests := []struct {
		name         string
		stmt         string
		tables       map[string]*model.TableDef
		expectResult AlterResult
	}{
		{
			name:         "ADD EXCLUDE matched and handled",
			stmt:         "ALTER TABLE foo ADD CONSTRAINT ex EXCLUDE USING gist (x WITH &&);",
			tables:       map[string]*model.TableDef{"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}})},
			expectResult: AlterHandled,
		},
		{
			name:         "no match returns NotMatched",
			stmt:         "ALTER TABLE foo ADD COLUMN x int;",
			tables:       map[string]*model.TableDef{"foo": newTable("foo", []*model.ColumnDef{{Name: "id", RawDef: "bigint"}})},
			expectResult: AlterNotMatched,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handleAddExclude(tt.stmt, tt.tables)
			assert.Equal(t, tt.expectResult, result)
		})
	}
}
