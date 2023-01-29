package config

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func InitLogger() zerolog.Logger {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if viper.GetBool(Debug) {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	return log.With().Logger()
}

func WithComponent(logger zerolog.Logger, name string) zerolog.Logger {
	return logger.With().Str("component", name).Logger()
}
