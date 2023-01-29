package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func commitMsg() *cobra.Command {
	cmd := &cobra.Command{
		Use: "commit-msg",
		Args: cobra.ExactArgs(1),
		Short: "Validate a commit message",
		Long: "Used as a commit-msg hook to ensure commits conform to tagbot expected format",
		RunE: func(cmd *cobra.Command, args []string) error {
			tagbot, err := createTagbot()
			if err != nil {
				return err
			}
			content, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("error reading commit message file: %w", err)
			}
			return tagbot.CommitMessage(string(content))
		},
	}

	return cmd
}
