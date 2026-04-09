package cmd

import (
	"os"

	"github.com/XDwanj/gx/internal/index"
	"github.com/XDwanj/gx/internal/query"

	"github.com/spf13/cobra"
)

func newOverviewCmd() *cobra.Command {
	var full bool

	command := &cobra.Command{
		Use:     "overview [flags] [path ...]",
		Aliases: []string{"o"},
		Short:   "Table of contents for a file or directory",
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime, err := buildRuntime()
			if err != nil {
				return err
			}

			targets, err := resolveTargetPaths(runtime.Root, args)
			if err != nil {
				return err
			}

			for _, target := range targets {
				debugf(rootCmd.ErrOrStderr(), "overview target=%s", target)
			}

			page, err := resolveOverviewPageRequest(cmd, targets)
			if err != nil {
				return err
			}

			var idx *index.Index
			if overviewTargetsRequireIndex(targets) {
				idx, err = loadIndex(runtime.Root, rootCmd.ErrOrStderr())
				if err != nil {
					return err
				}
			}

			return runtime.Query.Overview(idx, targets, full, page)
		},
	}

	command.Flags().BoolVar(&full, "full", false, "Show full per-file overview for directories")
	return command
}

func resolveOverviewPageRequest(cmd *cobra.Command, targets []string) (query.PageRequest, error) {
	if len(targets) > 1 {
		return resolvePageRequest(cmd, defaultDirectoryOverviewLimit)
	}

	if len(targets) == 1 && overviewTargetUsesDirectoryPaging(targets[0]) {
		return resolvePageRequest(cmd, defaultDirectoryOverviewLimit)
	}

	return query.PageRequest{}, nil
}

func overviewTargetsRequireIndex(targets []string) bool {
	for _, target := range targets {
		if overviewTargetRequiresIndex(target) {
			return true
		}
	}
	return false
}

func overviewTargetRequiresIndex(target string) bool {
	info, err := os.Stat(target)
	if err == nil && !info.IsDir() && query.IsMarkdownPath(target) {
		return false
	}
	return true
}

func overviewTargetUsesDirectoryPaging(target string) bool {
	info, err := os.Stat(target)
	return err == nil && info.IsDir()
}
