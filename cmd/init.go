package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"teriyake/go-git-it/config"
	"teriyake/go-git-it/gitops"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Set up Git credentials",
	Long:  `Set up credentials to grant ggi access to perform Git operations on your behalf.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Initializing...\n")
		if gitops.HasToken() {
			fmt.Printf("Git token exists!\n")
		} else {
			fmt.Printf("Git token does not exist. Please use the 'login' sub-command to authorize first.\n")
		}
		profile, err := config.LoadUserProfile()
		if err != nil {
			fmt.Errorf("failed to load user profile: %v", err)
			os.Exit(1)
		}
		fmt.Printf("Existing to-do repos:\n%v\n", profile.ListRepos())
		fmt.Printf("Current to-do repo: %v\n", profile.GetCurrentRepo())
	},
}
