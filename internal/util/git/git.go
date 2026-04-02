package git

import (
	"os"
	"path/filepath"
)

func FindProjectRoot(start string) string {
	current := filepath.Clean(start)
	for {
		if exists(filepath.Join(current, ".git")) {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			return filepath.Clean(start)
		}
		current = parent
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
