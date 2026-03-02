package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindDollarQuoteEnd_NonSQL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		i        int
		wantEndI int
		wantTag  string
	}{
		{
			name:     "empty string",
			input:    "",
			i:        0,
			wantEndI: -1,
			wantTag:  "",
		},
		{
			name:     "not at dollar sign",
			input:    "data 1",
			i:        0,
			wantEndI: -1,
			wantTag:  "",
		},
		{
			name:     "dollar in middle not at start",
			input:    "text $tag$ content",
			i:        5,
			wantEndI: -1,
			wantTag:  "",
		},
		{
			name:     "single dollar not followed by tag",
			input:    "$ not a quote",
			i:        0,
			wantEndI: -1,
			wantTag:  "",
		},
		{
			name:     "unclosed dollar quote",
			input:    "$$unclosed",
			i:        0,
			wantEndI: -1,
			wantTag:  "",
		},
		{
			name:     "numeric after dollar",
			input:    "$123$content$123$",
			i:        0,
			wantEndI: -1,
			wantTag:  "",
		},
		{
			name:     "invalid tag starts with number",
			input:    "$1tag$content$1tag$",
			i:        0,
			wantEndI: -1,
			wantTag:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endI, tag := FindDollarQuoteEnd(tt.input, tt.i)
			assert.Equal(t, tt.wantEndI, endI)
			assert.Equal(t, tt.wantTag, tag)
		})
	}
}

func TestFindSingleQuoteEnd_NonSQL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		i        int
		wantEndI int
	}{
		{
			name:     "empty string",
			input:    "",
			i:        0,
			wantEndI: -1,
		},
		{
			name:     "not at quote",
			input:    "data 1",
			i:        0,
			wantEndI: -1,
		},
		{
			name:     "quote in middle not at start",
			input:    "text hello",
			i:        5,
			wantEndI: -1,
		},
		{
			name:     "double quote not single",
			input:    "\"text\"",
			i:        0,
			wantEndI: -1,
		},
		{
			name:     "backtick not single quote",
			input:    "`text`",
			i:        0,
			wantEndI: -1,
		},
		{
			name:     "only quote",
			input:    "'",
			i:        0,
			wantEndI: -1,
		},
		{
			name:     "two quotes not escaped",
			input:    "''",
			i:        0,
			wantEndI: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endI := FindSingleQuoteEnd(tt.input, tt.i)
			assert.Equal(t, tt.wantEndI, endI)
		})
	}
}

func TestSkipLineComment_NonSQL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		i        int
		wantRes  string
		wantNewI int
	}{
		{
			name:     "empty string",
			input:    "",
			i:        0,
			wantRes:  "",
			wantNewI: 0,
		},
		{
			name:     "not at comment start",
			input:    "data 1",
			i:        0,
			wantRes:  "",
			wantNewI: 0,
		},
		{
			name:     "single dash",
			input:    "- more",
			i:        0,
			wantRes:  "",
			wantNewI: 0,
		},
		{
			name:     "dash not followed by dash",
			input:    "-- not a comment at position 1",
			i:        1,
			wantRes:  "",
			wantNewI: 1,
		},
		{
			name:     "starts with other character",
			input:    "# not a comment",
			i:        0,
			wantRes:  "",
			wantNewI: 0,
		},
		{
			name:     "slash not dash",
			input:    "/ comment",
			i:        0,
			wantRes:  "",
			wantNewI: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, newI := SkipLineComment(tt.input, tt.i)
			assert.Equal(t, tt.wantRes, result)
			assert.Equal(t, tt.wantNewI, newI)
		})
	}
}

