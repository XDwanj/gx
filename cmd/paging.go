package cmd

import (
	"fmt"
	"strconv"

	"github.com/XDwanj/gx/internal/query"

	"github.com/spf13/cobra"
)

const (
	defaultDefinitionLimit        = 5
	defaultSymbolsLimit           = 100
	defaultReferencesLimit        = 50
	defaultDirectoryOverviewLimit = 0
)

func resolvePageRequest(cmd *cobra.Command, defaultLimit int) (query.Page, error) {
	if rootFlags.Offset < 0 {
		return query.Page{}, fmt.Errorf("gx: --offset must be greater than or equal to 0")
	}

	if !pagingFlagChanged(cmd, "limit") || rootFlags.Limit == "" {
		return query.Page{Limit: defaultLimit, Offset: rootFlags.Offset}, nil
	}

	if rootFlags.Limit == "all" {
		return query.Page{Limit: 0, Offset: rootFlags.Offset}, nil
	}

	limit, err := strconv.Atoi(rootFlags.Limit)
	if err != nil || limit <= 0 {
		return query.Page{}, fmt.Errorf("gx: --limit must be a positive number or 'all'")
	}
	return query.Page{Limit: limit, Offset: rootFlags.Offset}, nil
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
