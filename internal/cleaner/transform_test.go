package cleaner

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
