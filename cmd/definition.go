package cmd

import (
	"github.com/XDwanj/gx/internal/index"
	"github.com/XDwanj/gx/internal/query"

	"github.com/spf13/cobra"
)

func newDefinitionCmd() *cobra.Command {
	var kind string
	var maxLines int
	var defineIn string
	var pathFilters pathFilterFlags

	command := &cobra.Command{
		Use:     "definition [flags] [path ...]",
		Aliases: []string{"d"},
		Short:   "Get a function or type body without reading the whole file",
		Long:    definitionLongDescription(),
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

			page, err := resolvePageRequest(cmd, defaultDefinitionLimit)
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

			debugf(rootCmd.ErrOrStderr(), "definition name=%s paths=%v define_in=%s", name, paths, resolvedDefineIn)

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

			return service.Definition(idx, query.Intent{
				Name:      name,
				Kind:      kindPtr,
				DefinedIn: resolvedDefineIn,
			}, pathFilters.scope(paths), page, maxLines)
		},
	}

	command.Flags().String("name", "", "Glob pattern to match symbol names")
	command.Flags().StringVar(&kind, "kind", "", "Filter by symbol kind")
	command.Flags().IntVar(&maxLines, "max-lines", 200, "Max lines for body output")
	registerDefineInFlag(command.Flags(), &defineIn)
	registerPathFilterFlags(command.Flags(), &pathFilters)
	if err := registerKindFlagCompletion(command, "kind"); err != nil {
		panic(err)
	}
	_ = command.MarkFlagRequired("name")

	return command
}
