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
	Limit   int
	Offset  int
	All     bool
}

func ResolveRoot(root string) (string, error) {
	if root != "" {
		absoluteRoot, err := filepath.Abs(root)
		if err != nil {
			return "", err
		}
		realRoot, err := filepath.EvalSymlinks(absoluteRoot)
		if err != nil {
			return "", err
		}
		return filepath.Clean(realRoot), nil
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
