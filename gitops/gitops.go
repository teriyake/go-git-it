package gitops

import (
	"os/exec"
	"os"
	"fmt"
	"strings"
	"path/filepath"
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

    // Check if the parent directory is a git repository
    parentDir := filepath.Dir(repoPath)
    isGitRepoCmd := exec.Command("git", "-C", parentDir, "rev-parse", "--is-inside-work-tree")
    if err := isGitRepoCmd.Run(); err == nil { // Parent directory is a git repo
        // Initialize as a submodule
        initSubmoduleCmd := exec.Command("git", "-C", parentDir, "submodule", "add", "./"+filepath.Base(repoPath), filepath.Base(repoPath))
        if err := initSubmoduleCmd.Run(); err != nil {
            return fmt.Errorf("failed to initialize new submodule: %v", err)
        }
    } else {
        // Initialize as a new repository
        initCmd := exec.Command("git", "-C", repoPath, "init")
        if err := initCmd.Run(); err != nil {
            return fmt.Errorf("failed to initialize new repository: %v", err)
        }
    }

    return nil
}

