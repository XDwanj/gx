package cmd

import (
	"github.com/XDwanj/gx/internal/index"
	"github.com/XDwanj/gx/internal/query"

	"github.com/spf13/cobra"
)

func newSymbolsCmd() *cobra.Command {
	var name string
	var kind string
	var defineIn string
	var pathFilters pathFilterFlags

	command := &cobra.Command{
		Use:     "symbols [flags] [path ...]",
		Aliases: []string{"s"},
		Short:   "Search symbols across project",
		Long:    symbolsLongDescription(),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := buildService(cmd)
			root := service.Root()
			if err != nil {
				return err
			}

			page, err := resolvePageRequest(cmd, defaultSymbolsLimit)
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

			debugf(rootCmd.ErrOrStderr(), "symbols paths=%v name=%s kind=%s define_in=%s", paths, name, kind, resolvedDefineIn)

			idx, err := loadIndex(root, rootCmd.ErrOrStderr())
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

			return service.Symbols(idx, query.Intent{
				Name:      name,
				Kind:      kindPtr,
				DefinedIn: resolvedDefineIn,
			}, pathFilters.scope(paths), page)
		},
	}

	command.Flags().StringVar(&name, "name", "", "Glob pattern to match symbol names")
	command.Flags().StringVar(&kind, "kind", "", "Filter by symbol kind")
	registerDefineInFlag(command.Flags(), &defineIn)
	registerPathFilterFlags(command.Flags(), &pathFilters)
	if err := registerKindFlagCompletion(command, "kind"); err != nil {
		panic(err)
	}
	return command
}
