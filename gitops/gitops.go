package gitops

import (
	"os/exec"
	"os"
	"fmt"
	"strings"
)

func IsGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	output, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(output)) == "true"
}

func AddAndCommit(filename string, message string) error {
	addCmd := exec.Command("git", "add", filename)
	if err := addCmd.Run(); err != nil {
		return err
	}
	commitCmd := exec.Command("git", "commit", "-m", message)
	return commitCmd.Run()
}

func CreateNewRepo(repoPath string) error {
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return fmt.Errorf("failed to create repository directory: %v", err)
	}

	cmd := exec.Command("git", "-C", repoPath, "init")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to initialize new repository: %v", err)
	}

	return nil
}
