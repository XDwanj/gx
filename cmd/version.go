package cmd

import (
	"github.com/XDwanj/gx/internal/app"

	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print gx version",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return app.PrintVersion(cmd.OutOrStdout(), rootFlags.JSON)
		},
	}
}
