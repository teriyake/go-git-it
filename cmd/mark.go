package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"teriyake/go-git-it/gitops"
	"teriyake/go-git-it/config"
)

var markCmd = &cobra.Command{
	Use:   "mark [issue number] ['done'/'doing'/'will-do']",
	Short: "Mark a to-do item with a status",
	Long: `Mark a to-do item as "done", "doing", or "will-do".
Example: mark 2 doing`,
	RunE: func(cmd *cobra.Command, args []string) error {

		profile, err := config.LoadUserProfile()
		if err != nil {
			return fmt.Errorf("failed to load user profile with %v", err)
		}

		repoName := profile.GetCurrentRepo()

		issues, err := gitops.ListIssues(repoName)
		if err != nil {
			fmt.Println("Error listing issues:", err)
			return err
		}

		if len(issues) == 0 {
			fmt.Println("No issues found.")
			return nil
		}

		fmt.Println("Select an issue by number:")
		for _, issue := range issues {
			fmt.Printf("#%d: %s\n", issue.Number, issue.Title)
		}

		var issueNumber int
		fmt.Print("Issue number: ")
		_, err = fmt.Scan(&issueNumber)
		if err != nil {
			fmt.Println("Invalid input:", err)
			return err
		}

		var status string
		fmt.Print("Enter status (done, doing, will-do): ")
		_, err = fmt.Scan(&status)
		if err != nil || (status != "done" && status != "doing" && status != "will-do") {
			fmt.Println("Invalid status:", err)
			return err
		}

		err = gitops.ChangeIssueLabel(repoName, issueNumber, []string{status})
		if err != nil {
			fmt.Println("Error updating issue:", err)
			return err
		}

		fmt.Printf("Issue #%d marked as %s.\n", issueNumber, status)
		return nil
	},
}
