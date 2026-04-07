package cmd

import (
	"github.com/XDwanj/gx/internal/lang"

	"github.com/spf13/cobra"
)

func newLangCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "lang",
		Short: "Enable or disable language grammars",
	}

	command.AddCommand(
		&cobra.Command{
			Use:   "enable <languages...>",
			Short: "Enable language grammars",
			RunE: func(cmd *cobra.Command, args []string) error {
				return lang.Add(cmd.OutOrStdout(), cmd.ErrOrStderr(), args)
			},
		},
		&cobra.Command{
			Use:   "disable <languages...>",
			Short: "Disable language grammars",
			RunE: func(cmd *cobra.Command, args []string) error {
				return lang.Remove(cmd.OutOrStdout(), cmd.ErrOrStderr(), args)
			},
		},
		&cobra.Command{
			Use:   "list",
			Short: "List supported languages and enable status",
			RunE: func(cmd *cobra.Command, _ []string) error {
				return lang.List(cmd.OutOrStdout(), cmd.ErrOrStderr())
			},
		},
	)

	return command
}
