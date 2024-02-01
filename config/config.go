package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type UserProfile struct {
	ToDoRepos []string `json:"to_do_repos"`
}

var (
	profilePath = filepath.Join(os.Getenv("HOME"), ".go-git-it", "profile.json")
)

func LoadUserProfile() (*UserProfile, error) {
	var profile UserProfile

	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return &profile, nil
	}

	data, err := ioutil.ReadFile(profilePath)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

func (p *UserProfile) Save() error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	dir := filepath.Dir(profilePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return ioutil.WriteFile(profilePath, data, 0644)
}

func (p *UserProfile) AddRepo(repoPath string) {
	for _, r := range p.ToDoRepos {
		if r == repoPath {
			return
		}
	}
	p.ToDoRepos = append(p.ToDoRepos, repoPath)
}
