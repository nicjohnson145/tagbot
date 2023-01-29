package cmd

import (
	"github.com/spf13/cobra"
)

func next() *cobra.Command {
	next := &cobra.Command{
		Use:   "next",
		Short: "check if tag required",
		Long:  "Check if new tag is required, but dont actually create it",
		RunE: func(cmd *cobra.Command, args []string) error {
			tagbot, err := createTagbot()
			if err != nil {
				return err
			}
			return tagbot.Next()
		},
	}

	return next
}
