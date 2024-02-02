package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"teriyake/go-git-it/config"
	"teriyake/go-git-it/gitauth"
	"teriyake/go-git-it/gitops"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Info on current user",
	Long:  `Display the current user's profile, including the auth status, the list of to-do repos, and current repos.`,
	Run: func(cmd *cobra.Command, args []string) {
		if gitops.HasToken() {
			fmt.Printf("Git token exists!\n")
			me := gitauth.Whoami()
			fmt.Printf("You are %s\n", me)
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
