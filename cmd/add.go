package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"teriyake/go-git-it/gitops"
)

var addCmd = &cobra.Command{
	Use:   "add [task-file] [task-description]",
	Short: "Add a new task",
	Long:  `Add a new task by creating a file for the task and committing it with a description.`,
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		taskFile := args[0]
		taskDescription := args[1]
		if err := gitops.AddAndCommit(taskFile, taskDescription); err != nil {
			fmt.Fprintf(os.Stderr, "Error adding task: %s\n", err)
			return
		}
		fmt.Printf("Task added: %s\n", taskDescription)
	},
}
