package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"teriyake/go-git-it/gitauth"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Verify your Git auth status",
	Long:  `Verify whether you have successfully authorized ggi to perform necessary Git operations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Authorizing...\n")
		gitauth.Whoami()
	},
}
