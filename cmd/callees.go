package cmd

import "github.com/spf13/cobra"

func newCalleesCmd() *cobra.Command {
	command := &cobra.Command{
		Use:     "callees [flags] [path ...]",
		Aliases: []string{"c"},
		Short:   "Find function calls made inside matching symbols",
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime, err := buildRuntime()
			if err != nil {
				return err
			}

			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}

			page, err := resolvePageRequest(cmd, defaultCalleesLimit)
			if err != nil {
				return err
			}

			paths, err := resolveTargetPaths(runtime.Root, args)
			if err != nil {
				return err
			}

			debugf(rootCmd.ErrOrStderr(), "callees name=%s paths=%v", name, paths)

			idx, err := loadIndex(runtime.Root, rootCmd.ErrOrStderr())
			if err != nil {
				return err
			}

			return runtime.Query.Callees(idx, name, paths, page)
		},
	}

	command.Flags().String("name", "", "Glob pattern to match caller symbol names")
	_ = command.MarkFlagRequired("name")
	return command
}
