package cmd

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strconv"
	"strings"
	"teriyake/go-git-it/config"
)

var chooseRepoCmd = &cobra.Command{
	Use:   "choose-repo",
	Short: "Choose an existing to-do repo to work with",
	Long:  `This command allows the user to choose an existing to-do repo from their profile and sets it as the current working directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, err := config.LoadUserProfile()
		if err != nil {
			return fmt.Errorf("failed to load user profile with %v", err)
		}

		if len(profile.ToDoRepos) == 0 {
			fmt.Println("No existing to-do repos found. Please use 'new-repo' command to create one.")
			return nil
		}

		fmt.Println("Select a to-do repo by entering the corresponding number:")
		for i, repo := range profile.ToDoRepos {
			fmt.Printf("%d. %s\n", i+1, repo)
		}

		reader := bufio.NewReader(os.Stdin)
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)
		index, err := strconv.Atoi(choice)
		if err != nil || index < 1 || index > len(profile.ToDoRepos) {
			return fmt.Errorf("invalid selection, please enter a number between 1 and %d", len(profile.ToDoRepos))
		}

		selectedRepo := profile.ToDoRepos[index-1]
		fmt.Printf("Setting current directory to: %s\n", selectedRepo)
		profile.SetCurrentRepo(selectedRepo)
		profile.Save()

		return nil
	},
}
