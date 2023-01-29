package main

import (
	"os"
	"github.com/nicjohnson145/tagbot/cmd"
)

func main() {
	if err := cmd.Root().Execute(); err != nil {
		os.Exit(1)
	}
}
