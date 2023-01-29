package config

import (
	"fmt"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	Debug       = "debug"
	Latest      = "latest"
	AlwaysPatch = "always-patch"
	RemoteName  = "remote-name"
	AuthMethod  = "auth-method"
	AuthToken   = "auth-token"
	AuthKeyPath = "auth-key-path"
)

func InitializeConfig(cmd *cobra.Command) error {
	path, err := homedir.Expand("~/.ssh/id_ecdsa")
	if err != nil {
		return fmt.Errorf("error expanding home dir: %w", err)
	}
	viper.SetDefault(AuthKeyPath, path)

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.BindPFlags(cmd.Flags())

	return nil
}
