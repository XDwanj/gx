package query

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
)

type compiledGlob struct {
	raw     string
	matcher glob.Glob
}

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

func compilePathGlob(pattern string) (glob.Glob, error) {
	return glob.Compile(displayPath(filepath.Clean(pattern)), '/')
}

func compilePathGlobs(patterns []string, root string) ([]compiledGlob, error) {
	compiled := make([]compiledGlob, 0, len(patterns))
	for _, pattern := range patterns {
		normalized := normalizeGlobPattern(pattern, root)
		if normalized == "" {
			return nil, fmt.Errorf("gx: empty path glob")
		}
		matcher, err := compilePathGlob(normalized)
		if err != nil {
			return nil, err
		}
		compiled = append(compiled, compiledGlob{
			raw:     normalized,
			matcher: matcher,
		})
	}
	return compiled, nil
}

func anyPathGlobMatches(matchers []compiledGlob, path string) bool {
	if len(matchers) == 0 {
		return false
	}
	display := displayPath(path)
	for _, matcher := range matchers {
		if matcher.matcher.Match(display) {
			return true
		}
	}
	return false
}

func hasGlobMeta(path string) bool {
	return strings.ContainsAny(path, "*?[{")
}

func normalizeGlobPattern(pattern string, root string) string {
	trimmed := strings.TrimSpace(pattern)
	if trimmed == "" {
		return ""
	}
	return displayPath(normalizeRelativePath(trimmed, root))
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
