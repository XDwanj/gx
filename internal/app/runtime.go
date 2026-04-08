package app

import (
	"os"
	"path/filepath"

	"github.com/XDwanj/gx/internal/query"
	"github.com/XDwanj/gx/internal/util/git"
)

type Flags struct {
	Directory string
	JSON      bool
	Verbose   bool
	Limit     int
	Offset    int
	All       bool
}

func ResolveRoot(directory string) (string, error) {
	if directory != "" {
		absoluteRoot, err := filepath.Abs(directory)
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
