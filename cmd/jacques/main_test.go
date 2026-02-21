package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getProjectRoot() string {
	return filepath.Join("..", "..")
}

func TestCLI_FileInput(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.sql")
	err := os.WriteFile(inputFile, []byte("CREATE TABLE foo (id int);"), 0644)
	require.NoError(t, err)

	bin := filepath.Join(getProjectRoot(), "jacques.exe")
	if _, err := os.Stat(bin); err != nil {
		t.Skip("Binary not found, skipping CLI test")
	}

	cmd := exec.Command(bin, "-i", inputFile)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
	assert.Contains(t, string(output), "CREATE TABLE foo")
}

func TestCLI_MissingInputFile(t *testing.T) {
	bin := filepath.Join(getProjectRoot(), "jacques.exe")
	if _, err := os.Stat(bin); err != nil {
		t.Skip("Binary not found, skipping CLI test")
	}

	cmd := exec.Command(bin, "-i", "nonexistent.sql")
	output, err := cmd.CombinedOutput()
	assert.NotZero(t, err)
	assert.Contains(t, string(output), "Error")
}

func TestCLI_HelpFlag(t *testing.T) {
	bin := filepath.Join(getProjectRoot(), "jacques.exe")
	if _, err := os.Stat(bin); err != nil {
		t.Skip("Binary not found, skipping CLI test")
	}

	tests := []struct {
		name string
		args []string
	}{
		{"short flag", []string{"-h"}},
		{"long flag", []string{"--help"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(bin, tt.args...)
			output, err := cmd.CombinedOutput()
			assert.Zero(t, err, string(output))
			assert.Contains(t, string(output), "Input file")
			assert.Contains(t, string(output), "Output file")
		})
	}
}

func TestCLI_EmptyInput(t *testing.T) {
	// Note: This tests the "empty data" path, not the terminal detection path.
	// When cmd.Stdin = nil, the subprocess stdin becomes /dev/null which is NOT a
	// terminal, so the terminal check fails and it falls through to reading empty data.
	// Both paths result in the same error message, so this test still validates
	// that no input produces an error.
	bin := filepath.Join(getProjectRoot(), "jacques.exe")
	if _, err := os.Stat(bin); err != nil {
		t.Skip("Binary not found, skipping CLI test")
	}

	cmd := exec.Command(bin)
	cmd.Stdin = nil
	output, err := cmd.CombinedOutput()
	assert.NotZero(t, err)
	assert.Contains(t, string(output), "Error: no input specified")
	assert.Contains(t, string(output), "Input file")
	assert.Contains(t, string(output), "Output file")
}

func TestCLI_PipedInput(t *testing.T) {
	bin := filepath.Join(getProjectRoot(), "jacques.exe")
	if _, err := os.Stat(bin); err != nil {
		t.Skip("Binary not found, skipping CLI test")
	}

	input := "CREATE TABLE foo (id int);"
	cmd := exec.Command(bin)
	cmd.Stdin = strings.NewReader(input)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
	assert.Contains(t, string(output), "CREATE TABLE foo")
	assert.Contains(t, string(output), "id int")
}

func TestCLI_VersionFlag(t *testing.T) {
	bin := filepath.Join(getProjectRoot(), "jacques.exe")
	if _, err := os.Stat(bin); err != nil {
		t.Skip("Binary not found, skipping CLI test")
	}

	tests := []struct {
		name string
		args []string
	}{
		{"short flag", []string{"-v"}},
		{"long flag", []string{"--version"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(bin, tt.args...)
			output, err := cmd.CombinedOutput()
			assert.Zero(t, err, string(output))
			assert.Contains(t, string(output), "Jacques")
		})
	}
}

func TestCLI_OutputFlag(t *testing.T) {
	bin := filepath.Join(getProjectRoot(), "jacques.exe")
	if _, err := os.Stat(bin); err != nil {
		t.Skip("Binary not found, skipping CLI test")
	}

	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.sql")
	outputFile := filepath.Join(tmpDir, "output.sql")

	err := os.WriteFile(inputFile, []byte("CREATE TABLE bar (id int);"), 0644)
	require.NoError(t, err)

	cmd := exec.Command(bin, "-i", inputFile, "-o", outputFile)
	_, err = cmd.CombinedOutput()
	require.NoError(t, err)

	output, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Contains(t, string(output), "CREATE TABLE bar")
}

func TestCLI_EmptyFile(t *testing.T) {
	bin := filepath.Join(getProjectRoot(), "jacques.exe")
	if _, err := os.Stat(bin); err != nil {
		t.Skip("Binary not found, skipping CLI test")
	}

	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "empty.sql")
	err := os.WriteFile(emptyFile, []byte(""), 0644)
	require.NoError(t, err)

	cmd := exec.Command(bin, "-i", emptyFile)
	output, err := cmd.CombinedOutput()
	assert.NotZero(t, err)
	assert.Contains(t, string(output), "Error: no input specified")
}

func TestCLI_ExplicitStdin(t *testing.T) {
	bin := filepath.Join(getProjectRoot(), "jacques.exe")
	if _, err := os.Stat(bin); err != nil {
		t.Skip("Binary not found, skipping CLI test")
	}

	input := "CREATE TABLE baz (id int);"
	cmd := exec.Command(bin, "-i", "-")
	cmd.Stdin = strings.NewReader(input)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
	assert.Contains(t, string(output), "CREATE TABLE baz")
}

func TestCLI_CombinedFlags(t *testing.T) {
	bin := filepath.Join(getProjectRoot(), "jacques.exe")
	if _, err := os.Stat(bin); err != nil {
		t.Skip("Binary not found, skipping CLI test")
	}

	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.sql")
	outputFile := filepath.Join(tmpDir, "output.sql")

	err := os.WriteFile(inputFile, []byte("CREATE TABLE qux (id int); CREATE INDEX idx_qux ON qux(id);"), 0644)
	require.NoError(t, err)

	cmd := exec.Command(bin, "--input", inputFile, "--output", outputFile)
	_, err = cmd.CombinedOutput()
	require.NoError(t, err)

	output, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Contains(t, string(output), "CREATE TABLE qux")
	assert.Contains(t, string(output), "CREATE INDEX")
}
