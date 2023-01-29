package cmd

import (
	"fmt"

	"github.com/nicjohnson145/tagbot/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Root() *cobra.Command {
	root := &cobra.Command{
		Use: "tagbot",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return config.InitializeConfig(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("remote name: ", viper.GetString(config.RemoteName))
			fmt.Println("auth method: ", viper.GetString(config.AuthMethod))
			fmt.Println("auth token: ", viper.GetString(config.AuthToken))
			fmt.Println("auth key path: ", viper.GetString(config.AuthKeyPath))
			fmt.Println("debug: ", viper.GetBool(config.Debug))
			fmt.Println("patch: ", viper.GetBool(config.AlwaysPatch))
			fmt.Println("latest: ", viper.GetBool(config.Latest))

			return nil
		},
	}
	root.Flags().StringP(config.RemoteName, "r", "origin", "The remote name to push tags to")
	root.Flags().StringP(config.AuthMethod, "a", "", "Override the auth method to use to push tags")
	root.Flags().StringP(config.AuthToken, "t", "", "The auth token to use during token based auth")
	root.Flags().StringP(config.AuthKeyPath, "k", "", "Path to key to use during key based auth")

	root.Flags().BoolP(config.Debug, "d", false, "Enable debug logging")
	root.Flags().BoolP(config.AlwaysPatch, "p", false, "Increment patch version even if no version bump is required")
	root.Flags().BoolP(config.Latest, "l", false, "Maintain a 'latest' tag")

	return root
}
