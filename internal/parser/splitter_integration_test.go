package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_SplitStatements(t *testing.T) {
	tests := []struct {
		name         string
		sql          string
		wantCount    int
		wantFirst    string
		wantContains string
	}{
		{
			name:      "basic semicolons",
			sql:       "SELECT 1; SELECT 2;",
			wantCount: 2,
			wantFirst: "SELECT 1;",
		},
		{
			name:      "single quotes",
			sql:       "SELECT 'a;b';",
			wantCount: 1,
			wantFirst: "SELECT 'a;b';",
		},
		{
			name:      "escaped quotes",
			sql:       "SELECT 'it''s';",
			wantCount: 1,
			wantFirst: "SELECT 'it''s';",
		},
		{
			name:      "dollar quotes",
			sql:       "SELECT $$a;b$$;",
			wantCount: 1,
			wantFirst: "SELECT $$a;b$$;",
		},
		{
			name:      "tagged dollar quotes",
			sql:       "$tag$text;here$tag$;",
			wantCount: 1,
			wantFirst: "$tag$text;here$tag$;",
		},
		{
			name:         "block comments",
			sql:          "/* a;b */ SELECT 1;",
			wantCount:    1,
			wantContains: "SELECT 1",
		},
		{
			name:         "line comments",
			sql:          "-- comment\nSELECT 1;",
			wantCount:    1,
			wantContains: "SELECT 1",
		},
		{
			name:         "unterminated dollar quote",
			sql:          "SELECT $$unclosed",
			wantCount:    1,
			wantContains: "$$unclosed",
		},
		{
			name:         "unterminated single quote",
			sql:          "SELECT 'unclosed",
			wantCount:    1,
			wantContains: "'unclosed",
		},
		{
			name:      "empty string",
			sql:       "",
			wantCount: 0,
		},
		{
			name:      "only semicolons",
			sql:       ";;;",
			wantCount: 0,
		},
		{
			name:      "multiple statements with whitespace",
			sql:       "  SELECT 1;   SELECT 2;  ",
			wantCount: 2,
			wantFirst: "SELECT 1;",
		},
		{
			name:      "nested parentheses in statement",
			sql:       "SELECT ((1)); SELECT 2;",
			wantCount: 2,
			wantFirst: "SELECT ((1));",
		},
		{
			name:      "multiple statements in one line",
			sql:       "SELECT 1; SELECT 2; SELECT 3;",
			wantCount: 3,
			wantFirst: "SELECT 1;",
		},
		{
			name:         "dollar quote with semicolon inside",
			sql:          "SELECT $$a;b$$; SELECT 2;",
			wantCount:    2,
			wantContains: "$$a;b$$",
		},
		{
			name:         "mixed comments and quotes",
			sql:          "/* comment */ SELECT 'a;'; -- comment\nSELECT 2;",
			wantCount:    2,
			wantContains: "SELECT 'a;'",
		},
		{
			name:      "basic semicolons no select",
			sql:       "foo; bar;",
			wantCount: 2,
			wantFirst: "foo;",
		},
		{
			name:         "block comments no select",
			sql:          "/* a;b */ foo;",
			wantCount:    1,
			wantContains: "foo",
		},
		{
			name:         "line comments no select",
			sql:          "-- note\nfoo;",
			wantCount:    1,
			wantContains: "foo",
		},
		{
			name:         "unterminated dollar quote no select",
			sql:          "x $$unclosed",
			wantCount:    1,
			wantContains: "$$unclosed",
		},
		{
			name:         "unterminated single quote no select",
			sql:          "x 'unclosed",
			wantCount:    1,
			wantContains: "'unclosed",
		},
		{
			name:      "multiple statements with whitespace no select",
			sql:       "  foo;   bar;  ",
			wantCount: 2,
			wantFirst: "foo;",
		},
		{
			name:      "nested parentheses no select",
			sql:       "x ((1)); foo;",
			wantCount: 2,
			wantFirst: "x ((1));",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SplitStatements(tt.sql)
			require.Equal(t, tt.wantCount, len(result))
			if tt.wantFirst != "" {
				assert.Equal(t, tt.wantFirst, result[0])
			}
			if tt.wantContains != "" {
				assert.Contains(t, result[0], tt.wantContains)
			}
		})
	}
}

