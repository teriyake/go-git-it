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
	"time"
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

type Issue struct {
	Number    int      `json:"number"`
	State     string   `json:"state"`
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Assignees []string `json:"assignees,omitempty"`
	Labels    []*Label `json:"labels,omitempty"`
}

type Label struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

func NewLabel(status string) *Label {
	if status == "done" {
		return &Label{Name: status, Description: "Mark a task as done", Color: "#98971a"}
	}
	if status == "doing" {
		return &Label{Name: status, Description: "Mark a task as in-progress", Color: "#d79921"}
	}
	if status == "will-do" {
		return &Label{Name: status, Description: "Mark a task as not-yet-started", Color: "#7c6f64"}
	}
	return nil
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
	var vis string
	if privacy {
		vis = "private"
	} else {
		vis = "public"
	}

	reqBody := map[string]interface{}{
		"name":                repoName,
		"description":         "a to-do repo generated with go-git-it (ggi): https://github.com/teriyake/go-git-it",
		"homepage":            "",
		"auto_init":           true,
		"readme":              "default",
		"visibility":          vis,
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

func DeleteRemoteRepo(repoName string) error {
	profile, err := config.LoadUserProfile()
	if err != nil {
		return fmt.Errorf("failed to load user profile with %v", err)
	}
	username := profile.Username

	token, err := config.GetToken()
	if err != nil {
		return fmt.Errorf("failed to get auth token with %v", err)
	}

	urlStr := fmt.Sprintf("https://api.github.com/repos/%s/%s", username, repoName)

	client := &http.Client{}
	request, err := http.NewRequestWithContext(context.Background(), "DELETE", urlStr, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request with %v", err)
	}

	request.Header.Set("Authorization", "token "+token)
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("HTTP request failed with %v\nresponse received:\n%v\n", err, response)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		data, _ := ioutil.ReadAll(response.Body)
		return fmt.Errorf("GitHub API responded with status code %d: %s", response.StatusCode, string(data))
	}

	return nil
}

func CreateMilestone(token, owner, repo, title, dueDate string) (int, error) {

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/milestones", owner, repo)

	requestBody, err := json.Marshal(map[string]string{
		"title":  title,
		"due_on": dueDate,
	})
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := ioutil.ReadAll(resp.Body)
		return 0, fmt.Errorf("failed to create milestone: %s", string(body))
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if idFloat, ok := result["number"].(float64); ok {
		return int(idFloat), nil
	}

	return 0, fmt.Errorf("could not parse milestone ID")
}

func CreateIssueWithMilestone(token, username, repoName, issueTitle, issueBody string, milestoneNumber int) error {
	urlStr := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", username, repoName)
	issueData := map[string]interface{}{
		"title":     issueTitle,
		"body":      issueBody,
		"milestone": milestoneNumber,
	}

	data, _ := json.Marshal(issueData)

	request, err := http.NewRequestWithContext(context.Background(), "POST", urlStr, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request with %v", err)
	}

	request.Header.Set("Authorization", "token "+token)
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("HTTP request failed with %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(response.Body)
		return fmt.Errorf("GitHub API responded with status code %d: %s", response.StatusCode, string(body))
	}

	return nil
}

func SetDeadline(taskDescription, deadlineStr string) error {
	profile, err := config.LoadUserProfile()
	if err != nil {
		return fmt.Errorf("failed to load user profile with %v", err)
	}

	token, err := config.GetToken()
	if err != nil {
		return fmt.Errorf("failed to get GitHub token with %v", err)
	}

	username := profile.GetUsername()
	repoName := profile.GetCurrentRepo()

	parsedDeadline, err := time.Parse("2006-01-02", deadlineStr)
	if err != nil {
		return fmt.Errorf("invalid deadline format: %v", err)
	}
	deadline := fmt.Sprintf("%sT00:00:00Z", parsedDeadline.Format("2006-01-02"))

	milestoneID, err := CreateMilestone(token, username, repoName, taskDescription, deadline)
	if err != nil {
		return fmt.Errorf("failed to create milestone with %v", err)
	}

	issueBody := fmt.Sprintf("This task is due on %s", deadline)
	if err := CreateIssueWithMilestone(token, username, repoName, taskDescription, issueBody, milestoneID); err != nil {
		return fmt.Errorf("failed to create issue with milestone with %v", err)
	}

	fmt.Printf("Milestone %v successfully set!\n", milestoneID)
	return nil
}

func ChangeIssueLabel(repoName string, issueNumber int, labels []string) error {
	profile, err := config.LoadUserProfile()
	if err != nil {
		return fmt.Errorf("failed to load user profile with %v", err)
	}

	token, err := config.GetToken()
	if err != nil {
		return fmt.Errorf("failed to get auth token with %v", err)
	}

	urlStr := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/labels", profile.Username, repoName, issueNumber)
	requestBody := map[string][]string{
		"labels": labels,
	}
	data, _ := json.Marshal(requestBody)

	client := &http.Client{}
	request, err := http.NewRequestWithContext(context.Background(), "PUT", urlStr, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request with %v", err)
	}

	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("HTTP request failed with %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(response.Body)
		return fmt.Errorf("GitHub API responded with status code %d: %s", response.StatusCode, string(body))
	}

	return nil
}

func ListIssues(repoName string) ([]Issue, error) {
	profile, err := config.LoadUserProfile()
	if err != nil {
		return nil, fmt.Errorf("failed to load user profile with %v", err)
	}
	username := profile.GetUsername()

	token, err := config.GetToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token with %v", err)
	}

	urlStr := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues?state=all", username, repoName)
	client := &http.Client{}
	request, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request with %v", err)
	}

	request.Header.Set("Authorization", "Bearer "+token)

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed with %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(response.Body)
		return nil, fmt.Errorf("GitHub API responded with status code %d: %s", response.StatusCode, string(body))
	}

	var issues []Issue
	if err := json.NewDecoder(response.Body).Decode(&issues); err != nil {
		return nil, fmt.Errorf("failed to decode response with %v", err)
	}

	var openIssues []Issue
	for _, issue := range issues {
		if issue.State == "open" {
			openIssues = append(openIssues, issue)
		}
	}

	return openIssues, nil
}

func CloseIssue(repoName string, issueNumber int) error {
	profile, err := config.LoadUserProfile()
	if err != nil {
		return fmt.Errorf("failed to load user profile with %v", err)
	}
	username := profile.GetUsername()

	token, err := config.GetToken()
	if err != nil {
		return fmt.Errorf("failed to get auth toke with: %v", err)
	}

	urlStr := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d", username, repoName, issueNumber)
	body := map[string]string{
		"state": "closed",
	}
	data, _ := json.Marshal(body)

	request, err := http.NewRequest("PATCH", urlStr, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request with %v", err)
	}

	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("HTTP request failed with %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(response.Body)
		return fmt.Errorf("GitHub API responded with status code %d: %s", response.StatusCode, string(body))
	}

	return nil
}
