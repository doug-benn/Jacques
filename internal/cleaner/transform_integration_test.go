package cleaner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/doug-benn/Jacques/internal/model"
)

func TestIntegration_RemoveOnlyKeyword(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "ONLY keyword removed",
			input: "ALTER TABLE ONLY x ADD y;",
			want:  "ALTER TABLE x ADD y;",
		},
		{
			name:  "no ONLY keyword",
			input: "ALTER TABLE x ADD y;",
			want:  "ALTER TABLE x ADD y;",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeOnlyKeyword(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestIntegration_RemoveNotNullAfterPrimaryKey(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "NOT NULL after PRIMARY KEY removed",
			input: "PRIMARY KEY (id) NOT NULL",
			want:  "PRIMARY KEY (ID) ",
		},
		{
			name:  "no PRIMARY KEY NOT NULL",
			input: "PRIMARY KEY (id)",
			want:  "PRIMARY KEY (id)",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeNotNullAfterPrimaryKey(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestIntegration_IsColumnDef(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect bool
	}{
		{
			name:   "basic column",
			input:  "id bigint",
			expect: true,
		},
		{
			name:   "column with constraints",
			input:  "name text not null",
			expect: true,
		},
		{
			name:   "PRIMARY KEY constraint",
			input:  "PRIMARY KEY (id)",
			expect: false,
		},
		{
			name:   "UNIQUE constraint",
			input:  "UNIQUE (email)",
			expect: false,
		},
		{
			name:   "CHECK constraint",
			input:  "CHECK (amount > 0)",
			expect: false,
		},
		{
			name:   "FOREIGN KEY constraint",
			input:  "FOREIGN KEY (user_id) REFERENCES users(id)",
			expect: false,
		},
		{
			name:   "CONSTRAINT",
			input:  "CONSTRAINT pk PRIMARY KEY (id)",
			expect: false,
		},
		{
			name:   "single word",
			input:  "id",
			expect: false,
		},
		{
			name:   "empty string",
			input:  "",
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isColumnDef(tt.input)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestIntegration_Transform(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(*testing.T, string)
	}{
		{
			name:  "ONLY keyword removal",
			input: "ALTER TABLE ONLY public.users ADD CONSTRAINT pk PRIMARY KEY (id);",
			check: func(t *testing.T, result string) {
				assert.NotContains(t, result, "ONLY")
			},
		},
		{
			name:  "NOT NULL after PRIMARY KEY removal",
			input: "ALTER TABLE ONLY public.users ADD CONSTRAINT pk PRIMARY KEY (id);",
			check: func(t *testing.T, result string) {
				assert.NotContains(t, result, "NOT NULL")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Transform(tt.input)
			tt.check(t, result)
		})
	}
}

func TestIntegration_RenderColumnWithFKCascade(t *testing.T) {
	tests := []struct {
		name             string
		col              *model.ColumnDef
		expectedContains string
	}{
		{
			name: "FK with ON DELETE CASCADE",
			col: &model.ColumnDef{
				Name:       "user_id",
				RawDef:     "bigint",
				References: "users(id)",
				OnDelete:   "CASCADE",
			},
			expectedContains: "REFERENCES users(id) ON DELETE CASCADE",
		},
		{
			name: "FK with ON DELETE SET NULL",
			col: &model.ColumnDef{
				Name:       "user_id",
				RawDef:     "bigint",
				References: "users(id)",
				OnDelete:   "SET NULL",
			},
			expectedContains: "REFERENCES users(id) ON DELETE SET NULL",
		},
		{
			name: "FK with ON DELETE CASCADE ON UPDATE RESTRICT",
			col: &model.ColumnDef{
				Name:       "user_id",
				RawDef:     "bigint",
				References: "users(id)",
				OnDelete:   "CASCADE",
				OnUpdate:   "RESTRICT",
			},
			expectedContains: "REFERENCES users(id) ON DELETE CASCADE ON UPDATE RESTRICT",
		},
		{
			name: "FK with MATCH FULL",
			col: &model.ColumnDef{
				Name:       "user_id",
				RawDef:     "bigint",
				References: "users(id)",
				Match:      "FULL",
			},
			expectedContains: "REFERENCES users(id) MATCH FULL",
		},
		{
			name: "FK with all actions",
			col: &model.ColumnDef{
				Name:       "user_id",
				RawDef:     "bigint",
				References: "users(id)",
				OnDelete:   "CASCADE",
				OnUpdate:   "SET NULL",
				Match:      "FULL",
			},
			expectedContains: "REFERENCES users(id) ON DELETE CASCADE ON UPDATE SET NULL MATCH FULL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderColumn(tt.col)
			assert.Contains(t, result, tt.expectedContains, "rendered column should contain cascade actions")
		})
	}
}

func TestIntegration_RenderTableWithFKCascade(t *testing.T) {
	td := &model.TableDef{
		Schema: "public",
		Name:   "orders",
		Columns: []*model.ColumnDef{
			{
				Name:         "id",
				RawDef:       "bigint",
				IsPrimaryKey: true,
			},
			{
				Name:       "user_id",
				RawDef:     "bigint",
				References: "users(id)",
				OnDelete:   "CASCADE",
			},
		},
	}

	result := RenderTable(td)
	require.NotEmpty(t, result)

	assert.Contains(t, result, "CREATE TABLE orders")
	assert.Contains(t, result, "user_id bigint REFERENCES users(id) ON DELETE CASCADE")
}

func TestIntegration_RenderPrimaryKey(t *testing.T) {
	tests := []struct {
		name     string
		td       *model.TableDef
		want     string
		wantBool bool
	}{
		{
			name:     "no pk",
			td:       &model.TableDef{},
			want:     "",
			wantBool: false,
		},
		{
			name: "with pk",
			td: &model.TableDef{
				TableLevelPK: "id",
			},
			want:     "    PRIMARY KEY (id)",
			wantBool: true,
		},
		{
			name: "with composite pk",
			td: &model.TableDef{
				TableLevelPK: "id, user_id",
			},
			want:     "    PRIMARY KEY (id, user_id)",
			wantBool: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := renderPrimaryKey(tt.td)
			assert.Equal(t, tt.want, result)
			assert.Equal(t, tt.wantBool, ok)
		})
	}
}

func TestIntegration_RenderUniqueConstraints(t *testing.T) {
	tests := []struct {
		name     string
		td       *model.TableDef
		want     []string
		wantBool bool
	}{
		{
			name:     "no uniques",
			td:       &model.TableDef{},
			want:     nil,
			wantBool: false,
		},
		{
			name: "with unique",
			td: &model.TableDef{
				TableLevelUniques: []string{"email"},
			},
			want:     []string{"    UNIQUE (email)"},
			wantBool: true,
		},
		{
			name: "with multiple uniques",
			td: &model.TableDef{
				TableLevelUniques: []string{"email", "phone"},
			},
			want:     []string{"    UNIQUE (email)", "    UNIQUE (phone)"},
			wantBool: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := renderUniqueConstraints(tt.td)
			assert.Equal(t, tt.want, result)
			assert.Equal(t, tt.wantBool, ok)
		})
	}
}

func TestIntegration_RenderTableConstraints(t *testing.T) {
	tests := []struct {
		name     string
		td       *model.TableDef
		want     []string
		wantBool bool
	}{
		{
			name:     "no constraints",
			td:       &model.TableDef{},
			want:     nil,
			wantBool: false,
		},
		{
			name: "with check constraint",
			td: &model.TableDef{
				TableConstraints: []string{"CHECK (amount > 0)"},
			},
			want:     []string{"    CHECK (amount > 0)"},
			wantBool: true,
		},
		{
			name: "with multiple constraints",
			td: &model.TableDef{
				TableConstraints: []string{"CHECK (amount > 0)", "CHECK (status IN ('a', 'b'))"},
			},
			want:     []string{"    CHECK (amount > 0)", "    CHECK (status IN ('a', 'b'))"},
			wantBool: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := renderTableConstraints(tt.td)
			assert.Equal(t, tt.want, result)
			assert.Equal(t, tt.wantBool, ok)
		})
	}
}

func TestIntegration_RenderTableLevelFKs(t *testing.T) {
	tests := []struct {
		name     string
		td       *model.TableDef
		want     []string
		wantBool bool
	}{
		{
			name:     "no fks",
			td:       &model.TableDef{},
			want:     nil,
			wantBool: false,
		},
		{
			name: "with table-level fk",
			td: &model.TableDef{
				TableLevelFKs: []string{"FOREIGN KEY (user_id) REFERENCES users(id)"},
			},
			want:     []string{"    FOREIGN KEY (user_id) REFERENCES users(id)"},
			wantBool: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := renderTableLevelFKs(tt.td)
			assert.Equal(t, tt.want, result)
			assert.Equal(t, tt.wantBool, ok)
		})
	}
}

