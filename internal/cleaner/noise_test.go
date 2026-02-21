package cleaner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsNoise(t *testing.T) {
	tests := []struct {
		name     string
		stmt     string
		expected bool
	}{
		{"SET statement", "SET statement_timeout = 0;", true},
		{"SELECT pg_catalog", "SELECT pg_catalog.set_config", true},
		{"GRANT", "GRANT SELECT ON foo TO bar;", true},
		{"REVOKE", "REVOKE ALL ON foo FROM bar;", true},
		{"COMMENT ON", "COMMENT ON TABLE foo IS 'test';", true},
		{"ALTER OWNER TO", "ALTER TABLE foo OWNER TO bar;", true},
		{"ALTER TABLE ONLY OWNER TO", "ALTER TABLE ONLY foo OWNER TO bar;", false},
		{"Valid CREATE TABLE", "CREATE TABLE foo (id int);", false},
		{"ALTER SEQUENCE", "ALTER SEQUENCE foo OWNED BY bar;", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNoise(tt.stmt)
			assert.Equal(t, tt.expected, result)
		})
	}
}
