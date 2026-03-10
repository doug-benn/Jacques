package cleaner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindTableBody(t *testing.T) {
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
			name:      "nested parentheses",
			stmt:      "CREATE TABLE users (id bigint CHECK (id > 0))",
			wantStart: 20,
			wantEnd:   44,
		},
		{
			name:      "no opening paren",
			stmt:      "CREATE TABLE users",
			wantStart: 0,
			wantEnd:   0,
		},
		{
			name:      "unclosed paren",
			stmt:      "CREATE TABLE users (id int",
			wantStart: 20,
			wantEnd:   26,
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

func TestParseTableBody(t *testing.T) {
	tests := []struct {
		name            string
		body            string
		wantCols        int
		wantConstraints int
	}{
		{
			name:            "only columns",
			body:            "id bigint, name text",
			wantCols:        2,
			wantConstraints: 0,
		},
		{
			name:            "columns and constraints",
			body:            "id bigint, name text, PRIMARY KEY (id)",
			wantCols:        2,
			wantConstraints: 1,
		},
		{
			name:            "nested parens in segments",
			body:            "id bigint CHECK (id > 0), UNIQUE (id)",
			wantCols:        1,
			wantConstraints: 1,
		},
		{
			name:            "empty body",
			body:            "",
			wantCols:        0,
			wantConstraints: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cols, constraints, err := parseTableBody(tt.body)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantCols, len(cols))
			assert.Equal(t, tt.wantConstraints, len(constraints))
		})
	}
}

func TestIsColumnDef(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"column", "id bigint", true},
		{"primary key", "PRIMARY KEY (id)", false},
		{"unique", "UNIQUE (id)", false},
		{"check", "CHECK (id > 0)", false},
		{"foreign key", "FOREIGN KEY (id) REFERENCES f(id)", false},
		{"constraint", "CONSTRAINT pk PRIMARY KEY (id)", false},
		{"short", "id", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isColumnDef(tt.input))
		})
	}
}

func TestIsPartitionOf(t *testing.T) {
	tests := []struct {
		name string
		stmt string
		want bool
	}{
		{"partition", "CREATE TABLE x PARTITION OF y", true},
		{"not partition", "CREATE TABLE x (id int)", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isPartitionOf(tt.stmt))
		})
	}
}

func TestExtractTableName(t *testing.T) {
	tests := []struct {
		name       string
		stmt       string
		wantSchema string
		wantName   string
	}{
		{"qualified", "CREATE TABLE public.users (", "public", "users"},
		{"unqualified", "CREATE TABLE users (", "", "users"},
		{"if not exists", "CREATE TABLE IF NOT EXISTS users (", "", "users"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, name := extractTableName(tt.stmt)
			assert.Equal(t, tt.wantSchema, schema)
			assert.Equal(t, tt.wantName, name)
		})
	}
}

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

func TestNeedsTrailingComma(t *testing.T) {
	tests := []struct {
		name         string
		currentIndex int
		totalCount   int
		hasMoreAfter bool
		want         bool
	}{
		{
			name:         "first of three",
			currentIndex: 0,
			totalCount:   3,
			hasMoreAfter: false,
			want:         true,
		},
		{
			name:         "middle of three",
			currentIndex: 1,
			totalCount:   3,
			hasMoreAfter: false,
			want:         true,
		},
		{
			name:         "last of three",
			currentIndex: 2,
			totalCount:   3,
			hasMoreAfter: false,
			want:         false,
		},
		{
			name:         "last with more after",
			currentIndex: 2,
			totalCount:   3,
			hasMoreAfter: true,
			want:         true,
		},
		{
			name:         "single item no more",
			currentIndex: 0,
			totalCount:   1,
			hasMoreAfter: false,
			want:         false,
		},
		{
			name:         "single item with more after",
			currentIndex: 0,
			totalCount:   1,
			hasMoreAfter: true,
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := needsTrailingComma(tt.currentIndex, tt.totalCount, tt.hasMoreAfter)
			assert.Equal(t, tt.want, result)
		})
	}
}
