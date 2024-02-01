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

	//taskPath := filepath.Join(repoPath, filepath.Base(filename))
	token, e := config.GetToken()
	if e != nil {
		return fmt.Errorf("failed to get auth token with %v", e)
	}
	//resp, s, e := uploadToGithub(token, repo, filename, dst, message, main)
	fmt.Printf("token (for debugging purposes):\n%v\n", token)
	return nil
}

/*
func githubRequest(token, method, urlString string, headers map[string]string, data interface{}, params url.Values) (*http.Response, map[string]interface{}, error) {
	if headers == nil {
		headers = make(map[string]string)
	}

	headers["User-Agent"] = "Agent 007"
	headers["Authorization"] = "Bearer " + token

	urlParsed, err := url.Parse(urlString)
	if err != nil {
		return nil, nil, err
	}

	urlPath := urlParsed.Path
	if params != nil {
		urlPath += "?" + params.Encode()
	}

	var body []byte
	if data != nil {
		body, err = json.Marshal(data)
		if err != nil {
			return nil, nil, err
		}
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, urlParsed.Scheme+"://"+urlParsed.Host+urlPath, bytes.NewBuffer(body))
	if err != nil {
		return nil, nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	if resp.StatusCode == 302 {
		location := resp.Header.Get("Location")
		return githubRequest(method, location, headers, data, params)
	}

	if resp.StatusCode >= 400 {
		delete(headers, "Authorization")
		respBody, _ := ioutil.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("error: %d - %s - %s - %s - %s - %v", resp.StatusCode, string(respBody), method, urlString, string(body), headers)
	}

	respBody, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	err = json.Unmarshal(respBody, &result)
	return resp, result, err
}

func uploadToGithub(token, repo, src, dst, gitMessage, branch string) (*http.Response, map[string]interface{}, error) {
	resp, data, err := githubRequest(token, "GET", "https://api.github.com/repos/"+repo+"/git/ref/"+branch, nil, nil, nil)
	if err != nil {
		return nil, nil, err
	}
	lastCommitSHA := data["object"].(map[string]interface{})["sha"].(string)
	fmt.Println("Last commit SHA: " + lastCommitSHA)

	fileData, err := ioutil.ReadFile(src)
	if err != nil {
		return nil, nil, err
	}
	base64Content := base64.StdEncoding.EncodeToString(fileData)

	resp, data, err = githubRequest(token, "POST", "https://api.github.com/repos/"+repository+"/git/blobs", nil, map[string]string{
		"content":  base64Content,
		"encoding": "base64",
	}, nil)
	if err != nil {
		return nil, nil, err
	}
	blobContentSHA := data["sha"].(string)

	resp, data, err = githubRequest(token, "POST", "https://api.github.com/repos/"+repository+"/git/trees", nil, map[string]interface{}{
		"base_tree": lastCommitSHA,
		"tree": []map[string]string{
			{
				"path": dst,
				"mode": "100644",
				"type": "blob",
				"sha":  blobContentSHA,
			},
		},
	}, nil)
	if err != nil {
		return nil, nil, err
	}
	treeSHA := data["sha"].(string)

	resp, data, err = githubRequest(token, "POST", "https://api.github.com/repos/"+repository+"/git/commits", nil, map[string]interface{}{
		"message": gitMessage,
		"parents": []string{lastCommitSHA},
		"tree":    treeSHA,
	}, nil)
	if err != nil {
		return nil, nil, err
	}
	newCommitSHA := data["sha"].(string)

	resp, data, err = githubRequest(token, "PATCH", "https://api.github.com/repos/"+repository+"/git/refs/"+strings.TrimPrefix(branch, "heads/"), nil, map[string]string{
		"sha": newCommitSHA,
	}, nil)
	return resp, data, err
}
*/

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
