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
	// E2E fixtures: run with default mode only
	t.Run("testdata/e2e/", func(t *testing.T) {
		fixtures := DiscoverFixtures("testdata/e2e/")
		require.NotEmpty(t, fixtures, "no fixtures found in testdata/e2e/")

		for _, f := range fixtures {
			t.Run(f.Name, func(t *testing.T) {
				input := LoadFixture(t, f.Dir, f.Name)
				expected := LoadFixture(t, f.Dir, strings.Replace(f.Name, "_input.sql", "_expected.sql", 1))

				result := processDefault(input)
				assert.Equal(t, NormalizeSQL(expected), NormalizeSQL(result),
					"Cleaned output should match expected file")
			})
		}
	})

	// Gated fixtures: run with BOTH default and experimental modes
	// This ensures ExperimentalFolding doesn't break default behavior
	t.Run("testdata/fixtures/", func(t *testing.T) {
		fixtures := DiscoverFixtures("testdata/fixtures/")
		require.NotEmpty(t, fixtures, "no fixtures found in testdata/fixtures/")

		for _, f := range fixtures {
			t.Run(f.Name+"/default", func(t *testing.T) {
				input := LoadFixture(t, f.Dir, f.Name)
				expected := LoadFixture(t, f.Dir, strings.Replace(f.Name, "_input.sql", "_expected.sql", 1))

				result := processDefault(input)
				assert.Equal(t, NormalizeSQL(expected), NormalizeSQL(result),
					"Default mode output should match expected file")
			})

			t.Run(f.Name+"/experimental", func(t *testing.T) {
				input := LoadFixture(t, f.Dir, f.Name)
				expected := LoadFixture(t, f.Dir, strings.Replace(f.Name, "_input.sql", "_experimental_expected.sql", 1))

				result := processExperimental(input)
				assert.Equal(t, NormalizeSQL(expected), NormalizeSQL(result),
					"Experimental mode output should match experimental expected file")
			})
		}
	})
}
