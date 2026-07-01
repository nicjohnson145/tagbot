package cmd

import (
	"context"
	"fmt"

	"github.com/nicjohnson145/hlp"
	"github.com/nicjohnson145/tagbot/internal/bot"
	"github.com/nicjohnson145/tagbot/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Root() *cobra.Command {
	cmd := &cobra.Command{
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			return config.InitConfig(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Base logging setup
			logger := config.NewLoggerFromEnv()

			// Construct our monorepo config, faking one if we're not in a monorepo
			var monorepoConf *config.MonoRepoConfig
			if viper.GetBool(config.MonoRepo) {
				c, err := config.ParseMonoRepoConfig(viper.GetString(config.MonoRepoConfigPath))
				if err != nil {
					logger.Err(err).Msg("error parsing monorepo config")
					return err
				}
				monorepoConf = c
			} else {
				monorepoConf = &config.MonoRepoConfig{
					Components: map[string]config.MonoRepoComponent{
						"core": {
							Name:           "core",
							ChangeSetGlobs: []string{"**/*"},
							Prefix:         hlp.Ptr(""),
							MaintainLatest: hlp.Ptr(viper.GetBool(config.MaintainLatest)),
							LatestName:     hlp.Ptr(viper.GetString(config.LatestName)),
							NoV:            hlp.Ptr(viper.GetBool(config.NoV)),
							AlwaysPatch:    hlp.Ptr(viper.GetBool(config.AlwaysPatch)),
						},
					},
				}
			}

			// Construct our git repo
			repo, err := bot.NewGitRepo(bot.GitRepoConfig{
				Path:              ".", // TODO: config option
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
				MonorepoConfig: monorepoConf,
				Repo:           repo,
			})

			// Embed our logger in a context so we can send it around
			ctx := logger.WithContext(context.Background())

			// Do the thing
			if err := tagbot.Run(ctx); err != nil {
				logger.Err(err).Msg("error executing")
				return err
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringP(config.LogLevel, "v", config.DefaultLogLevel, fmt.Sprintf("Logging output level, one of %v", config.LoggingLevelNames()))

	cmd.Flags().String(config.RemoteName, config.DefaultRemoteName, "The remote name to push tags to")
	cmd.Flags().String(config.AuthMethod, "", "Force the auth method to use to push tags, otherwise inferred from remote")
	cmd.Flags().String(config.AuthToken, "", "The auth token to use during token based auth")
	cmd.Flags().String(config.AuthTokenUsername, "TagBot", "The auth username to use during token based auth")
	cmd.Flags().String(config.AuthKeyPath, "", "Path to key to use during key based auth, sane defaults used otherwise")

	cmd.Flags().Bool(config.MonoRepo, config.DefaultMonoRepo, "Indicates this repo is a monorepo, and multiple tags should be managed")
	cmd.Flags().String(config.MonoRepoConfigPath, config.DefaultMonoRepoConfigPath, "Path to monorepo configuration file")

	cmd.Flags().Bool(config.MaintainLatest, config.DefaultMaintainLatest, "Maintain a latest tag. Applied to all non-overriden components in monorepo mode")
	cmd.Flags().String(config.LatestName, config.DefaultLatestName, "Name of latest, if maintained. Applied to all non-overriden components in monorepo mode")
	cmd.Flags().Bool(config.NoV, config.DefaultNoV, "Do not include the 'v' prefix on created tags. Applied to all non-overriden components in monorepo mode")
	cmd.Flags().Bool(config.AlwaysPatch, config.DefaultAlwaysPatch, "If commits would result in no version bump, instead patch. Applied to all non-overriden components in monorepo mode")

	cmd.AddCommand(CommitMessage())

	return cmd
}
