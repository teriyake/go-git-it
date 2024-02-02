package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"teriyake/go-git-it/gitops"
)

var (
	deadline string
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

		if deadline != "" {
			if err := gitops.SetDeadline(taskDescription, deadline); err != nil {
				fmt.Printf("failed to set deadline with %v\n", err)
				return
			}
			fmt.Println("Deadline set successfully.")
		}
	},
}

func init() {
	addCmd.Flags().StringVarP(&deadline, "deadline", "d", "", "Optional deadline for the task (format: YYYY-MM-DD)")
}
