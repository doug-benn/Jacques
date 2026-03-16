package processor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Baseline values captured 2026-03-14
const (
	baselineLines = 41711
	baselineBytes = 2331618
)

// TestSizeReduction_GitLab reports the reduction in lines and bytes
// for the GitLab schema and fails if regression is detected.
func TestSizeReduction_GitLab(t *testing.T) {
	cwd, _ := os.Getwd()
	corpusPath := filepath.Join(cwd, "..", "..", "testdata", "corpus", "gitlabs_dump.sql")

	inputBytes, err := os.ReadFile(corpusPath)
	if err != nil {
		t.Skip("GitLab corpus not found")
	}
	inputSQL := string(inputBytes)

	cleanedSQL := Process(inputSQL, nil)

	origLines := len(strings.Split(inputSQL, "\n"))
	cleanLines := len(strings.Split(cleanedSQL, "\n"))

	origBytes := len(inputBytes)
	cleanBytes := len(cleanedSQL)

	fmt.Printf("\n--- Size Reduction Analysis: GitLab ---\n")
	fmt.Printf("Original: %d lines, %.2f MB\n", origLines, float64(origBytes)/(1024*1024))
	fmt.Printf("Cleaned:  %d lines, %.2f MB\n", cleanLines, float64(cleanBytes)/(1024*1024))
	fmt.Printf("Reduction: %d lines (%.1f%%), %.2f MB (%.1f%%)\n",
		origLines-cleanLines,
		100.0*float64(origLines-cleanLines)/float64(origLines),
		float64(origBytes-cleanBytes)/(1024*1024),
		100.0*float64(origBytes-cleanBytes)/float64(origBytes))
	fmt.Printf("Baseline:  %d lines, %.2f MB\n", baselineLines, float64(baselineBytes)/(1024*1024))
	fmt.Printf("---------------------------------------\n")

	// Fail if regression detected
	if cleanLines > baselineLines {
		t.Errorf("REGRESSION: Line count increased from %d to %d (delta: +%d)",
			baselineLines, cleanLines, cleanLines-baselineLines)
	}
	if cleanBytes > baselineBytes {
		t.Errorf("REGRESSION: Byte count increased from %d to %d (delta: +%d)",
			baselineBytes, cleanBytes, cleanBytes-baselineBytes)
	}
}

// BenchmarkProcess_GitLab measures the performance of Jacques on a large,
// complex production schema (GitLab).
func BenchmarkProcess_GitLab(b *testing.B) {
	cwd, _ := os.Getwd()
	corpusPath := filepath.Join(cwd, "..", "..", "testdata", "corpus", "gitlabs_dump.sql")

	inputBytes, err := os.ReadFile(corpusPath)
	if err != nil {
		b.Skipf("GitLab corpus not found at %s. Run curl first.", corpusPath)
	}
	inputSQL := string(inputBytes)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = Process(inputSQL, nil)
	}
}
