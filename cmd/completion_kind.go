package cmd

import (
	"gx/internal/index"

	"github.com/spf13/cobra"
)

func registerKindFlagCompletion(command *cobra.Command, flagName string) error {
	return command.RegisterFlagCompletionFunc(flagName, kindFlagCompletion)
}

func kindFlagCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	kinds := index.PublicSymbolKinds()
	values := make([]string, 0, len(kinds))
	for _, kind := range kinds {
		values = append(values, string(kind))
	}
	return values, cobra.ShellCompDirectiveNoFileComp
}
