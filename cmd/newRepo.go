package cmd

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	_ "path/filepath"
	"strings"
	"teriyake/go-git-it/config"
	"teriyake/go-git-it/gitops"
)

var newRepoCmd = &cobra.Command{
	Use:   "new-repo",
	Short: "Create a new to-do repo",
	Long:  `This command creates a new to-do repo in a specified directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Enter the path for the new to-do repo or press Enter to use the current directory:")
		reader := bufio.NewReader(os.Stdin)
		path, _ := reader.ReadString('\n')
		path = strings.TrimSpace(path)
		if path == "" {
			path, _ = os.Getwd()
		} else {
			/*
				var err error
				path, err = filepath.Abs(path)
				if err != nil {
					return fmt.Errorf("failed to resolve absolute path: %v", err)
				}
			*/
			path = path
		}

		if err := gitops.CreateNewRepo(path, false); err != nil {
			return err
		}

		profile, err := config.LoadUserProfile()
		if err != nil {
			return fmt.Errorf("failed to load user profile: %v", err)
		}
		profile.AddRepo(path)
		profile.SetCurrentRepo(path)
		if err := profile.Save(); err != nil {
			return fmt.Errorf("failed to save user profile: %v", err)
		}

		fmt.Printf("New to-do repo initialized at %s\n", path)
		fmt.Printf("Current to-do repo: %s\n", profile.GetCurrentRepo())
		return nil
	},
}
