package main

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/spf13/cobra"
	"io"
)

func build(w io.Writer) *cobra.Command {
	root := rootCmd(w)
	root.AddCommand(nextCmd())
	root.AddCommand(commitMessage())
	root.AddCommand(pullRequest())

	return root
}

func rootCmd(w io.Writer) *cobra.Command {
	opts := IncrementOpts{}

	var debug bool

	cmd := &cobra.Command{
		Use: "tagbot",
		Short: "analyze and create tag",
		Long: "Analyze commits and create new tag if necessary",
		Args: cobra.MaximumNArgs(1),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.SetHandler(cli.New(w))
			if debug {
				log.SetLevel(log.DebugLevel)
			}
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			setPath(args, &opts)
			return IncrementTag(opts)
		},
	}
	cmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")
	cmd.Flags().BoolVarP(&opts.AlwaysPatch, "always-patch", "p",  false, "Always increment the patch version, unless another version bump is required")
	cmd.Flags().BoolVarP(&opts.MaintainLatest, "maintain-latest", "l",  false, "Maintain a 'latest' tag")

	return cmd
}

func nextCmd() *cobra.Command {
	opts := NextOpts{}

	cmd := &cobra.Command{
		Use: "next",
		Short: "check if tag required",
		Long: "Check if new tag is required, but dont actually create it",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			setPath(args, &opts)
			return Next(opts)
		},
	}
	return cmd
}

func setPath(args []string, opts PathSetter) {
	if len(args) == 0 {
		opts.SetPath(".")
	} else {
		opts.SetPath(args[0])
	}
}

func commitMessage() *cobra.Command {
	cmd := &cobra.Command{
		Use: "commit-msg",
		Short: "Validate a commit message",
		Long: "Used as a commit-msg hook to ensure commits conform to tagbot expected format",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return CommitMessage(args[0])
		},
	}

	return cmd
}

func pullRequest() *cobra.Command {
	opts := PullRequestOpts{}

	cmd := &cobra.Command{
		Use: "pull-request",
		Short: "Validate pull request",
		Long: "Validate all commits on a given pull request conform to tagbots expected format",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			setPath(args, &opts)
			return PullRequest(opts)
		},
	}

	return cmd
}