func TestIntegration_SkipLineComment(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		i        int
		wantRes  string
		wantNewI int
	}{
		{
			name:     "basic line comment",
			sql:      "-- note\ndata",
			i:        0,
			wantRes:  "-- note\n",
			wantNewI: 8,
		},
		{
			name:     "line comment without newline",
			sql:      "-- note",
			i:        0,
			wantRes:  "-- note",
			wantNewI: 7,
		},
		{
			name:     "not at comment start",
			sql:      "data 1",
			i:        0,
			wantRes:  "",
			wantNewI: 0,
		},
		{
			name:     "single dash",
			sql:      "- more",
			i:        0,
			wantRes:  "",
			wantNewI: 0,
		},
		{
			name:     "empty string",
			sql:      "",
			i:        0,
			wantRes:  "",
			wantNewI: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, newI := SkipLineComment(tt.sql, tt.i)
			assert.Equal(t, tt.wantRes, result)
			assert.Equal(t, tt.wantNewI, newI)
		})
	}
}

func TestIntegration_SkipBlockComment(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		i        int
		wantRes  string
		wantNewI int
	}{
		{
			name:     "basic block comment",
			sql:      "/* note */ data",
			i:        0,
			wantRes:  "/* note */",
			wantNewI: 10,
		},
		{
			name:     "multiline block comment",
			sql:      "/* line1\nline2\nline3 */ data",
			i:        0,
			wantRes:  "/* line1\nline2\nline3 */",
			wantNewI: 23,
		},
		{
			name:     "not at comment start",
			sql:      "data 1",
			i:        0,
			wantRes:  "",
			wantNewI: 0,
		},
		{
			name:     "single slash",
			sql:      "/ more",
			i:        0,
			wantRes:  "",
			wantNewI: 0,
		},
		{
			name:     "unterminated block comment",
			sql:      "/* unclosed",
			i:        0,
			wantRes:  "/* unclosed",
			wantNewI: 11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, newI := SkipBlockComment(tt.sql, tt.i)
			assert.Equal(t, tt.wantRes, result)
			assert.Equal(t, tt.wantNewI, newI)
		})
	}
}

func TestIntegration_FindDollarQuoteEnd(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		i        int
		wantEndI int
		wantTag  string
	}{
		{
			name:     "anonymous dollar quote",
			sql:      "$$content$$ more",
			i:        0,
			wantEndI: 11,
			wantTag:  "$$",
		},
		{
			name:     "tagged dollar quote",
			sql:      "$tag$content$tag$ more",
			i:        0,
			wantEndI: 17,
			wantTag:  "$tag$",
		},
		{
			name:     "dollar quote in middle",
			sql:      "x $$content$$ from t",
			i:        2,
			wantEndI: 13,
			wantTag:  "$$",
		},
		{
			name:     "not at dollar sign",
			sql:      "data 1",
			i:        0,
			wantEndI: -1,
			wantTag:  "",
		},
		{
			name:     "empty string",
			sql:      "",
			i:        0,
			wantEndI: -1,
			wantTag:  "",
		},
		{
			name:     "unclosed dollar quote",
			sql:      "$$unclosed",
			i:        0,
			wantEndI: -1,
			wantTag:  "",
		},
		{
			name:     "tag with underscore",
			sql:      "$tag_123$content$tag_123$",
			i:        0,
			wantEndI: 25,
			wantTag:  "$tag_123$",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endI, tag := FindDollarQuoteEnd(tt.sql, tt.i)
			assert.Equal(t, tt.wantEndI, endI)
			assert.Equal(t, tt.wantTag, tag)
		})
	}
}

