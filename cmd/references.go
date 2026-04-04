package cmd

import "github.com/spf13/cobra"

func newReferencesCmd() *cobra.Command {
	var unique bool

	command := &cobra.Command{
		Use:     "references [flags] [path ...]",
		Aliases: []string{"r"},
		Short:   "Find all usages of a symbol across the project",
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime, err := buildRuntime()
			if err != nil {
				return err
			}

			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}

			paths, err := resolveTargetPaths(runtime.Root, args)
			if err != nil {
				return err
			}

			debugf(rootCmd.ErrOrStderr(), "references name=%s paths=%v unique=%t", name, paths, unique)

			idx, err := loadIndex(runtime.Root, rootCmd.ErrOrStderr())
			if err != nil {
				return err
			}

			return runtime.Query.References(idx, name, paths, unique)
		},
	}

	command.Flags().String("name", "", "Glob pattern to match symbol names")
	command.Flags().BoolVar(&unique, "unique", false, "Deduplicate by enclosing function")
	_ = command.MarkFlagRequired("name")
	return command
}
