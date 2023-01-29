package cmd

import (
	"github.com/nicjohnson145/tagbot/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func pullRequest() *cobra.Command {
	cmd := &cobra.Command{
		Use: "pull-request",
		Short: "Validate pull request",
		Long: "Validate all commits on a given pull request conform to tagbots expected format",
		RunE: func(cmd *cobra.Command, args []string) error {
			tagbot, err := createTagbot()
			if err != nil {
				return err
			}
			return tagbot.PullRequest(viper.GetString(config.BaseBranch))
		},
	}

	cmd.Flags().StringP(config.BaseBranch, "b", "", "The base branch for a pull request")

	return cmd
}
