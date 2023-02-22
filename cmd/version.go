package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// Version info set by goreleaser
var (
	version = "development"
	date    = "unknown"
)

func versionCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "version",
		Short: "Show tagbot version",
		Long:  "Show the version of tagbot & exit",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("   Version: %v\nBuild Date: %v\n", version, date)
			return nil
		},
	}

	return root
}
