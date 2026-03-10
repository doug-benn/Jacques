package cleaner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFirstLine(t *testing.T) {
	tests := []struct {
		name string
		stmt string
		want string
	}{
		{"single line", "x 1", "x 1"},
		{"two lines", "x 1\ny 2", "x 1"},
		{"with trailing newline", "x 1\n", "x 1"},
		{"empty string", "", ""},
		{"newline only", "\n", ""},
		{"multiline with spaces", "  x 1  \n  y 2  ", "  x 1  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFirstLine(tt.stmt)
			assert.Equal(t, tt.want, result)
		})
	}
}
