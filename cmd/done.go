package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"teriyake/go-git-it/config"
	"teriyake/go-git-it/gitops"
)

var doneCmd = &cobra.Command{
	Use:   "done",
	Short: "Mark a to-do item as done by closing the corresponding Github issue",
	Long:  `This command lists all open issues in the current to-do repo, and the user will select one to close.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		profile, err := config.LoadUserProfile()
		if err != nil {
			return fmt.Errorf("failed to load user profile with %v", err)
		}
		repoName := profile.GetCurrentRepo()

		issues, err := gitops.ListIssues(repoName)
		if err != nil {
			fmt.Println("Error listing issues: ", err)
			return err
		}

		if len(issues) == 0 {
			fmt.Println("No issues found.")
			return nil
		}

		fmt.Println("Select an ongoing to-do item by number:")
		for _, issue := range issues {
			if issue.State == "open" {
				fmt.Printf("#%d: %s\n", issue.Number, issue.Title)
			}
		}

		var issueNumber int
		fmt.Print("Issue number: ")
		_, err = fmt.Scan(&issueNumber)
		if err != nil {
			fmt.Println("Invalid input:", err)
			return err
		}

		err = gitops.CloseIssue(repoName, issueNumber)
		if err != nil {
			fmt.Println("Error closing issue: ", err)
			return err
		}

		fmt.Printf("To-do item associated with issue #%d completed successfully.\n", issueNumber)
		return nil
	},
}
