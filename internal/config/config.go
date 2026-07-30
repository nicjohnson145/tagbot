package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/nicjohnson145/hlp"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//go:generate go-enum -f $GOFILE -marshal -names

/*
ENUM(
trace
debug
info
warning
error
)
*/
type LoggingLevel string

const (
	LogLevel = "log-level"

	RemoteName        = "remote-name"
	AuthMethod        = "auth-method"
	AuthToken         = "auth-token"
	AuthTokenUsername = "auth-token-username"
	AuthKeyPath       = "auth-key-path"

	MonoRepo           = "monorepo"
	MonoRepoConfigPath = "monorepo-config-path"

	MaintainLatest = "maintain-latest"
	LatestName     = "latest-name"
	NoV            = "no-v"
	AlwaysPatch    = "always-patch"

	DryRun = "dry-run"
)

var (
	DefaultLogLevel = LoggingLevelInfo.String()

	DefaultRemoteName = "origin"
	DefaultAuthTokenUsername = "TagBot"

	DefaultMonoRepo           = false
	DefaultMonoRepoConfigPath = "./.tagbot.yaml"

	DefaultMaintainLatest = false
	DefaultLatestName     = "latest"
	DefaultNoV            = false
	DefaultAlwaysPatch    = false

	DefaultDryRun = false
)

func InitConfig(cmd *cobra.Command) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting user home directory: %w", err)
	}

	setDefaultKeyPath := func() error {
		options := []string{
			filepath.Join(home, ".ssh", "id_rsa"),
			filepath.Join(home, ".ssh", "id_ecdsa"),
		}

		for _, option := range options {
			if _, err := os.Stat(option); err == nil {
				viper.SetDefault(AuthKeyPath, option)
				return nil
			} else {
				if !os.IsNotExist(err) {
					return fmt.Errorf("error checking existence of %v: %w", option, err)
				}
			}
		}

		return nil
	}
	if err := setDefaultKeyPath(); err != nil {
		return err
	}

	viper.SetDefault(LogLevel, DefaultLogLevel)

	viper.SetDefault(RemoteName, DefaultRemoteName)
	viper.SetDefault(AuthTokenUsername, DefaultAuthTokenUsername)

	viper.SetDefault(MonoRepo, DefaultMonoRepo)
	viper.SetDefault(MonoRepoConfigPath, DefaultMonoRepoConfigPath)

	viper.SetDefault(MaintainLatest, DefaultMaintainLatest)
	viper.SetDefault(LatestName, DefaultLatestName)
	viper.SetDefault(NoV, DefaultNoV)
	viper.SetDefault(AlwaysPatch, DefaultAlwaysPatch)

	viper.SetDefault(DryRun, DefaultDryRun)

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return fmt.Errorf("error binding flags: %w", err)
	}

	return nil
}

func NewLoggerFromEnv() zerolog.Logger {
	level, err := ParseLoggingLevel(viper.GetString(LogLevel))
	if err != nil {
		fmt.Printf("unable to parse logging level, falling back to info: %v\n", err)
		level = LoggingLevelInfo
	}

	return NewLogger(LoggerOpts{
		Level: level,
	})
}

type LoggerOpts struct {
	Level LoggingLevel
}

func NewLogger(opts LoggerOpts) zerolog.Logger {
	logger := zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}).With().Timestamp().Logger()

	var level zerolog.Level
	switch opts.Level {
	case LoggingLevelWarning:
		level = zerolog.WarnLevel
	case LoggingLevelInfo:
		level = zerolog.InfoLevel
	case LoggingLevelDebug:
		level = zerolog.DebugLevel
	case LoggingLevelTrace:
		level = zerolog.TraceLevel
	default:
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)

	return logger
}

type MonoRepoConfig struct {
	Components map[string]MonoRepoComponent `yaml:"components"`
}

type MonoRepoComponent struct {
	Name           string   `yaml:"-"`
	ChangeSetGlobs []string `yaml:"change-set-globs"`
	Prefix         *string  `yaml:"prefix,omitempty"`
	MaintainLatest *bool    `yaml:"maintain-latest,omitempty"`
	LatestName     *string  `yaml:"latest-name,omitempty"`
	NoV            *bool    `yaml:"no-v,omitempty"`
	AlwaysPatch    *bool    `yaml:"always-patch,omitempty"`
}

func ParseMonoRepoConfig(path string) (*MonoRepoConfig, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	conf := &MonoRepoConfig{}
	if err := yaml.Unmarshal(content, conf); err != nil {
		return nil, fmt.Errorf("erorr unmarshalling: %w", err)
	}

	// Validate & post-process the loaded config
	for name := range conf.Components {
		component := conf.Components[name]

		component.Name = name
		if len(component.ChangeSetGlobs) == 0 {
			return nil, fmt.Errorf("%v: component has no changeset globs", name)
		}
		if component.MaintainLatest == nil {
			component.MaintainLatest = hlp.Ptr(viper.GetBool(MaintainLatest))
		}
		if component.LatestName == nil {
			component.LatestName = hlp.Ptr(viper.GetString(LatestName))
		}
		if component.NoV == nil {
			component.NoV = hlp.Ptr(viper.GetBool(NoV))
		}
		if component.AlwaysPatch == nil {
			component.AlwaysPatch = hlp.Ptr(viper.GetBool(AlwaysPatch))
		}

		conf.Components[name] = component
	}

	return conf, nil
}

/*
ENUM(
ssh
https
)
*/
type RemoteType string

/*
ENUM(
public-key
token
)
*/
type AuthKind string

var AuthToRemoteMap = map[AuthKind]RemoteType{
	AuthKindToken:     RemoteTypeHttps,
	AuthKindPublicKey: RemoteTypeSsh,
}
