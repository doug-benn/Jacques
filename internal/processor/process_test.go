package processor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func processDefault(sql string) string {
	return Process(sql, nil)
}

func processExperimental(sql string) string {
	return Process(sql, &Options{ExperimentalFolding: true})
}

func loadTestFixture(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "e2e", name)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(data)
}

func loadIntegrationFixture(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "integration", name)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(data)
}

func normalizeSQL(sql string) string {
	sql = strings.ReplaceAll(sql, "\r\n", "\n")
	return strings.TrimSpace(sql)
}

func extractStatements(sql string) []string {
	var statements []string
	var current strings.Builder
	lines := strings.Split(sql, "\n")
	inBlockComment := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "/*") {
			inBlockComment = true
		}
		if strings.HasSuffix(trimmed, "*/") {
			inBlockComment = false
			continue
		}
		if inBlockComment || trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}

		current.WriteString(line)
		current.WriteString("\n")

		if strings.HasSuffix(strings.TrimSpace(line), ";") {
			stmt := strings.TrimSpace(current.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
		}
	}

	return statements
}

func assertStatementContains(t *testing.T, sql, keyword, context string) {
	t.Helper()
	statements := extractStatements(sql)
	found := false
	for _, stmt := range statements {
		if strings.Contains(stmt, keyword) {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected to find statement containing %s for %s", keyword, context)
}

func assertStatementOrder(t *testing.T, sql, firstKeyword, secondKeyword string) {
	t.Helper()
	statements := extractStatements(sql)
	firstIdx := -1
	secondIdx := -1

	for i, stmt := range statements {
		if strings.Contains(stmt, firstKeyword) && firstIdx == -1 {
			firstIdx = i
		}
		if strings.Contains(stmt, secondKeyword) && secondIdx == -1 {
			secondIdx = i
		}
	}

	if firstIdx >= 0 && secondIdx >= 0 {
		assert.Less(t, firstIdx, secondIdx, "Expected %s to come before %s", firstKeyword, secondKeyword)
	}
}

func assertTypeBeforeTable(t *testing.T, sql, typeName, tableName string) {
	t.Helper()
	assertStatementOrder(t, sql, "CREATE TYPE "+typeName, "CREATE TABLE "+tableName)
}

func assertViewPreserved(t *testing.T, sql, viewName string) {
	t.Helper()
	assertStatementContains(t, sql, "CREATE VIEW "+viewName, "View "+viewName)
}

func assertMaterializedViewPreserved(t *testing.T, sql, viewName string) {
	t.Helper()
	assertStatementContains(t, sql, "CREATE MATERIALIZED VIEW "+viewName, "Materialized view "+viewName)
}

func assertIndexPreserved(t *testing.T, sql, indexName string) {
	t.Helper()
	assertStatementContains(t, sql, "CREATE INDEX "+indexName, "Index "+indexName)
}

func TestExtractSequenceName(t *testing.T) {
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

func TestExtractAlterSequenceName(t *testing.T) {
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

func TestRemoveBlockComments(t *testing.T) {
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

func TestRemoveLineComments(t *testing.T) {
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

func TestPreprocessSQL(t *testing.T) {
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
