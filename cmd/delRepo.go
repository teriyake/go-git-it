package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"teriyake/go-git-it/config"
	"teriyake/go-git-it/gitops"
)

var delRepoCmd = &cobra.Command{
	Use:   "del-repo",
	Short: "Delete an existing to-do repo",
	Long:  `This command deletes an existing to-do repo, both locally and remotely, and updates the user profile.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, err := config.LoadUserProfile()
		if err != nil {
			return fmt.Errorf("failed to load user profile with %v", err)
		}

		if len(profile.ToDoRepos) == 0 {
			fmt.Println("No existing to-do repos found.")
			return nil
		}

		fmt.Println("Select a to-do repo to delete by entering the corresponding number:")
		for i, repo := range profile.ToDoRepos {
			fmt.Printf("%d. %s\n", i+1, repo)
		}

		var index int
		fmt.Print("Enter number: ")
		_, err = fmt.Scan(&index)
		if err != nil || index < 1 || index > len(profile.ToDoRepos) {
			return fmt.Errorf("invalid selection")
		}

		selectedRepo := profile.ToDoRepos[index-1]

		fmt.Printf("Are you sure you want to delete %s? [y/N]: ", selectedRepo)
		var confirmation string
		_, err = fmt.Scan(&confirmation)
		if err != nil || strings.ToLower(confirmation) != "y" {
			fmt.Println("Deletion cancelled.")
			return nil
		}

		err = gitops.DeleteRemoteRepo(selectedRepo)
		if err != nil {
			return fmt.Errorf("failed to delete remote repo:\n%v", err)
		}

		localRepoPath := filepath.Join(os.Getenv("HOME"), ".go-git-it", "repos", selectedRepo)
		rmCmd := exec.Command("rm", "-rf", localRepoPath)
		err = rmCmd.Run()
		if err != nil {
			return fmt.Errorf("failed to delete repo located at %s with %v", localRepoPath, err)
		}

		profile.RemoveRepo(selectedRepo)
		profile.Save()
		if profile.GetCurrentRepo() == selectedRepo {
			profile.SetCurrentRepo("")
		}
		if err := profile.Save(); err != nil {
			return fmt.Errorf("failed to update user profile with %v", err)
		}

		fmt.Printf("Repository %s has been deleted.\n", selectedRepo)
		return nil
	},
}
