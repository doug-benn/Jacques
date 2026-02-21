package main

import (
	"os"
	"os/exec"
	"path/filepath"
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
	bin := filepath.Join(getProjectRoot(), "jacques.exe")
	if _, err := os.Stat(bin); err != nil {
		t.Skip("Binary not found, skipping CLI test")
	}

	cmd := exec.Command(bin)
	cmd.Stdin = nil
	output, err := cmd.CombinedOutput()
	assert.Zero(t, err, string(output))
	assert.Equal(t, "\n", string(output))
}
