package main

import (
	"github.com/nicjohnson145/tagbot/cmd"
	"github.com/rs/zerolog/log"
)

func main() {
	if err := cmd.Root().Execute(); err != nil {
		log.Fatal().Msg(err.Error())
	}
}
