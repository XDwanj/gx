package query

import (
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
)

func displayPath(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

func normalizeRelativePath(path string, root string) string {
	if filepath.IsAbs(path) {
		rel, err := filepath.Rel(root, path)
		if err == nil {
			return filepath.Clean(rel)
		}
	}
	return filepath.Clean(path)
}

func globMatch(pattern string, text string) (bool, error) {
	matcher, err := glob.Compile(pattern)
	if err != nil {
		return false, err
	}
	return matcher.Match(text), nil
}
