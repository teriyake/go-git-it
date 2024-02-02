package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"teriyake/go-git-it/gitauth"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Verify your Github auth status",
	Long:  `Verify whether you have successfully authorized ggi to perform necessary Git operations via Github API calls.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Authorizing...\n")
		me := gitauth.Whoami()
		fmt.Printf("You are %s\n", me)
	},
}
