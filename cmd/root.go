package cmd

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strconv"
	"strings"
	"teriyake/go-git-it/config"
	"teriyake/go-git-it/gitops"
)

var rootCmd = &cobra.Command{
	Use:   "go-git-it",
	Short: "Go-Git-It is a CLI tool for managing your to-do list using Git functionalities",
	Long: `Go-Git-It is a CLI tool that leverages Git functionalities to manage your to-do list.
You can create tasks with git commits, manage task categories through branches,
set deadlines using issues, collaborate using pull requests, and pause tasks using stash.`,
}

var chooseRepoCmd = &cobra.Command{
	Use:   "choose-repo",
	Short: "Choose an existing to-do repo to work with",
	Long:  `This command allows the user to choose an existing to-do repo from their profile and sets it as the current working directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, err := config.LoadUserProfile()
		if err != nil {
			return fmt.Errorf("failed to load user profile: %v", err)
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

		if err := os.Chdir(selectedRepo); err != nil {
			return fmt.Errorf("failed to change directory: %v", err)
		}
		// save the current repo as the last used repo in the profile?

		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(newRepoCmd)
	// more cmds... 

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if !gitops.IsGitRepo() {
			profile, err := config.LoadUserProfile()
			if err != nil {
				return fmt.Errorf("failed to load user profile: %v", err)
			}
			if len(profile.ToDoRepos) == 0 {
				fmt.Println("No existing to-do repos found. Please create a new to-do repo.")
				return nil
				// handle creating new repo
			}
			fmt.Println("Select a to-do repo to work with:")
			for i, repo := range profile.ToDoRepos {
				fmt.Printf("%d. %s\n", i+1, repo)
			}
		}
		return nil
	}
}
