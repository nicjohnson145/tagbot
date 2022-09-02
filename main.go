package main

import (
	"github.com/apex/log"
	"os"
)

func main() {
	if err := build(os.Stdout).Execute(); err != nil {
		log.WithError(err).Fatal("error during execution")
	}
}
