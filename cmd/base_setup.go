package cmd

import (
	"fmt"

	"github.com/nicjohnson145/tagbot/bot"
	"github.com/nicjohnson145/tagbot/config"
	"github.com/nicjohnson145/tagbot/git"
	"github.com/spf13/viper"
)

func createTagbot() (*bot.Tagbot, error) {
	logger := config.InitLogger()

	// Initialize git repo
	repo, err := git.NewGitRepo(git.Config{
		Logger:      config.WithComponent(logger, "gitrepo"),
		Path:        ".",
		Remote:      viper.GetString(config.RemoteName),
		AuthMethod:  viper.GetString(config.AuthMethod),
		AuthToken:   viper.GetString(config.AuthToken),
		AuthKeyPath: viper.GetString(config.AuthKeyPath),
	})
	if err != nil {
		logger.Err(err).Msg("initializing git repo")
		return nil, fmt.Errorf("error intializing git repo: %w", err)
	}

	// Initialize tagbot
	tagbot := bot.New(bot.Config{
		Logger:      config.WithComponent(logger, "tagbot"),
		Repo:        repo,
		AlwaysPatch: viper.GetBool(config.AlwaysPatch),
		Latest:      viper.GetBool(config.Latest),
	})
	return tagbot, nil
}
