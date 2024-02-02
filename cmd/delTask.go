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

var delTaskCmd = &cobra.Command{
	Use:   "del-task",
	Short: "Delete a task file in the current to-do repo",
	Long:  `This command lists all the files in the current to-do repo and deletes the file selected by the user. The deletion is done both locally and remotely.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, err := config.LoadUserProfile()
		if err != nil {
			return fmt.Errorf("failed to load user profile with %v", err)
		}
		repoName := profile.GetCurrentRepo()
		repoPath := filepath.Join(os.Getenv("HOME"), ".go-git-it", "repos", repoName)

		lsCmd := exec.Command("ls", "-1", repoPath)
		output, err := lsCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to list filesw with %v", err)
		}
		files := strings.Split(string(output), "\n")
		for i, file := range files {
			fmt.Printf("%d: %s\n", i+1, file)
		}

		var choice int
		fmt.Print("Select a file to delete: ")
		_, err = fmt.Scan(&choice)
		if err != nil || choice < 1 || choice > len(files) {
			return fmt.Errorf("invalid selection")
		}
		selectedFile := files[choice-1]

		fmt.Printf("Are you sure you want to delete %s? [y/N]: ", selectedFile)
		var confirmation string
		_, err = fmt.Scan(&confirmation)
		if err != nil || strings.ToLower(confirmation) != "y" {
			fmt.Println("Deletion cancelled.")
			return nil
		}

		localFilePath := filepath.Join(os.Getenv("HOME"), ".go-git-it", "repos", repoName, selectedFile)
		rmCmd := exec.Command("rm", "-f", localFilePath)
		err = rmCmd.Run()
		if err != nil {
			return fmt.Errorf("failed to delete file located at %s with %v", localFilePath, err)
		}

		fmt.Println("Deleted", selectedFile, "locally. \nNow deleting ", selectedFile, " remotely...")
		sha, e := gitops.GetFileSHA(repoName, selectedFile)
		if e != nil {
			return fmt.Errorf("failed to get file SHA: %v", e)
		}
		return gitops.DeleteRemoteFile(repoName, selectedFile, sha)
	},
}
