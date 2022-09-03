package main

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/spf13/cobra"
	"io"
)

func build(w io.Writer) *cobra.Command {
	root := rootCmd(w)

	return root
}

func rootCmd(w io.Writer) *cobra.Command {
	opts := IncrementOpts{}

	var debug bool

	root := &cobra.Command{
		Use: "tagbot",
		Short: "analyze and create tag",
		Long: "Analyze commits and create new tag if necessary",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.SetHandler(cli.New(w))
			if debug {
				log.SetLevel(log.DebugLevel)
			}
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return IncrementTag(opts)
		},
	}
	root.Flags().StringVarP(&opts.Path, "repo-path", "r", ".", "Path to repo")
	root.Flags().BoolVarP(&opts.AlwaysPatch, "always-patch", "p",  false, "Always increment the patch version, unless another version bump is required")
	root.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")

	return root
}
