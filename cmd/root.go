package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"teriyake/go-git-it/config"
	"teriyake/go-git-it/gitops"
)

var rootCmd = &cobra.Command{
	Use:   "ggi",
	Short: "Go-Git-It (ggi) is a CLI tool for managing your to-do list using Git functionalities",
	Long: `Go-Git-It (ggi) is a CLI tool that leverages Git functionalities to manage your to-do list.
You can create tasks with git commits, manage task categories through branches,
set deadlines using issues, collaborate using pull requests, and pause tasks using stash.`,
	Aliases: []string{"gg-it", "go-git-it"},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(newRepoCmd)
	rootCmd.AddCommand(chooseRepoCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(whoamiCmd)
	rootCmd.AddCommand(delRepoCmd)
	rootCmd.AddCommand(markCmd)
	rootCmd.AddCommand(doneCmd)
	rootCmd.AddCommand(delTaskCmd)
	// more cmds...

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		profile, err := config.LoadUserProfile()
		if err != nil {
			return fmt.Errorf("failed to load user profile: %v", err)
		}

		if len(profile.ToDoRepos) == 0 {
			fmt.Println("No existing to-do repos found. Please use 'new-repo' command to create one.")
			return nil
		}

		fmt.Println("Select a to-do repo to work with by entering the corresponding number:")
		for i, repo := range profile.ToDoRepos {
			fmt.Printf("%d. %s\n", i+1, repo)
		}

		var index int
		fmt.Print("Enter number: ")
		_, err = fmt.Scan(&index)
		if err != nil || index < 1 || index > len(profile.ToDoRepos) {
			fmt.Println("Invalid selection. Please restart the command and select a valid number.")
			os.Exit(1)
		}

		selectedRepo := profile.ToDoRepos[index-1]
		fmt.Printf("Setting current directory to: %s\n", selectedRepo)

		if err := os.Chdir(selectedRepo); err != nil {
			fmt.Errorf("failed to change directory: %v", err)
			os.Exit(1)
		}

		return nil
	}

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
