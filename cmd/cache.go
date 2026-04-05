package cmd

import (
	"os"

	"github.com/XDwanj/gx/internal/index"

	"github.com/spf13/cobra"
)

func newCacheCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "cache",
		Short: "Manage the index cache",
	}

	command.AddCommand(
		&cobra.Command{
			Use:   "path",
			Short: "Print the index cache path for the current project",
			RunE: func(cmd *cobra.Command, _ []string) error {
				root, err := resolveRoot()
				if err != nil {
					return err
				}
				_, err = cmd.OutOrStdout().Write([]byte(index.CachePathFor(root) + "\n"))
				return err
			},
		},
		&cobra.Command{
			Use:   "clean",
			Short: "Delete the cached index for the current project",
			RunE: func(cmd *cobra.Command, _ []string) error {
				root, err := resolveRoot()
				if err != nil {
					return err
				}
				cachePath := index.CachePathFor(root)
				if _, statErr := os.Stat(cachePath); statErr == nil {
					if removeErr := os.Remove(cachePath); removeErr != nil {
						return removeErr
					}
					_, err = cmd.ErrOrStderr().Write([]byte("gx: removed " + cachePath + "\n"))
					return err
				}
				_, err = cmd.ErrOrStderr().Write([]byte("gx: no cached index for this project\n"))
				return err
			},
		},
	)

	return command
}
