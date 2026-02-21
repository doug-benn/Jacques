package cleaner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/doug-benn/Jacques/internal/model"
)

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
