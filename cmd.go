package main

import (
	"github.com/spf13/cobra"
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"io"
)

func build(w io.Writer) *cobra.Command {
	root := rootCmd(w)

	return root
}

func rootCmd(w io.Writer) *cobra.Command {
	return &cobra.Command{
		Use: "tagbot",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.SetHandler(cli.New(w))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info("wooo")
			return nil
		},
	}
}
