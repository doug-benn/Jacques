package processor

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegration_Passthrough_DropTableIfExists(t *testing.T) {
	sql := "DROP TABLE IF EXISTS foo;"
	result := processDefault(sql)
	assert.Contains(t, result, "DROP TABLE IF EXISTS foo")
}

func TestIntegration_Passthrough_CreateSequence(t *testing.T) {
	sql := "CREATE SEQUENCE foo_seq;"
	result := processDefault(sql)
	assert.Contains(t, result, "CREATE SEQUENCE foo_seq")
}

func TestIntegration_Passthrough_CreateIndex(t *testing.T) {
	sql := "CREATE INDEX idx ON foo (bar);"
	result := processDefault(sql)
	assert.Contains(t, result, "CREATE INDEX idx ON foo")
}

func TestIntegration_Passthrough_CreateType(t *testing.T) {
	sql := "CREATE TYPE foo_type AS ENUM ('a', 'b');"
	result := processDefault(sql)
	assert.Contains(t, result, "CREATE TYPE foo_type")
}

func TestIntegration_FK_WithCascadeActions(t *testing.T) {
	input := `
CREATE TABLE users (id bigint NOT NULL);
CREATE TABLE orders (id bigint NOT NULL, user_id bigint NOT NULL);
ALTER TABLE orders ADD CONSTRAINT orders_user_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
`
	result := processDefault(input)
	assert.Contains(t, result, "ON DELETE CASCADE")
	assert.Contains(t, result, "REFERENCES users(id) ON DELETE CASCADE")
}

func TestIntegration_FK_WithMultipleActions(t *testing.T) {
	input := `
CREATE TABLE users (id bigint NOT NULL);
CREATE TABLE posts (id bigint NOT NULL, author_id bigint NOT NULL);
ALTER TABLE posts ADD CONSTRAINT posts_author_fkey FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE RESTRICT ON UPDATE CASCADE;
`
	result := processDefault(input)
	assert.Contains(t, result, "ON DELETE RESTRICT")
	assert.Contains(t, result, "ON UPDATE CASCADE")
}

func TestIntegration_Feature_BlockCommentRemoval(t *testing.T) {
	input := `/* This is a block comment */ CREATE TABLE foo (id int); /* Another block */`
	result := processDefault(input)
	assert.NotContains(t, result, "/*")
	assert.NotContains(t, result, "*/")
	assert.Contains(t, result, "CREATE TABLE foo")
}

func TestIntegration_Feature_BlockCommentRemoval_Multiline(t *testing.T) {
	input := `/* 
		Multi-line 
		block comment 
	*/ 
	CREATE TABLE bar (id int);`
	result := processDefault(input)
	assert.NotContains(t, result, "/*")
	assert.NotContains(t, result, "Multi-line")
	assert.Contains(t, result, "CREATE TABLE bar")
}

func TestIntegration_Feature_IfExistsForDrop(t *testing.T) {
	input := `DROP TABLE foo;
CREATE TABLE foo (id int);`
	result := processExperimental(input)
	assert.Contains(t, result, "DROP TABLE IF EXISTS foo")
	assert.NotContains(t, result, "DROP TABLE foo;")
}

func TestIntegration_Feature_IfExistsForDrop_AlreadyExists(t *testing.T) {
	input := `DROP TABLE IF EXISTS foo;
CREATE TABLE foo (id int);`
	result := processExperimental(input)
	assert.Contains(t, result, "DROP TABLE IF EXISTS foo")
}

func TestIntegration_Feature_IfExistsForDrop_Index(t *testing.T) {
	input := `DROP INDEX idx_foo;
CREATE INDEX idx_foo ON foo(id);`
	result := processExperimental(input)
	assert.Contains(t, result, "DROP INDEX IF EXISTS idx_foo")
}

func TestIntegration_Feature_AlterSequenceFilteredWhenSerial(t *testing.T) {
	input := `CREATE SEQUENCE public.order_ids START WITH 1;
CREATE TABLE public.orders (id bigint NOT NULL);
ALTER TABLE public.orders ALTER COLUMN id SET DEFAULT nextval('public.order_ids'::regclass);
ALTER SEQUENCE public.order_ids RESTART WITH 2000;`
	result := Process(input, nil)
	assert.NotContains(t, result, "ALTER SEQUENCE")
	assert.Contains(t, result, "BIGSERIAL")
}

