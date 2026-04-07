package query

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
)

func displayPath(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

func displayScopePath(path string) string {
	if path == "" {
		return "."
	}
	return displayPath(path)
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
	patterns, err := expandNamePatterns(pattern)
	if err != nil {
		return false, err
	}
	for _, item := range patterns {
		matcher, compileErr := glob.Compile(item)
		if compileErr != nil {
			return false, compileErr
		}
		if matcher.Match(text) {
			return true, nil
		}
	}
	return false, nil
}

func expandNamePatterns(pattern string) ([]string, error) {
	parts := strings.Split(pattern, "|")
	if len(parts) == 1 {
		return parts, nil
	}

	patterns := make([]string, 0, len(parts))
	for _, item := range parts {
		candidate := strings.TrimSpace(item)
		if candidate == "" {
			return nil, fmt.Errorf("gx: invalid name pattern %q: empty alternative", pattern)
		}
		patterns = append(patterns, candidate)
	}
	return patterns, nil
}
