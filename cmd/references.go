package cmd

import "github.com/spf13/cobra"

func newReferencesCmd() *cobra.Command {
	var file string
	var unique bool

	command := &cobra.Command{
		Use:     "references --name <name>",
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

			debugf(rootCmd.ErrOrStderr(), "references name=%s file=%s unique=%t", name, file, unique)

			idx, err := loadIndex(runtime.Root, rootCmd.ErrOrStderr())
			if err != nil {
				return err
			}

			var filePtr *string
			if file != "" {
				filePtr = &file
			}

			return runtime.Query.References(idx, name, filePtr, unique)
		},
	}

	command.Flags().String("name", "", "Symbol name to find")
	command.Flags().StringVar(&file, "file", "", "Limit search to a specific file")
	command.Flags().BoolVar(&unique, "unique", false, "Deduplicate by enclosing function")
	_ = command.MarkFlagRequired("name")
	return command
}
