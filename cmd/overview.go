package cmd

import (
	"gx/internal/query"
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

			target := args[0]
			debugf(rootCmd.ErrOrStderr(), "overview target=%s", target)

			abs := resolveTargetPath(target, runtime.Root)
			info, err := os.Stat(abs)
			if err == nil && info.IsDir() {
				idx, loadErr := loadIndex(runtime.Root, rootCmd.ErrOrStderr())
				if loadErr != nil {
					return loadErr
				}
				return runtime.Query.DirectoryOverview(idx, target, full)
			}
			if err == nil && query.IsMarkdownPath(abs) {
				return runtime.Query.MarkdownOverview(target)
			}

			idx, err := loadIndex(runtime.Root, rootCmd.ErrOrStderr())
			if err != nil {
				return err
			}
			return runtime.Query.Symbols(idx, &target, nil, nil)
		},
	}

	command.Flags().BoolVar(&full, "full", false, "Show full per-file overview for directories")
	return command
}
