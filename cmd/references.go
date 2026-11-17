package cmd

import (
	"github.com/XDwanj/gx/internal/query"

	"github.com/spf13/cobra"
)

func newReferencesCmd() *cobra.Command {
	var unique bool
	var defineIn string
	var pathFilters pathFilterFlags

	command := &cobra.Command{
		Use:     "references [flags] [path ...]",
		Aliases: []string{"r"},
		Short:   "Find all usages of a symbol across the project",
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := buildService(cmd)
			root := service.Root()
			if err != nil {
				return err
			}

			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}

			page, err := resolvePageRequest(cmd, defaultReferencesLimit)
			if err != nil {
				return err
			}

			paths, err := resolveTargetPaths(root, args)
			if err != nil {
				return err
			}
			resolvedDefineIn, err := resolveDefineInPath(root, defineIn)
			if err != nil {
				return err
			}

			debugf(rootCmd.ErrOrStderr(), "references name=%s paths=%v unique=%t define_in=%s", name, paths, unique, resolvedDefineIn)

			idx, err := loadIndex(root, rootCmd.ErrOrStderr())
			if err != nil {
				return err
			}

			return service.References(idx, query.Intent{
				Name:      name,
				DefinedIn: resolvedDefineIn,
			}, pathFilters.scope(paths), page, unique)
		},
	}

	command.Flags().String("name", "", "Glob pattern to match symbol names")
	command.Flags().BoolVar(&unique, "unique", false, "Deduplicate by enclosing function")
	registerDefineInFlag(command.Flags(), &defineIn)
	registerPathFilterFlags(command.Flags(), &pathFilters)
	_ = command.MarkFlagRequired("name")
	return command
}
