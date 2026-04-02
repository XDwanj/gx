package cmd

import (
	"gx/internal/index"

	"github.com/spf13/cobra"
)

func newDefinitionCmd() *cobra.Command {
	var from string
	var kind string
	var maxLines int

	command := &cobra.Command{
		Use:     "definition --name <name>",
		Aliases: []string{"d"},
		Short:   "Get a function or type body without reading the whole file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			runtime, err := buildRuntime()
			if err != nil {
				return err
			}

			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}

			debugf(rootCmd.ErrOrStderr(), "definition name=%s", name)

			idx, err := loadIndex(runtime.Root, rootCmd.ErrOrStderr())
			if err != nil {
				return err
			}

			var fromPtr *string
			if from != "" {
				fromPtr = &from
			}

			var kindPtr *index.SymbolKind
			if kind != "" {
				value, parseErr := index.ParseSymbolKind(kind)
				if parseErr != nil {
					return parseErr
				}
				kindPtr = &value
			}

			return runtime.Query.Definition(idx, name, fromPtr, kindPtr, maxLines)
		},
	}

	command.Flags().String("name", "", "Symbol name to look up")
	command.Flags().StringVar(&from, "from", "", "Disambiguate by source file")
	command.Flags().StringVar(&kind, "kind", "", "Filter by symbol kind")
	command.Flags().IntVar(&maxLines, "max-lines", 200, "Max lines for body output")
	_ = command.MarkFlagRequired("name")

	return command
}
