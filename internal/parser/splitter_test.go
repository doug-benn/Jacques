package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitStatements(t *testing.T) {
	tests := []struct {
		name         string
		sql          string
		wantCount    int
		wantFirst    string
		wantContains string
	}{
		{
			name:      "basic semicolons",
			sql:       "foo; bar;",
			wantCount: 2,
			wantFirst: "foo;",
		},
		{
			name:      "single quotes",
			sql:       "x 'a;b';",
			wantCount: 1,
			wantFirst: "x 'a;b';",
		},
		{
			name:      "escaped quotes",
			sql:       "x 'a''b';",
			wantCount: 1,
			wantFirst: "x 'a''b';",
		},
		{
			name:      "dollar quotes",
			sql:       "x $$a;b$$;",
			wantCount: 1,
			wantFirst: "x $$a;b$$;",
		},
		{
			name:      "tagged dollar quotes",
			sql:       "$tag$text;here$tag$;",
			wantCount: 1,
			wantFirst: "$tag$text;here$tag$;",
		},
		{
			name:         "block comments",
			sql:          "/* a;b */ foo;",
			wantCount:    1,
			wantContains: "foo",
		},
		{
			name:         "line comments",
			sql:          "-- note\nfoo;",
			wantCount:    1,
			wantContains: "foo",
		},
		{
			name:         "unterminated dollar quote",
			sql:          "x $$unclosed",
			wantCount:    1,
			wantContains: "$$unclosed",
		},
		{
			name:         "unterminated single quote",
			sql:          "x 'unclosed",
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
			sql:       "  foo;   bar;  ",
			wantCount: 2,
			wantFirst: "foo;",
		},
		{
			name:      "nested parentheses in statement",
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

func TestSkipLineComment(t *testing.T) {
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

func TestSkipBlockComment(t *testing.T) {
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

func TestFindDollarQuoteEnd(t *testing.T) {
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

func TestFindSingleQuoteEnd(t *testing.T) {
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
