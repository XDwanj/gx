package cmd

import (
	"gx/internal/index"

	"github.com/spf13/cobra"
)

func newSymbolsCmd() *cobra.Command {
	var scope string
	var name string
	var kind string

	command := &cobra.Command{
		Use:     "symbols",
		Aliases: []string{"s"},
		Short:   "Search symbols across project",
		RunE: func(_ *cobra.Command, _ []string) error {
			runtime, err := buildRuntime()
			if err != nil {
				return err
			}

			debugf(rootCmd.ErrOrStderr(), "symbols scope=%s name=%s kind=%s", scope, name, kind)

			idx, err := loadIndex(runtime.Root, rootCmd.ErrOrStderr())
			if err != nil {
				return err
			}

			var scopePtr *string
			if scope != "" {
				scopePtr = &scope
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

			return runtime.Query.Symbols(idx, scopePtr, namePtr, kindPtr)
		},
	}

	command.Flags().StringVar(&scope, "scope", "", "Filter to a specific file or directory")
	command.Flags().StringVar(&name, "name", "", "Glob pattern to match symbol names")
	command.Flags().StringVar(&kind, "kind", "", "Filter by symbol kind")
	return command
}
