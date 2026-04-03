package cmd

import (
	"fmt"
	"gx/internal/app"
	"gx/internal/index"
	"gx/internal/query"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	rootFlags   app.Flags
	showVersion bool
)

var rootCmd *cobra.Command

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
	rootCmd = newRootCmd()
}

func newRootCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "gx",
		Short: "Semantic code navigation for AI agents",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if showVersion {
				return app.PrintVersion(cmd.OutOrStdout(), rootFlags.JSON)
			}
			return cmd.Help()
		},
	}

	command.SilenceUsage = true
	command.SetOut(command.OutOrStdout())
	command.SetErr(command.ErrOrStderr())

	command.PersistentFlags().StringVar(&rootFlags.Root, "root", "", "Project root (default: git root from cwd, then cwd)")
	command.PersistentFlags().BoolVar(&rootFlags.JSON, "json", false, "Emit JSON instead of TOON")
	command.PersistentFlags().BoolVar(&rootFlags.Verbose, "verbose", false, "Emit debug progress to stderr")

	command.Flags().BoolVar(&showVersion, "version", false, "Print gx version and exit")
	command.Flags().BoolVarP(&showVersion, "version-lower-short", "v", false, "Print gx version and exit")
	command.Flags().BoolVarP(&showVersion, "version-upper-short", "V", false, "Print gx version and exit")
	_ = command.Flags().MarkHidden("version-lower-short")
	_ = command.Flags().MarkHidden("version-upper-short")

	command.AddCommand(
		newOverviewCmd(),
		newSymbolsCmd(),
		newDefinitionCmd(),
		newReferencesCmd(),
		newLangCmd(),
		newVersionCmd(),
		newSkillCmd(),
		newCacheCmd(),
	)

	return command
}

func resolveRoot() (string, error) {
	return app.ResolveRoot(rootFlags.Root)
}

func resolveTargetPaths(paths []string) ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	if len(paths) == 0 {
		return []string{cwd}, nil
	}

	resolved := make([]string, 0, len(paths))
	for _, path := range paths {
		if filepath.IsAbs(path) {
			resolved = append(resolved, filepath.Clean(path))
			continue
		}
		resolved = append(resolved, filepath.Clean(filepath.Join(cwd, path)))
	}
	return resolved, nil
}

func buildRuntime() (*query.Runtime, error) {
	root, err := resolveRoot()
	if err != nil {
		return nil, err
	}
	debugf(rootCmd.ErrOrStderr(), "resolved root %s", root)
	return query.NewRuntime(root, rootFlags.JSON, rootFlags.Verbose), nil
}

func loadIndex(root string, stderr io.Writer) (*index.Index, error) {
	return index.LoadOrBuildWithOptions(root, index.LoadOptions{
		Verbose: rootFlags.Verbose,
		Stderr:  stderr,
	})
}

func debugf(stderr io.Writer, format string, args ...any) {
	if !rootFlags.Verbose {
		return
	}
	_, _ = fmt.Fprintf(stderr, "gx: debug: "+format+"\n", args...)
}
