package cleaner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/doug-benn/Jacques/internal/model"
)

func TestCleanRawDef(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "trailing comma removed",
			input: "int,",
			want:  "int",
		},
		{
			name:  "whitespace normalized",
			input: "  bigint  ",
			want:  "bigint",
		},
		{
			name:  "multiple spaces normalized",
			input: "bigint   not   null",
			want:  "bigint not null",
		},
		{
			name:  "no change needed",
			input: "int",
			want:  "int",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanRawDef(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestRemoveOnlyKeyword(t *testing.T) {
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

func TestRemoveNotNullAfterPrimaryKey(t *testing.T) {
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

func TestIsColumnDef(t *testing.T) {
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

func TestTransform(t *testing.T) {
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

func TestRenderColumnWithFKCascade(t *testing.T) {
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

func TestRenderTableWithFKCascade(t *testing.T) {
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

	assert.Contains(t, result, "CREATE TABLE public.orders")
	assert.Contains(t, result, "user_id bigint REFERENCES users(id) ON DELETE CASCADE")
}
