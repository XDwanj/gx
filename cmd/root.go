package cmd

import (
	"fmt"
	"gx/internal/app"
	"gx/internal/query"
	"path/filepath"

	"github.com/spf13/cobra"
)

var rootFlags app.Flags

var rootCmd = &cobra.Command{
	Use:   "gx",
	Short: "Semantic code navigation for AI agents",
}

func Execute() int {
	if err := rootCmd.Execute(); err != nil {
		if err.Error() != "" {
			fmt.Fprintln(rootCmd.ErrOrStderr(), err)
		}
		return 1
	}
	return 0
}

func init() {
	rootCmd.SilenceUsage = true
	rootCmd.SetOut(rootCmd.OutOrStdout())
	rootCmd.SetErr(rootCmd.ErrOrStderr())

	rootCmd.PersistentFlags().StringVar(&rootFlags.Root, "root", "", "Project root (default: git root from cwd, then cwd)")
	rootCmd.PersistentFlags().BoolVar(&rootFlags.JSON, "json", false, "Emit JSON instead of TOON")

	rootCmd.AddCommand(
		newOverviewCmd(),
		newSymbolsCmd(),
		newDefinitionCmd(),
		newReferencesCmd(),
		newLangCmd(),
		newSkillCmd(),
		newCacheCmd(),
	)
}

func resolveRoot() (string, error) {
	return app.ResolveRoot(rootFlags.Root)
}

func resolveTargetPath(path string, root string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}

func buildRuntime() (*query.Runtime, error) {
	root, err := resolveRoot()
	if err != nil {
		return nil, err
	}
	return query.NewRuntime(root, rootFlags.JSON), nil
}
