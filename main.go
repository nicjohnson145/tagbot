package main

import (
	"os"

	"github.com/nicjohnson145/tagbot/internal/cmd"
)

func main() {
	if err := cmd.Root().Execute(); err != nil {
		// root command takes care of logging
		os.Exit(1)
	}
}