func TestIntegration_RenderExclusions(t *testing.T) {
	tests := []struct {
		name     string
		td       *model.TableDef
		want     []string
		wantBool bool
	}{
		{
			name:     "no exclusions",
			td:       &model.TableDef{},
			want:     nil,
			wantBool: false,
		},
		{
			name: "with exclusion",
			td: &model.TableDef{
				TableExclusions: []string{"EXCLUDE USING gist (name WITH =)"},
			},
			want:     []string{"    EXCLUDE USING gist (name WITH =)"},
			wantBool: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := renderExclusions(tt.td)
			assert.Equal(t, tt.want, result)
			assert.Equal(t, tt.wantBool, ok)
		})
	}
}

func TestIntegration_HasTableSuffixConstraints(t *testing.T) {
	tests := []struct {
		name string
		td   *model.TableDef
		want bool
	}{
		{
			name: "no constraints",
			td:   &model.TableDef{},
			want: false,
		},
		{
			name: "with pk",
			td: &model.TableDef{
				TableLevelPK: "id",
			},
			want: true,
		},
		{
			name: "with unique",
			td: &model.TableDef{
				TableLevelUniques: []string{"email"},
			},
			want: true,
		},
		{
			name: "with check",
			td: &model.TableDef{
				TableConstraints: []string{"CHECK (x > 0)"},
			},
			want: true,
		},
		{
			name: "with fk",
			td: &model.TableDef{
				TableLevelFKs: []string{"FOREIGN KEY (x) REFERENCES y(id)"},
			},
			want: true,
		},
		{
			name: "with exclusion",
			td: &model.TableDef{
				TableExclusions: []string{"EXCLUDE (x WITH =)"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasTableSuffixConstraints(tt.td)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestIntegration_RenderSerialColumn(t *testing.T) {
	tests := []struct {
		name string
		col  *model.ColumnDef
		want string
	}{
		{
			name: "bigint serial",
			col: &model.ColumnDef{
				RawDef:       "bigint",
				IsPrimaryKey: true,
			},
			want: "BIGSERIAL PRIMARY KEY",
		},
		{
			name: "integer serial",
			col: &model.ColumnDef{
				RawDef: "integer",
			},
			want: "SERIAL",
		},
		{
			name: "smallint serial",
			col: &model.ColumnDef{
				RawDef: "smallint",
			},
			want: "SMALLSERIAL",
		},
		{
			name: "bigint non-serial (not converted)",
			col: &model.ColumnDef{
				RawDef: "bigint",
			},
			want: "BIGSERIAL",
		},
		{
			name: "non-pk bigint serial",
			col: &model.ColumnDef{
				RawDef:       "bigint",
				IsPrimaryKey: false,
			},
			want: "BIGSERIAL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderSerialColumn(tt.col)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestIntegration_RenderReferences(t *testing.T) {
	tests := []struct {
		name string
		col  *model.ColumnDef
		want string
	}{
		{
			name: "basic reference",
			col: &model.ColumnDef{
				References: "users(id)",
			},
			want: " REFERENCES users(id)",
		},
		{
			name: "with on delete",
			col: &model.ColumnDef{
				References: "users(id)",
				OnDelete:   "CASCADE",
			},
			want: " REFERENCES users(id) ON DELETE CASCADE",
		},
		{
			name: "with on update",
			col: &model.ColumnDef{
				References: "users(id)",
				OnUpdate:   "RESTRICT",
			},
			want: " REFERENCES users(id) ON UPDATE RESTRICT",
		},
		{
			name: "with match",
			col: &model.ColumnDef{
				References: "users(id)",
				Match:      "FULL",
			},
			want: " REFERENCES users(id) MATCH FULL",
		},
		{
			name: "full cascade",
			col: &model.ColumnDef{
				References: "users(id)",
				OnDelete:   "CASCADE",
				OnUpdate:   "SET NULL",
				Match:      "FULL",
			},
			want: " REFERENCES users(id) ON DELETE CASCADE ON UPDATE SET NULL MATCH FULL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderReferences(tt.col)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestIntegration_ParseSequence(t *testing.T) {
	tests := []struct {
		name  string
		col   *model.ColumnDef
		check func(*testing.T, *model.ColumnDef)
	}{
		{
			name: "bigint with nextval becomes serial",
			col:  &model.ColumnDef{RawDef: "bigint nextval('myseq')"},
			check: func(t *testing.T, col *model.ColumnDef) {
				assert.Equal(t, "myseq", col.SequenceName)
				assert.True(t, col.IsSerial)
				assert.Contains(t, col.Default, "nextval")
			},
		},
		{
			name: "integer with nextval becomes serial",
			col:  &model.ColumnDef{RawDef: "integer nextval('myseq')"},
			check: func(t *testing.T, col *model.ColumnDef) {
				assert.Equal(t, "myseq", col.SequenceName)
				assert.True(t, col.IsSerial)
			},
		},
		{
			name: "smallint with nextval NOT serial",
			col:  &model.ColumnDef{RawDef: "smallint nextval('myseq')"},
			check: func(t *testing.T, col *model.ColumnDef) {
				assert.False(t, col.IsSerial)
			},
		},
		{
			name: "bigint without nextval unchanged",
			col:  &model.ColumnDef{RawDef: "bigint"},
			check: func(t *testing.T, col *model.ColumnDef) {
				assert.Empty(t, col.SequenceName)
				assert.False(t, col.IsSerial)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parseSequence(tt.col)
			tt.check(t, tt.col)
		})
	}
}

func TestIntegration_ParsePrimaryKey(t *testing.T) {
	tests := []struct {
		name string
		col  *model.ColumnDef
		want bool
	}{
		{
			name: "primary key in rawdef",
			col:  &model.ColumnDef{RawDef: "bigint PRIMARY KEY"},
			want: true,
		},
		{
			name: "primary key uppercase",
			col:  &model.ColumnDef{RawDef: "bigint primary key"},
			want: true,
		},
		{
			name: "no primary key",
			col:  &model.ColumnDef{RawDef: "bigint"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsePrimaryKey(tt.col)
			assert.Equal(t, tt.want, tt.col.IsPrimaryKey)
		})
	}
}

func TestIntegration_ParseUnique(t *testing.T) {
	tests := []struct {
		name string
		col  *model.ColumnDef
		want bool
	}{
		{
			name: "unique in rawdef",
			col:  &model.ColumnDef{RawDef: "varchar(255) UNIQUE"},
			want: true,
		},
		{
			name: "unique lowercase",
			col:  &model.ColumnDef{RawDef: "varchar(255) unique"},
			want: true,
		},
		{
			name: "no unique",
			col:  &model.ColumnDef{RawDef: "bigint"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parseUnique(tt.col)
			assert.Equal(t, tt.want, tt.col.IsUnique)
		})
	}
}

func TestIntegration_ParseReferences(t *testing.T) {
	tests := []struct {
		name  string
		col   *model.ColumnDef
		check func(*testing.T, *model.ColumnDef)
	}{
		{
			name: "basic references",
			col:  &model.ColumnDef{RawDef: "bigint REFERENCES users(id)"},
			check: func(t *testing.T, col *model.ColumnDef) {
				assert.Equal(t, "users(id)", col.References)
			},
		},
		{
			name: "references without column defaults to id",
			col:  &model.ColumnDef{RawDef: "bigint REFERENCES users"},
			check: func(t *testing.T, col *model.ColumnDef) {
				assert.Equal(t, "users(id)", col.References)
			},
		},
		{
			name: "with on delete cascade",
			col:  &model.ColumnDef{RawDef: "bigint REFERENCES users(id) ON DELETE CASCADE"},
			check: func(t *testing.T, col *model.ColumnDef) {
				assert.Equal(t, "CASCADE", col.OnDelete)
			},
		},
		{
			name: "with on update restrict",
			col:  &model.ColumnDef{RawDef: "bigint REFERENCES users(id) ON UPDATE RESTRICT"},
			check: func(t *testing.T, col *model.ColumnDef) {
				assert.Equal(t, "RESTRICT", col.OnUpdate)
			},
		},
		{
			name: "with match full",
			col:  &model.ColumnDef{RawDef: "bigint REFERENCES users(id) MATCH FULL"},
			check: func(t *testing.T, col *model.ColumnDef) {
				assert.Equal(t, "FULL", col.Match)
			},
		},
		{
			name: "no references",
			col:  &model.ColumnDef{RawDef: "bigint"},
			check: func(t *testing.T, col *model.ColumnDef) {
				assert.Empty(t, col.References)
				assert.Empty(t, col.OnDelete)
				assert.Empty(t, col.OnUpdate)
				assert.Empty(t, col.Match)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parseReferences(tt.col)
			tt.check(t, tt.col)
		})
	}
}

func TestIntegration_IsPartitionOf(t *testing.T) {
	tests := []struct {
		name string
		stmt string
		want bool
	}{
		{
			name: "partition of child",
			stmt: "CREATE TABLE orders PARTITION OF orders_parent FOR VALUES IN ('a')",
			want: true,
		},
		{
			name: "regular create table",
			stmt: "CREATE TABLE users (id bigint)",
			want: false,
		},
		{
			name: "if not exists partition of",
			stmt: "CREATE TABLE IF NOT EXISTS orders PARTITION OF orders_parent",
			want: true,
		},
		{
			name: "empty string",
			stmt: "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPartitionOf(tt.stmt)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestIntegration_ExtractTableName(t *testing.T) {
	tests := []struct {
		name       string
		stmt       string
		wantSchema string
		wantName   string
	}{
		{
			name:       "qualified table",
			stmt:       "CREATE TABLE public.users (",
			wantSchema: "public",
			wantName:   "users",
		},
		{
			name:       "unqualified table",
			stmt:       "CREATE TABLE users (",
			wantSchema: "",
			wantName:   "users",
		},
		{
			name:       "if not exists",
			stmt:       "CREATE TABLE IF NOT EXISTS users (",
			wantSchema: "",
			wantName:   "users",
		},
		{
			name:       "qualified if not exists",
			stmt:       "CREATE TABLE IF NOT EXISTS schema.table_name (",
			wantSchema: "schema",
			wantName:   "table_name",
		},
		{
			name:       "not a create table",
			stmt:       "SELECT * FROM users",
			wantSchema: "",
			wantName:   "",
		},
		{
			name:       "empty string",
			stmt:       "",
			wantSchema: "",
			wantName:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, name := extractTableName(tt.stmt)
			assert.Equal(t, tt.wantSchema, schema)
			assert.Equal(t, tt.wantName, name)
		})
	}
}

func TestIntegration_FindTableBody(t *testing.T) {
	tests := []struct {
		name      string
		stmt      string
		wantStart int
		wantEnd   int
	}{
		{
			name:      "simple table",
			stmt:      "CREATE TABLE users (id bigint)",
			wantStart: 20,
			wantEnd:   29,
		},
		{
			name:      "table with nested parens",
			stmt:      "CREATE TABLE users (id bigint CHECK (id > 0))",
			wantStart: 20,
			wantEnd:   44,
		},
		{
			name:      "no parens",
			stmt:      "CREATE TABLE users",
			wantStart: 0,
			wantEnd:   0,
		},
		{
			name:      "empty string",
			stmt:      "",
			wantStart: 0,
			wantEnd:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := findTableBody(tt.stmt)
			assert.Equal(t, tt.wantStart, start)
			assert.Equal(t, tt.wantEnd, end)
		})
	}
}
