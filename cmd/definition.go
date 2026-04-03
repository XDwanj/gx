package cmd

import (
	"gx/internal/index"

	"github.com/spf13/cobra"
)

func newDefinitionCmd() *cobra.Command {
	var kind string
	var maxLines int

	command := &cobra.Command{
		Use:     "definition [flags] [path ...]",
		Aliases: []string{"d"},
		Short:   "Get a function or type body without reading the whole file",
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime, err := buildRuntime()
			if err != nil {
				return err
			}

			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}

			paths, err := resolveTargetPaths(args)
			if err != nil {
				return err
			}

			debugf(rootCmd.ErrOrStderr(), "definition name=%s paths=%v", name, paths)

			idx, err := loadIndex(runtime.Root, rootCmd.ErrOrStderr())
			if err != nil {
				return err
			}

			var kindPtr *index.SymbolKind
			if kind != "" {
				value, parseErr := index.ParseSymbolKind(kind)
				if parseErr != nil {
					return parseErr
				}
				kindPtr = &value
			}

			return runtime.Query.Definition(idx, name, paths, kindPtr, maxLines)
		},
	}

	command.Flags().String("name", "", "Glob pattern to match symbol names")
	command.Flags().StringVar(&kind, "kind", "", "Filter by symbol kind")
	command.Flags().IntVar(&maxLines, "max-lines", 200, "Max lines for body output")
	_ = command.MarkFlagRequired("name")

	return command
}
