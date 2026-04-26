package cmd

import (
	"fmt"

	"github.com/XDwanj/gx/internal/query"

	"github.com/spf13/cobra"
)

const defaultTreeDepth = 8

func newTreeCmd() *cobra.Command {
	var defineIn string
	var direction string
	var depth int
	var pathFilters pathFilterFlags

	command := &cobra.Command{
		Use:   "tree [flags] [path ...]",
		Short: "Find AI-pruned in/out trees for a function",
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime, err := buildRuntime()
			if err != nil {
				return err
			}

			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}
			if defineIn == "" {
				return fmt.Errorf("gx: tree requires --define-in")
			}
			if depth < 0 {
				return fmt.Errorf("gx: --depth must be greater than or equal to 0")
			}
			if rootFlags.All || rootFlags.Offset != 0 || pagingFlagChanged(cmd, "limit") {
				return fmt.Errorf("gx: tree does not support pagination; use --depth to control output size")
			}

			paths, err := resolveTargetPaths(runtime.Root, args)
			if err != nil {
				return err
			}
			resolvedDefineIn, err := resolveDefineInPath(runtime.Root, defineIn)
			if err != nil {
				return err
			}

			debugf(rootCmd.ErrOrStderr(), "tree name=%s paths=%v define_in=%s direction=%s depth=%d", name, paths, resolvedDefineIn, direction, depth)

			idx, err := loadIndex(runtime.Root, rootCmd.ErrOrStderr())
			if err != nil {
				return err
			}

			return runtime.Query.Tree(idx, query.TreeOptions{
				Paths:     pathFilters.query(paths),
				NameGlob:  name,
				Direction: direction,
				Depth:     depth,
				AI:        query.AIOptions{DefineIn: resolvedDefineIn},
			})
		},
	}

	command.Flags().String("name", "", "Glob pattern to match the root function name")
	command.Flags().StringVar(&direction, "direction", query.TreeDirectionBoth, "Tree direction: in, out, or both")
	command.Flags().IntVar(&depth, "depth", defaultTreeDepth, "Maximum tree depth")
	registerDefineInFlag(command.Flags(), &defineIn)
	registerPathFilterFlags(command.Flags(), &pathFilters)
	_ = command.MarkFlagRequired("name")
	return command
}
