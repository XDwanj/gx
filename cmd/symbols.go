package cmd

import (
	"gx/internal/index"

	"github.com/spf13/cobra"
)

func newSymbolsCmd() *cobra.Command {
	var name string
	var kind string

	command := &cobra.Command{
		Use:     "symbols [flags] [path ...]",
		Aliases: []string{"s"},
		Short:   "Search symbols across project",
		RunE: func(_ *cobra.Command, args []string) error {
			runtime, err := buildRuntime()
			if err != nil {
				return err
			}

			paths, err := resolveTargetPaths(args)
			if err != nil {
				return err
			}

			debugf(rootCmd.ErrOrStderr(), "symbols paths=%v name=%s kind=%s", paths, name, kind)

			idx, err := loadIndex(runtime.Root, rootCmd.ErrOrStderr())
			if err != nil {
				return err
			}

			var namePtr *string
			if name != "" {
				namePtr = &name
			}

			var kindPtr *index.SymbolKind
			if kind != "" {
				value, parseErr := index.ParseSymbolKind(kind)
				if parseErr != nil {
					return parseErr
				}
				kindPtr = &value
			}

			return runtime.Query.Symbols(idx, paths, namePtr, kindPtr)
		},
	}

	command.Flags().StringVar(&name, "name", "", "Glob pattern to match symbol names")
	command.Flags().StringVar(&kind, "kind", "", "Filter by symbol kind")
	return command
}
