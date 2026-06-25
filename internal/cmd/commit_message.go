package cmd

import (
	"context"
	"os"

	"github.com/nicjohnson145/tagbot/internal/bot"
	"github.com/nicjohnson145/tagbot/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func CommitMessage() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit-msg <commit file>",
		Args:  cobra.ExactArgs(1),
		Short: "Validate a commit message",
		Long:  "Used as a commit-msg hook to ensure commits conform to tagbot expected format",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Base logging setup
			logger := config.NewLoggerFromEnv()

			// Construct our git repo
			repo, err := bot.NewGitRepo(bot.GitRepoConfig{
				Path: ".", // TODO: config option
				Remote:            viper.GetString(config.RemoteName),
				AuthMethod:        viper.GetString(config.AuthMethod),
				AuthToken:         viper.GetString(config.AuthToken),
				AuthKeyPath:       viper.GetString(config.AuthKeyPath),
				AuthTokenUsername: viper.GetString(config.AuthTokenUsername),
			})
			if err != nil {
				logger.Err(err).Msg("error creating git repo handle")
				return err
			}

			tagbot := bot.NewTagbot(bot.TagbotConfig{
				Repo: repo,
			})

			content, err := os.ReadFile(args[0])
			if err != nil {
				logger.Err(err).Msg("error reading commit message file")
				return err
			}

			// Embed our logger in a context so we can send it around
			ctx := logger.WithContext(context.Background())

			// Do the thing
			if err := tagbot.CommitMessage(ctx, string(content)); err != nil {
				logger.Err(err).Msg("error executing")
				return err
			}

			return nil
		},
	}

	return cmd
}
