package cmd

import (
	"fmt"

	"github.com/XDwanj/gx/internal/query"

	"github.com/spf13/cobra"
)

const (
	defaultDefinitionLimit        = 5
	defaultSymbolsLimit           = 100
	defaultReferencesLimit        = 50
	defaultDirectoryOverviewLimit = 0
)

func resolvePageRequest(cmd *cobra.Command, defaultLimit int) (query.PageRequest, error) {
	if rootFlags.Offset < 0 {
		return query.PageRequest{}, fmt.Errorf("gx: --offset must be greater than or equal to 0")
	}

	limitChanged := pagingFlagChanged(cmd, "limit")
	if limitChanged && rootFlags.Limit <= 0 {
		return query.PageRequest{}, fmt.Errorf("gx: --limit must be greater than 0")
	}

	if rootFlags.All {
		return query.PageRequest{Offset: rootFlags.Offset}, nil
	}

	limit := defaultLimit
	if limitChanged {
		limit = rootFlags.Limit
	}

	return query.PageRequest{
		Limit:  limit,
		Offset: rootFlags.Offset,
	}, nil
}

func pagingFlagChanged(cmd *cobra.Command, name string) bool {
	if cmd != nil {
		if flag := cmd.Flags().Lookup(name); flag != nil && flag.Changed {
			return true
		}
	}
	if rootCmd == nil {
		return false
	}
	flag := rootCmd.PersistentFlags().Lookup(name)
	return flag != nil && flag.Changed
}
