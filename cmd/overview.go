package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func newOverviewCmd() *cobra.Command {
	var full bool

	command := &cobra.Command{
		Use:     "overview <path>",
		Aliases: []string{"o"},
		Short:   "Table of contents for a file or directory",
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			runtime, err := buildRuntime()
			if err != nil {
				return err
			}

			debugf(rootCmd.ErrOrStderr(), "overview target=%s", args[0])

			idx, err := loadIndex(runtime.Root, rootCmd.ErrOrStderr())
			if err != nil {
				return err
			}

			target := args[0]
			abs := resolveTargetPath(target, runtime.Root)
			info, err := os.Stat(abs)
			if err == nil && info.IsDir() {
				return runtime.Query.DirectoryOverview(idx, target, full)
			}
			return runtime.Query.Symbols(idx, &target, nil, nil)
		},
	}

	command.Flags().BoolVar(&full, "full", false, "Show full per-file overview for directories")
	return command
}
