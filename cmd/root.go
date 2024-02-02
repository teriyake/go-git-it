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
	Use:   "ggi",
	Short: "Go-Git-It (ggi) is a CLI tool for managing your to-do list using Git functionalities",
	Long: `Go-Git-It (ggi) is a CLI tool that leverages Git functionalities to manage your to-do list.
You can create tasks with git commits, manage task categories through branches,
set deadlines using issues, collaborate using pull requests, and pause tasks using stash.`,
	Aliases: []string{"gg-it", "go-git-it"},
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
		profile.SetCurrentRepo(selectedRepo)
		profile.Save()

		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(newRepoCmd)
	rootCmd.AddCommand(chooseRepoCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(whoamiCmd)
	rootCmd.AddCommand(delRepoCmd)
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
