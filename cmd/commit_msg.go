package cmd

import (
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
			return tagbot.CommitMessage(args[0])
		},
	}

	return cmd
}
