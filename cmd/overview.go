package cmd

import (
	"gx/internal/query"
	"os"

	"github.com/spf13/cobra"
)

func newOverviewCmd() *cobra.Command {
	var full bool

	command := &cobra.Command{
		Use:     "overview [flags] [path]",
		Aliases: []string{"o"},
		Short:   "Table of contents for a file or directory",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime, err := buildRuntime()
			if err != nil {
				return err
			}

			targets, err := resolveTargetPaths(runtime.Root, args)
			if err != nil {
				return err
			}
			target := targets[0]
			debugf(rootCmd.ErrOrStderr(), "overview target=%s", target)

			info, err := os.Stat(target)
			if err == nil && info.IsDir() {
				page, pageErr := resolvePageRequest(cmd, defaultDirectoryOverviewLimit)
				if pageErr != nil {
					return pageErr
				}
				idx, loadErr := loadIndex(runtime.Root, rootCmd.ErrOrStderr())
				if loadErr != nil {
					return loadErr
				}
				return runtime.Query.DirectoryOverview(idx, target, full, page)
			}
			if err == nil && query.IsMarkdownPath(target) {
				return runtime.Query.MarkdownOverview(target)
			}

			idx, err := loadIndex(runtime.Root, rootCmd.ErrOrStderr())
			if err != nil {
				return err
			}
			return runtime.Query.Symbols(idx, []string{target}, nil, nil, query.PageRequest{})
		},
	}

	command.Flags().BoolVar(&full, "full", false, "Show full per-file overview for directories")
	return command
}
