package cmd

import (
	"github.com/XDwanj/gx/internal/lang"

	"github.com/spf13/cobra"
)

func newLangCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "lang",
		Short: "Manage language grammars",
	}

	command.AddCommand(
		&cobra.Command{
			Use:   "add <languages...>",
			Short: "Install language grammars",
			RunE: func(cmd *cobra.Command, args []string) error {
				return lang.Add(cmd.OutOrStdout(), cmd.ErrOrStderr(), args)
			},
		},
		&cobra.Command{
			Use:   "remove <languages...>",
			Short: "Remove installed language grammars",
			RunE: func(cmd *cobra.Command, args []string) error {
				return lang.Remove(cmd.OutOrStdout(), cmd.ErrOrStderr(), args)
			},
		},
		&cobra.Command{
			Use:   "list",
			Short: "List supported languages and install status",
			RunE: func(cmd *cobra.Command, _ []string) error {
				return lang.List(cmd.OutOrStdout(), cmd.ErrOrStderr())
			},
		},
	)

	return command
}
