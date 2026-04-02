package cmd

import (
	"gx/internal/index"

	"github.com/spf13/cobra"
)

func newSymbolsCmd() *cobra.Command {
	var file string
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

			idx, err := index.LoadOrBuild(runtime.Root)
			if err != nil {
				return err
			}

			var filePtr *string
			if file != "" {
				filePtr = &file
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

			return runtime.Query.Symbols(idx, filePtr, namePtr, kindPtr)
		},
	}

	command.Flags().StringVar(&file, "file", "", "Filter to a specific file")
	command.Flags().StringVar(&name, "name", "", "Glob pattern to match symbol names")
	command.Flags().StringVar(&kind, "kind", "", "Filter by symbol kind")
	return command
}
