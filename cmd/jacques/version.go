package main

import (
	"os/exec"
	"strings"
)

// Version is set at build time by goreleaser via ldflags (-X main.Version={{.Version}}).
// For local development builds, it falls back to git describe output.
var Version = versionFromGit()

func versionFromGit() string {
	cmd := exec.Command("git", "describe", "--tags", "--always", "--dirty")
	cmd.Dir = "."
	output, err := cmd.Output()
	if err != nil {
		return "dev"
	}
	return strings.TrimSpace(string(output))
}
