package gitops

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"teriyake/go-git-it/config"
)

var (
	tokenPath = filepath.Join(os.Getenv("HOME"), ".go-git-it", ".token")
)

func IsGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	output, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(output)) == "true"
}

func HasToken() bool {
	if _, err := os.Stat(tokenPath); err != nil {
		return false
	}

	if _, err := ioutil.ReadFile(tokenPath); err != nil {
		return false
	}

	return true
}

func copyFile(src, dest string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("invalid source path: %v", err)
	}
	if srcInfo.IsDir() {
		return fmt.Errorf("source is a directory, not a file")
	}

	source, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("invalid source file: %v", err)
	}
	defer source.Close()

	dest = filepath.Join(dest, filepath.Base(src))
	destination, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func AddAndCommit(filename string, message string) error {
	profile, err := config.LoadUserProfile()
	if err != nil {
		return fmt.Errorf("failed to load user profile with %v", err)
	}
	repoPath := profile.GetCurrentRepo()

	err = copyFile(filename, repoPath)
	if err != nil {
		return fmt.Errorf("failed to copy task file to repo with %v", err)
	}

	taskPath := filepath.Join(repoPath, filepath.Base(filename))
	addCmd := exec.Command("git", "add", taskPath)
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("git add failed with %v", err)
	}
	commitCmd := exec.Command("git", "commit", "-m", message)
	return commitCmd.Run()
}

func CreateNewRepo(repoPath string) error {
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return fmt.Errorf("failed to create repository directory: %v", err)
	}

	parentDir := filepath.Dir(repoPath)
	isGitRepoCmd := exec.Command("git", "-C", parentDir, "rev-parse", "--is-inside-work-tree")
	if err := isGitRepoCmd.Run(); err == nil {
		initSubmoduleCmd := exec.Command("git", "-C", parentDir, "submodule", "add", "./"+filepath.Base(repoPath), filepath.Base(repoPath))
		if err := initSubmoduleCmd.Run(); err != nil {
			return fmt.Errorf("failed to initialize new submodule: %v", err)
		}
	} else {
		initCmd := exec.Command("git", "-C", repoPath, "init")
		if err := initCmd.Run(); err != nil {
			return fmt.Errorf("failed to initialize new repository: %v", err)
		}
	}

	return nil
}
