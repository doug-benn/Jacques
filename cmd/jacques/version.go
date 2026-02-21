package main

import (
	"os/exec"
	"strings"
)

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
