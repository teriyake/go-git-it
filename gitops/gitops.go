package gitops

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"teriyake/go-git-it/config"
)

const baseUrl = "https://api.github.com"

var (
	tokenPath     = filepath.Join(os.Getenv("HOME"), ".go-git-it", ".token")
	localReposDir = filepath.Join(os.Getenv("HOME"), ".go-git-it", "repos")
)

type Repo struct {
	Owner         string   `json:"owner"`
	ID            int64    `json:"id"`
	NodeID        string   `json:"nodeId"`
	Name          string   `json:"name"`
	FullName      string   `json:"fullName"`
	Description   string   `json:"description"`
	Private       bool     `json:"private"`
	LicenseInfo   *License `json:"license"`
	DefaultBranch string   `json:"defaultBranch"`
}

type License struct {
	Key    string `json:"key"`
	Name   string `json:"name"`
	SpdxID string `json:"spdxId"`
	Url    string `json:"url"`
}

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

	repoName := profile.GetCurrentRepo()
	repoPath := filepath.Join(os.Getenv("HOME"), ".go-git-it", "repos", repoName)

	err = copyFile(filename, repoPath)
	if err != nil {
		return fmt.Errorf("failed to copy task file to repo with %v", err)
	}

	_, fileName := filepath.Split(filename)
	addCmd := exec.Command("git", "-C", repoPath, "add", fileName)
	if out, err := addCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add %s failed with %v and output: %v", fileName, err, string(out))
	}
	commitCmd := exec.Command("git", "-C", repoPath, "commit", "-m", message)
	if out, err := commitCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit failed with %v and output: %v", err, string(out))
	}
	pushCmd := exec.Command("git", "-C", repoPath, "push")
	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("git push failed with %v", err)
	}

	return nil
}

func CreateNewRepo(repoName string, privacy bool) error {
	profile, err := config.LoadUserProfile()
	if err != nil {
		return fmt.Errorf("failed to load user profile with %v", err)
	}
	username := profile.GetUsername()

	token, e := config.GetToken()
	if e != nil {
		return fmt.Errorf("failed to get auth token with %v", e)
	}

	urlStr := strings.Join([]string{baseUrl, "user", "repos"}, "/")

	reqBody := map[string]interface{}{
		"name":                repoName,
		"description":         "a to-do repo generated with go-git-it (ggi): https://github.com/teriyake/go-git-it",
		"homepage":            "",
		"auto_init":           true,
		"readme":              "default",
		"visibility":          "private",
		"hasIssues":           true,
		"hasProjects":         true,
		"hasWiki":             true,
		"isTemplate":          false,
		"allowSquashMerge":    true,
		"allowRebaseMerge":    false,
		"deleteBranchOnMerge": false,
		"licenseTemplate":     "MIT",
	}

	bodyBytes, _ := json.Marshal(reqBody)

	client := &http.Client{}
	request, err := http.NewRequestWithContext(context.Background(), "POST", urlStr, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create http request with %v", err)
	}

	request.Header.Set("Authorization", "token "+token)
	response, err := client.Do(request)
	defer response.Body.Close()

	if err != nil || response.StatusCode >= 400 {
		data, _ := ioutil.ReadAll(response.Body)
		return fmt.Errorf("http request failed with %v\nresponse received: %v\n", err, string(data))
	}

	respData := make(map[string]interface{}, 0)
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&respData)
	if err != nil {
		return fmt.Errorf("failed to decode json response with %v", err)
	}

	targetDir := filepath.Join(os.Getenv("HOME"), ".go-git-it", "repos", repoName)

	if _, err := os.Stat(targetDir); !os.IsNotExist(err) {
		return fmt.Errorf("target directory %s already exists", targetDir)
	}

	if err := os.MkdirAll(filepath.Dir(targetDir), 0755); err != nil {
		return fmt.Errorf("unable to create parent directories for %s: %w", targetDir, err)
	}

	remoteURL := fmt.Sprintf("https://github.com/%s/%s.git", username, repoName)
	cmd := exec.Command("git", "clone", remoteURL, targetDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed with %v", err)
	}

	fmt.Printf("Repository cloned successfully into %s\n", targetDir)

	return nil
}
