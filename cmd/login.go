package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"teriyake/go-git-it/gitauth"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Set up Git credentials",
	Long:  `Set up credentials to grant ggi access to perform Git operations on your behalf.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Please follow the prompts to log in...\n")
		gitauth.Login()
	},
}
