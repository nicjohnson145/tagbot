package cmd

import (
	"github.com/nicjohnson145/tagbot/config"
	"github.com/spf13/cobra"
)

func Root() *cobra.Command {
	root := &cobra.Command{
		Use:   "tagbot",
		Short: "analyze and create tag",
		Long:  "Analyze commits and create new tag if necessary",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			return config.InitializeConfig(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			tagbot, err := createTagbot()
			if err != nil {
				return err
			}
			return tagbot.Increment()
		},
	}
	root.PersistentFlags().BoolP(config.Debug, "d", false, "Enable debug logging")

	root.Flags().StringP(config.RemoteName, "r", "origin", "The remote name to push tags to")
	root.Flags().StringP(config.AuthMethod, "a", "", "Override the auth method to use to push tags")
	root.Flags().StringP(config.AuthToken, "t", "", "The auth token to use during token based auth")
	root.Flags().StringP(config.AuthTokenUsername, "u", "TagBot", "The auth token to use during token based auth")
	root.Flags().StringP(config.AuthKeyPath, "k", "", "Path to key to use during key based auth")

	root.Flags().BoolP(config.AlwaysPatch, "p", false, "Increment patch version even if no version bump is required")
	root.Flags().BoolP(config.Latest, "l", false, "Maintain a 'latest' tag")
	root.Flags().StringP(config.LatestName, "n", "", "Override latest tag name of 'latest'")

	root.Flags().Bool(config.NoPrefix, false, "Do not add a 'v' prefix to tags")

	root.AddCommand(
		versionCmd(),
		next(),
		commitMsg(),
		pullRequest(),
	)

	return root
}