func TestIntegration_Feature_AlterSequenceFiltered_Multiple(t *testing.T) {
	input := `CREATE SEQUENCE seq1 START WITH 1;
CREATE SEQUENCE seq2 START WITH 1;
CREATE TABLE t1 (id bigint NOT NULL);
CREATE TABLE t2 (id bigint NOT NULL);
ALTER TABLE t1 ALTER COLUMN id SET DEFAULT nextval('seq1'::regclass);
ALTER TABLE t2 ALTER COLUMN id SET DEFAULT nextval('seq2'::regclass);
ALTER SEQUENCE seq1 RESTART WITH 100;
ALTER SEQUENCE seq2 RESTART WITH 200;
ALTER SEQUENCE seq1 INCREMENT BY 5;`
	result := Process(input, nil)
	assert.NotContains(t, result, "ALTER SEQUENCE")
	assert.Contains(t, result, "BIGSERIAL")
}

func TestIntegration_Feature_AlterSequenceKeptWhenNotSerial(t *testing.T) {
	input := `CREATE SEQUENCE order_ids START WITH 1;
CREATE TABLE orders (id bigint NOT NULL);
ALTER TABLE orders ALTER COLUMN id SET DEFAULT nextval('order_ids'::regclass);
ALTER SEQUENCE order_ids RESTART WITH 2000;`
	result := Process(input, nil)
	assert.NotContains(t, result, "ALTER SEQUENCE")
	assert.Contains(t, result, "BIGSERIAL")
}

func TestIntegration_EdgeCases_EmptyAndComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   \n\t\n   ",
			expected: "",
		},
		{
			name:     "only line comments",
			input:    "-- comment\n-- another comment",
			expected: "",
		},
		{
			name:     "only block comment",
			input:    "/* this is a block comment */",
			expected: "",
		},
		{
			name:     "mixed comments and whitespace",
			input:    "-- header\n\n/* body */\n\nSELECT 1;",
			expected: "SELECT 1;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processDefault(tt.input)
			assert.Equal(t, tt.expected, strings.TrimSpace(result))
		})
	}
}

func TestIntegration_PreprocessSQL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "block comments removed",
			input:    "/* comment */ SELECT 1;",
			expected: " SELECT 1;\n",
		},
		{
			name:     "line comments removed",
			input:    "-- comment\nSELECT 1;",
			expected: "SELECT 1;\n",
		},
		{
			name:     "mixed comments removed",
			input:    "/* block */ SELECT 1; -- line",
			expected: " SELECT 1; \n",
		},
		{
			name:     "comment in middle of line",
			input:    "SELECT 1; -- comment",
			expected: "SELECT 1; \n",
		},
		{
			name:     "multiline block comment",
			input:    "/*\nmulti\nline\n*/ SELECT 1;",
			expected: " SELECT 1;\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := preprocessSQL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIntegration_RemoveBlockComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple block comment",
			input:    "/* comment */ SELECT 1;",
			expected: " SELECT 1;",
		},
		{
			name:     "multiline block comment",
			input:    "/*\nmulti\nline\n*/ SELECT 1;",
			expected: " SELECT 1;",
		},
		{
			name:     "no block comments",
			input:    "SELECT 1;",
			expected: "SELECT 1;",
		},
		{
			name:     "multiple block comments",
			input:    "/* a */ SELECT 1; /* b */",
			expected: " SELECT 1; ",
		},
		{
			name:     "block comment in string literal",
			input:    "SELECT '/* not a comment */';",
			expected: "SELECT '';",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeBlockComments(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIntegration_RemoveLineComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "line comment at start",
			input:    "-- comment\nSELECT 1;",
			expected: "SELECT 1;\n",
		},
		{
			name:     "line comment at end",
			input:    "SELECT 1; -- comment",
			expected: "SELECT 1; \n",
		},
		{
			name:     "line comment in middle",
			input:    "SELECT 1; -- comment\nSELECT 2;",
			expected: "SELECT 1; \nSELECT 2;\n",
		},
		{
			name:     "no line comments",
			input:    "SELECT 1;",
			expected: "SELECT 1;\n",
		},
		{
			name:     "multiple line comments",
			input:    "-- comment 1\n-- comment 2\nSELECT 1;",
			expected: "SELECT 1;\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeLineComments(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