func TestIntegration_FindSingleQuoteEnd(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		i        int
		wantEndI int
	}{
		{
			name:     "basic single quote",
			sql:      "'text' more",
			i:        0,
			wantEndI: 6,
		},
		{
			name:     "quote in middle",
			sql:      "x 'hello'",
			i:        2,
			wantEndI: 9,
		},
		{
			name:     "escaped quote",
			sql:      "'a''b' more",
			i:        0,
			wantEndI: 6,
		},
		{
			name:     "multiple escaped quotes",
			sql:      "'a''b''c'",
			i:        0,
			wantEndI: 9,
		},
		{
			name:     "not at quote",
			sql:      "data 1",
			i:        0,
			wantEndI: -1,
		},
		{
			name:     "unclosed quote",
			sql:      "'unclosed",
			i:        0,
			wantEndI: -1,
		},
		{
			name:     "empty string",
			sql:      "",
			i:        0,
			wantEndI: -1,
		},
		{
			name:     "only quote",
			sql:      "'",
			i:        0,
			wantEndI: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endI := FindSingleQuoteEnd(tt.sql, tt.i)
			assert.Equal(t, tt.wantEndI, endI)
		})
	}
}

func TestIntegration_HandleLineComment(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		i       int
		wantOk  bool
		wantI   int
		wantStr string
	}{
		{
			name:    "line comment",
			sql:     "-- comment\nSELECT 1",
			i:       0,
			wantOk:  true,
			wantI:   11,
			wantStr: "-- comment\n",
		},
		{
			name:    "not a line comment",
			sql:     "SELECT 1",
			i:       0,
			wantOk:  false,
			wantI:   0,
			wantStr: "",
		},
		{
			name:    "single dash",
			sql:     "- more text",
			i:       0,
			wantOk:  false,
			wantI:   0,
			wantStr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			ok, newI := handleLineComment(tt.sql, tt.i, &sb)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantI, newI)
			assert.Equal(t, tt.wantStr, sb.String())
		})
	}
}

func TestIntegration_HandleBlockComment(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		i       int
		wantOk  bool
		wantI   int
		wantStr string
	}{
		{
			name:    "block comment",
			sql:     "/* comment */ SELECT 1",
			i:       0,
			wantOk:  true,
			wantI:   13,
			wantStr: "/* comment */",
		},
		{
			name:    "not a block comment",
			sql:     "SELECT 1",
			i:       0,
			wantOk:  false,
			wantI:   0,
			wantStr: "",
		},
		{
			name:    "single slash",
			sql:     "/ more text",
			i:       0,
			wantOk:  false,
			wantI:   0,
			wantStr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			ok, newI := handleBlockComment(tt.sql, tt.i, &sb)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantI, newI)
			assert.Equal(t, tt.wantStr, sb.String())
		})
	}
}

func TestIntegration_HandleDollarQuote(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		i       int
		wantOk  bool
		wantI   int
		wantStr string
	}{
		{
			name:    "dollar quote",
			sql:     "$$text;here$$ SELECT 1",
			i:       0,
			wantOk:  true,
			wantI:   13,
			wantStr: "$$text;here$$",
		},
		{
			name:    "tagged dollar quote",
			sql:     "$tag$content;$tag$ SELECT 1",
			i:       0,
			wantOk:  true,
			wantI:   18,
			wantStr: "$tag$content;$tag$",
		},
		{
			name:    "not a dollar quote",
			sql:     "SELECT 1",
			i:       0,
			wantOk:  false,
			wantI:   0,
			wantStr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			ok, newI := handleDollarQuote(tt.sql, tt.i, &sb)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantI, newI)
			assert.Equal(t, tt.wantStr, sb.String())
		})
	}
}

func TestIntegration_HandleSingleQuote(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		i       int
		n       int
		wantOk  bool
		wantI   int
		wantStr string
	}{
		{
			name:    "single quote",
			sql:     "'text;here'",
			i:       0,
			n:       11,
			wantOk:  true,
			wantI:   11,
			wantStr: "'text;here'",
		},
		{
			name:    "escaped quote",
			sql:     "'text''here' SELECT 1",
			i:       0,
			n:       20,
			wantOk:  true,
			wantI:   12,
			wantStr: "'text''here'",
		},
		{
			name:    "unterminated quote",
			sql:     "'unclosed",
			i:       0,
			n:       9,
			wantOk:  true,
			wantI:   9,
			wantStr: "'unclosed",
		},
		{
			name:    "not a single quote",
			sql:     "SELECT 1",
			i:       0,
			n:       9,
			wantOk:  false,
			wantI:   0,
			wantStr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			ok, newI := handleSingleQuote(tt.sql, tt.i, tt.n, &sb)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantI, newI)
			assert.Equal(t, tt.wantStr, sb.String())
		})
	}
}
