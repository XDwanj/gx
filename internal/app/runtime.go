package app

import (
	"gx/internal/query"
	"gx/internal/util/git"
	"os"
	"path/filepath"
)

type Flags struct {
	Root    string
	JSON    bool
	Verbose bool
}

func ResolveRoot(root string) (string, error) {
	if root != "" {
		return filepath.Clean(root), nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	return git.FindProjectRoot(cwd), nil
}

func NewQuery(root string, json bool, verbose bool) *query.Runtime {
	return query.NewRuntime(root, json, verbose)
}