func TestSkipBlockComment_NonSQL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		i        int
		wantRes  string
		wantNewI int
	}{
		{
			name:     "empty string",
			input:    "",
			i:        0,
			wantRes:  "",
			wantNewI: 0,
		},
		{
			name:     "not at comment start",
			input:    "data 1",
			i:        0,
			wantRes:  "",
			wantNewI: 0,
		},
		{
			name:     "single slash",
			input:    "/ more",
			i:        0,
			wantRes:  "",
			wantNewI: 0,
		},
		{
			name:     "slash not followed by asterisk",
			input:    "/* not a comment at position 1",
			i:        1,
			wantRes:  "",
			wantNewI: 1,
		},
		{
			name:     "starts with asterisk",
			input:    "* not a comment",
			i:        0,
			wantRes:  "",
			wantNewI: 0,
		},
		{
			name:     "dash not slash",
			input:    "-- comment",
			i:        0,
			wantRes:  "",
			wantNewI: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, newI := SkipBlockComment(tt.input, tt.i)
			assert.Equal(t, tt.wantRes, result)
			assert.Equal(t, tt.wantNewI, newI)
		})
	}
}

func TestHandleLineComment_NonSQL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		i       int
		wantOk  bool
		wantI   int
		wantStr string
	}{
		{
			name:    "empty string",
			input:   "",
			i:       0,
			wantOk:  false,
			wantI:   0,
			wantStr: "",
		},
		{
			name:    "not a line comment",
			input:   "data 1",
			i:       0,
			wantOk:  false,
			wantI:   0,
			wantStr: "",
		},
		{
			name:    "single dash",
			input:   "- more text",
			i:       0,
			wantOk:  false,
			wantI:   0,
			wantStr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			ok, newI := handleLineComment(tt.input, tt.i, &sb)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantI, newI)
			assert.Equal(t, tt.wantStr, sb.String())
		})
	}
}

func TestHandleBlockComment_NonSQL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		i       int
		wantOk  bool
		wantI   int
		wantStr string
	}{
		{
			name:    "empty string",
			input:   "",
			i:       0,
			wantOk:  false,
			wantI:   0,
			wantStr: "",
		},
		{
			name:    "not a block comment",
			input:   "data 1",
			i:       0,
			wantOk:  false,
			wantI:   0,
			wantStr: "",
		},
		{
			name:    "single slash",
			input:   "/ more text",
			i:       0,
			wantOk:  false,
			wantI:   0,
			wantStr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			ok, newI := handleBlockComment(tt.input, tt.i, &sb)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantI, newI)
			assert.Equal(t, tt.wantStr, sb.String())
		})
	}
}

func TestHandleDollarQuote_NonSQL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		i       int
		wantOk  bool
		wantI   int
		wantStr string
	}{
		{
			name:    "empty string",
			input:   "",
			i:       0,
			wantOk:  false,
			wantI:   0,
			wantStr: "",
		},
		{
			name:    "not a dollar quote",
			input:   "data 1",
			i:       0,
			wantOk:  false,
			wantI:   0,
			wantStr: "",
		},
		{
			name:    "dollar in middle not at position",
			input:   "text $tag$ content",
			i:       5,
			wantOk:  false,
			wantI:   5,
			wantStr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			ok, newI := handleDollarQuote(tt.input, tt.i, &sb)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantI, newI)
			assert.Equal(t, tt.wantStr, sb.String())
		})
	}
}

func TestHandleSingleQuote_NonSQL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		i       int
		n       int
		wantOk  bool
		wantI   int
		wantStr string
	}{
		{
			name:    "empty string",
			input:   "",
			i:       0,
			n:       0,
			wantOk:  false,
			wantI:   0,
			wantStr: "",
		},
		{
			name:    "not a single quote",
			input:   "data 1",
			i:       0,
			n:       9,
			wantOk:  false,
			wantI:   0,
			wantStr: "",
		},
		{
			name:    "quote in middle not at position",
			input:   "text hello",
			i:       5,
			n:       10,
			wantOk:  false,
			wantI:   5,
			wantStr: "",
		},
		{
			name:    "double quote not single",
			input:   "\"text\"",
			i:       0,
			n:       7,
			wantOk:  false,
			wantI:   0,
			wantStr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			ok, newI := handleSingleQuote(tt.input, tt.i, tt.n, &sb)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantI, newI)
			assert.Equal(t, tt.wantStr, sb.String())
		})
	}
}
