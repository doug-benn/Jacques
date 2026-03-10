package processor

import (
	"testing"

	"github.com/doug-benn/Jacques/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeIdentifier(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "unquoted lowercase",
			input: "myschema",
			want:  "myschema",
		},
		{
			name:  "unquoted uppercase",
			input: "MYSCHEMA",
			want:  "myschema",
		},
		{
			name:  "unquoted mixed case",
			input: "MySchema",
			want:  "myschema",
		},
		{
			name:  "quoted case-sensitive",
			input: "\"MySchema\"",
			want:  "MySchema",
		},
		{
			name:  "quoted lowercase",
			input: "\"myschema\"",
			want:  "myschema",
		},
		{
			name:  "quoted with special characters",
			input: "\"schema-name\"",
			want:  "schema-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeIdentifier(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestNormalizeSequenceName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no schema",
			input: "my_seq",
			want:  "my_seq",
		},
		{
			name:  "with schema",
			input: "public.my_seq",
			want:  "my_seq",
		},
		{
			name:  "nested schema",
			input: "app.public.my_seq",
			want:  "my_seq",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeSequenceName(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestNormalizeIndexDef(t *testing.T) {
	tests := []struct {
		name string
		stmt string
		want string
	}{
		{
			name: "basic index",
			stmt: "CREATE INDEX idx ON public.users (email)",
			want: "public.users(email)",
		},
		{
			name: "unique index",
			stmt: "CREATE UNIQUE INDEX idx ON public.users (email)",
			want: "UNIQUE public.users(email)",
		},
		{
			name: "multi-column index",
			stmt: "CREATE INDEX idx ON public.users (first_name, last_name)",
			want: "public.users(first_name,last_name)",
		},
		{
			name: "index with WHERE",
			stmt: "CREATE INDEX idx ON public.users (email) WHERE active = true",
			want: "public.users(email)[WHERE active = true]",
		},
		{
			name: "index with INCLUDE",
			stmt: "CREATE INDEX idx ON public.users (email) INCLUDE (id, name)",
			want: "public.users(email)[INCLUDE(id,name)]",
		},
		{
			name: "quoted identifiers",
			stmt: "CREATE INDEX idx ON \"MySchema\".\"Users\" (\"FirstName\")",
			want: "\"MySchema\".\"Users\"(\"firstname\")",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeIndexDef(tt.stmt)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestIsRedundantIndex(t *testing.T) {
	implicit := map[string]bool{
		"public.users.id":    true,
		"public.users.email": true,
	}

	tests := []struct {
		name string
		stmt string
		want bool
	}{
		{
			name: "redundant simple index",
			stmt: "CREATE INDEX idx ON public.users (id)",
			want: true,
		},
		{
			name: "not redundant unique index",
			stmt: "CREATE UNIQUE INDEX idx ON public.users (id)",
			want: false,
		},
		{
			name: "not redundant multi-column",
			stmt: "CREATE INDEX idx ON public.users (id, other)",
			want: false,
		},
		{
			name: "not redundant partial",
			stmt: "CREATE INDEX idx ON public.users (id) WHERE x > 0",
			want: false,
		},
		{
			name: "not redundant different table",
			stmt: "CREATE INDEX idx ON public.other (id)",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRedundantIndex(tt.stmt, implicit)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestBuildImplicitIndexMap(t *testing.T) {
	tables := map[string]*model.TableDef{
		"public.users": {
			Schema: "public",
			Name:   "users",
			Columns: []*model.ColumnDef{
				{Name: "id", IsPrimaryKey: true},
				{Name: "email", IsUnique: true},
				{Name: "name"},
			},
		},
		"app.orders": {
			Schema:       "app",
			Name:         "orders",
			TableLevelPK: "id, user_id",
			TableLevelUniques: []string{
				"order_number",
			},
			Columns: []*model.ColumnDef{
				{Name: "id"},
				{Name: "user_id"},
				{Name: "order_number"},
			},
		},
	}

	result := buildImplicitIndexMap(tables)

	assert.True(t, result["public.users.id"])
	assert.True(t, result["public.users.email"])
	assert.False(t, result["public.users.name"])
	assert.True(t, result["app.orders.id"])
	assert.True(t, result["app.orders.user_id"])
	assert.True(t, result["app.orders.order_number"])
}
