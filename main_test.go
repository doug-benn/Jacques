package main

import (
	"bytes"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFlags_Defaults(t *testing.T) {
	input, output, version, experimentalFolding, err := parseFlags([]string{})
	require.NoError(t, err)
	assert.Equal(t, "-", input)
	assert.Equal(t, "-", output)
	assert.False(t, version)
	assert.False(t, experimentalFolding)
}

func TestParseFlags_ShortFlags(t *testing.T) {
	input, output, version, experimentalFolding, err := parseFlags([]string{"-i", "in.sql", "-o", "out.sql", "-v", "-experimental-folding"})
	require.NoError(t, err)
	assert.Equal(t, "in.sql", input)
	assert.Equal(t, "out.sql", output)
	assert.True(t, version)
	assert.True(t, experimentalFolding)
}

func TestParseFlags_LongFlags(t *testing.T) {
	input, output, version, experimentalFolding, err := parseFlags([]string{"--input", "in.sql", "--output", "out.sql", "--version", "--experimental-folding"})
	require.NoError(t, err)
	assert.Equal(t, "in.sql", input)
	assert.Equal(t, "out.sql", output)
	assert.True(t, version)
	assert.True(t, experimentalFolding)
}

func TestParseFlags_UnknownFlag(t *testing.T) {
	_, _, _, _, err := parseFlags([]string{"--unknown"})
	assert.Error(t, err)
}

func TestCLI_HelpFlag(t *testing.T) {
	_, _, _, _, err := parseFlags([]string{"-h"})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, flag.ErrHelp))
}

func TestCLI_VersionFlag(t *testing.T) {
	_, _, version, _, err := parseFlags([]string{"-v"})
	require.NoError(t, err)
	assert.True(t, version)
}

func TestCLI_FileInput(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.sql")
	err := os.WriteFile(inputFile, []byte("CREATE TABLE foo (id int);"), 0644)
	require.NoError(t, err)

	var output bytes.Buffer
	err = run(inputFile, "-", false, nil, &output)
	require.NoError(t, err)
	assert.Contains(t, output.String(), "CREATE TABLE foo")
}

func TestCLI_MissingInputFile(t *testing.T) {
	var output bytes.Buffer
	err := run("nonexistent.sql", "-", false, nil, &output)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent.sql")
}

func TestCLI_EmptyInput(t *testing.T) {
	var output bytes.Buffer
	err := run("-", "-", false, strings.NewReader(""), &output)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no input specified")
}

func TestCLI_PipedInput(t *testing.T) {
	input := "CREATE TABLE foo (id int);"
	var output bytes.Buffer
	err := run("-", "-", false, strings.NewReader(input), &output)
	require.NoError(t, err)
	assert.Contains(t, output.String(), "CREATE TABLE foo")
	assert.Contains(t, output.String(), "id int")
}

func TestCLI_OutputFlag(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.sql")
	outputFile := filepath.Join(tmpDir, "output.sql")

	err := os.WriteFile(inputFile, []byte("CREATE TABLE bar (id int);"), 0644)
	require.NoError(t, err)

	err = run(inputFile, outputFile, false, nil, nil)
	require.NoError(t, err)

	output, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Contains(t, string(output), "CREATE TABLE bar")
}

func TestCLI_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "empty.sql")
	err := os.WriteFile(emptyFile, []byte(""), 0644)
	require.NoError(t, err)

	var output bytes.Buffer
	err = run(emptyFile, "-", false, nil, &output)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no input specified")
}

func TestCLI_ExplicitStdin(t *testing.T) {
	input := "CREATE TABLE baz (id int);"
	var output bytes.Buffer
	err := run("-", "-", false, strings.NewReader(input), &output)
	require.NoError(t, err)
	assert.Contains(t, output.String(), "CREATE TABLE baz")
}

func TestCLI_CombinedFlags(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.sql")
	outputFile := filepath.Join(tmpDir, "output.sql")

	err := os.WriteFile(inputFile, []byte("CREATE TABLE qux (id int); CREATE INDEX idx_qux ON qux(id);"), 0644)
	require.NoError(t, err)

	err = run(inputFile, outputFile, false, nil, nil)
	require.NoError(t, err)

	output, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Contains(t, string(output), "CREATE TABLE qux")
	assert.Contains(t, string(output), "CREATE INDEX")
}

func TestCLI_ExperimentalFoldingFlag_Default(t *testing.T) {
	input := "CREATE TABLE public.users (id int, email public.email); CREATE DOMAIN public.email AS text;"
	var output bytes.Buffer
	err := run("-", "-", false, strings.NewReader(input), &output)
	require.NoError(t, err)
	assert.NotContains(t, output.String(), "CREATE DOMAIN")
	assert.Contains(t, output.String(), "email public.email")
}

func TestCLI_ExperimentalFoldingFlag_Enabled(t *testing.T) {
	input := "CREATE TABLE public.users (id int, email public.email); CREATE DOMAIN public.email AS text;"
	var output bytes.Buffer
	err := run("-", "-", true, strings.NewReader(input), &output)
	require.NoError(t, err)
	assert.Contains(t, output.String(), "CREATE DOMAIN")
}

func TestCLI_PartitionChildren_Default(t *testing.T) {
	input := "CREATE TABLE public.orders (id int) PARTITION BY RANGE (id); CREATE TABLE public.orders_2024 PARTITION OF public.orders FOR VALUES FROM (MINVALUE) TO (2025);"
	var output bytes.Buffer
	err := run("-", "-", false, strings.NewReader(input), &output)
	require.NoError(t, err)
	assert.NotContains(t, output.String(), "PARTITION OF")
	assert.Contains(t, output.String(), "PARTITION BY RANGE")
}

func TestCLI_PartitionChildren_ExperimentalFolding(t *testing.T) {
	input := "CREATE TABLE public.orders (id int) PARTITION BY RANGE (id); CREATE TABLE public.orders_2024 PARTITION OF public.orders FOR VALUES FROM (MINVALUE) TO (2025);"
	var output bytes.Buffer
	err := run("-", "-", true, strings.NewReader(input), &output)
	require.NoError(t, err)
	assert.Contains(t, output.String(), "PARTITION OF")
	assert.Contains(t, output.String(), "PARTITION BY RANGE")
}
