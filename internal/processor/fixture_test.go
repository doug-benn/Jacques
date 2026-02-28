//go:build !short

package processor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type FixtureInfo struct {
	Name         string
	Dir          string
	InputPath    string
	ExpectedPath string
}

func NormalizeSQL(sql string) string {
	sql = strings.ReplaceAll(sql, "\r\n", "\n")
	return strings.TrimSpace(sql)
}

func DiscoverFixtures(dir string) []FixtureInfo {
	baseDir := filepath.Join("..", "..", dir)
	inputPattern := filepath.Join(baseDir, "*_input.sql")

	matches, err := filepath.Glob(inputPattern)
	if err != nil || len(matches) == 0 {
		return nil
	}

	var fixtures []FixtureInfo
	for _, inputPath := range matches {
		name := filepath.Base(inputPath)
		expectedName := strings.Replace(name, "_input.sql", "_expected.sql", 1)
		expectedPath := filepath.Join(baseDir, expectedName)

		fixtures = append(fixtures, FixtureInfo{
			Name:         name,
			Dir:          dir,
			InputPath:    inputPath,
			ExpectedPath: expectedPath,
		})
	}

	return fixtures
}

func LoadFixture(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join("..", "..", dir, name)
	data, err := os.ReadFile(path)
	require.NoError(t, err, "failed to read fixture: %s", path)
	return string(data)
}

func TestFixtures(t *testing.T) {
	tests := []struct {
		dir       string
		processor func(string) string
	}{
		{"testdata/e2e/", processDefault},
		{"testdata/fixtures/", processExperimental},
	}

	for _, tc := range tests {
		t.Run(tc.dir, func(t *testing.T) {
			fixtures := DiscoverFixtures(tc.dir)
			require.NotEmpty(t, fixtures, "no fixtures found in %s", tc.dir)

			for _, f := range fixtures {
				t.Run(f.Name, func(t *testing.T) {
					input := LoadFixture(t, f.Dir, f.Name)
					expected := LoadFixture(t, f.Dir, strings.Replace(f.Name, "_input.sql", "_expected.sql", 1))

					result := tc.processor(input)
					assert.Equal(t, NormalizeSQL(expected), NormalizeSQL(result),
						"Cleaned output should match expected file")
				})
			}
		})
	}
}
