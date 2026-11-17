package app

import (
	"os"
	"path/filepath"

	"github.com/XDwanj/gx/internal/util/git"
)

type Flags struct {
	Directory string
	JSON      bool
	Verbose   bool
	Limit     string
	Offset    int
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
