package cmd

import "github.com/spf13/cobra"

func newReferencesCmd() *cobra.Command {
	var scope string
	var unique bool

	command := &cobra.Command{
		Use:     "references --name <glob>",
		Aliases: []string{"r"},
		Short:   "Find all usages of a symbol across the project",
		RunE: func(cmd *cobra.Command, _ []string) error {
			runtime, err := buildRuntime()
			if err != nil {
				return err
			}

			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}

			debugf(rootCmd.ErrOrStderr(), "references name=%s scope=%s unique=%t", name, scope, unique)

			idx, err := loadIndex(runtime.Root, rootCmd.ErrOrStderr())
			if err != nil {
				return err
			}

			var scopePtr *string
			if scope != "" {
				scopePtr = &scope
			}

			return runtime.Query.References(idx, name, scopePtr, unique)
		},
	}

	command.Flags().String("name", "", "Glob pattern to match symbol names")
	command.Flags().StringVar(&scope, "scope", "", "Limit search to a specific file or directory")
	command.Flags().BoolVar(&unique, "unique", false, "Deduplicate by enclosing function")
	_ = command.MarkFlagRequired("name")
	return command
}
