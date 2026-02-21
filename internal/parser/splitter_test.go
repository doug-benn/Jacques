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
