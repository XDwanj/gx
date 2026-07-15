package cmd

import (
	"github.com/XDwanj/gx/internal/app"

	"github.com/spf13/cobra"
)

func newIndexCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "index",
		Short: "Build or refresh the project index",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := resolveRoot()
			if err != nil {
				return err
			}
			return app.Preindex(
				root,
				rootFlags.JSON,
				rootFlags.Verbose,
				cmd.OutOrStdout(),
				cmd.ErrOrStderr(),
			)
		},
	}
}
