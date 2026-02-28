package processor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func processDefault(sql string) string {
	return Process(sql, nil)
}

func processExperimental(sql string) string {
	return Process(sql, &Options{ExperimentalFolding: true})
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
