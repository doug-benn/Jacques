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
		{"SET statement", "SET foo_bar = baz;", true},
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

func TestIsSET(t *testing.T) {
	tests := []struct {
		name     string
		stmt     string
		expected bool
	}{
		{"basic SET", "SET foo = bar;", true},
		{"SET with whitespace", "   SET  foo  =  bar  ;", true},
		{"lowercase set", "set foo = bar;", false},
		{"not SET", "x 1", false},
		{"SET in middle", "x 1; SET foo = bar;", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSET(tt.stmt)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsSelectPgCatalog(t *testing.T) {
	tests := []struct {
		name     string
		stmt     string
		expected bool
	}{
		{"SELECT pg_catalog", "SELECT pg_catalog.x", true},
		{"SELECT pg_catalog uppercase", "SELECT X.Y", false},
		{"pg_catalog in middle", "SELECT 1 FROM x.y", false},
		{"not pg_catalog", "x 1", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSelectPgCatalog(tt.stmt)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsGRANT(t *testing.T) {
	tests := []struct {
		name     string
		stmt     string
		expected bool
	}{
		{"basic GRANT", "GRANT SELECT ON foo TO bar;", true},
		{"GRANT ALL", "GRANT ALL ON foo TO bar;", true},
		{"lowercase grant", "grant select on foo to bar;", false},
		{"not GRANT", "x 1", false},
		{"GRANT in comment", "-- GRANT SELECT ON foo", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsGRANT(tt.stmt)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsREVOKE(t *testing.T) {
	tests := []struct {
		name     string
		stmt     string
		expected bool
	}{
		{"basic REVOKE", "REVOKE ALL ON foo FROM bar;", true},
		{"REVOKE SELECT", "REVOKE SELECT ON foo FROM bar;", true},
		{"lowercase revoke", "revoke all on foo from bar;", false},
		{"not REVOKE", "x 1", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsREVOKE(tt.stmt)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsCOMMENT(t *testing.T) {
	tests := []struct {
		name     string
		stmt     string
		expected bool
	}{
		{"COMMENT ON TABLE", "COMMENT ON TABLE foo IS 'test';", true},
		{"COMMENT ON COLUMN", "COMMENT ON COLUMN foo.bar IS 'test';", true},
		{"lowercase comment", "comment on table foo is 'test';", false},
		{"not COMMENT", "x 1", false},
		{"COMMENT without ON", "COMMENT 'test';", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCOMMENT(tt.stmt)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsALTEROwner(t *testing.T) {
	tests := []struct {
		name     string
		stmt     string
		expected bool
	}{
		{"ALTER TABLE OWNER TO", "ALTER TABLE foo OWNER TO bar;", true},
		{"ALTER SCHEMA OWNER TO", "ALTER SCHEMA foo OWNER TO bar;", true},
		{"ALTER VIEW OWNER TO", "ALTER VIEW foo OWNER TO bar;", true},
		{"ALTER TABLE ONLY OWNER TO", "ALTER TABLE ONLY foo OWNER TO bar;", false},
		{"ALTER SEQUENCE OWNED BY", "ALTER SEQUENCE foo OWNED BY bar;", false},
		{"lowercase alter", "alter table foo owner to bar;", false},
		{"not ALTER", "x 1", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsALTEROwner(tt.stmt)
			assert.Equal(t, tt.expected, result)
		})
	}
}
