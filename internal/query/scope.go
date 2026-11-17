package query

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/XDwanj/gx/internal/index"
	"github.com/XDwanj/gx/internal/language"
)

func (service *Service) resolveQueryContext(idx *index.Index, scope Scope) (*index.Index, *scopeFilter, error) {
	preparedIdx, matchers, err := preparePathQueryIndex(idx, scope)
	if err != nil {
		return nil, nil, err
	}
	filter, err := service.resolvePaths(preparedIdx, scope, matchers)
	if err != nil {
		return nil, nil, err
	}
	return preparedIdx, filter, nil
}

func preparePathQueryIndex(idx *index.Index, scope Scope) (*index.Index, pathFilterMatchers, error) {
	includeMatchers, err := compilePathGlobs(scope.Include, idx.Root)
	if err != nil {
		return nil, pathFilterMatchers{}, err
	}
	excludeMatchers, err := compilePathGlobs(scope.Exclude, idx.Root)
	if err != nil {
		return nil, pathFilterMatchers{}, err
	}

	preparedIdx, err := withIncludedEntries(idx, includeMatchers)
	if err != nil {
		return nil, pathFilterMatchers{}, err
	}

	return preparedIdx, pathFilterMatchers{
		includes: includeMatchers,
		excludes: excludeMatchers,
		ignore:   index.NewIgnoreMatcher(idx.Root),
	}, nil
}

func withIncludedEntries(idx *index.Index, includes []compiledGlob) (*index.Index, error) {
	if len(includes) == 0 {
		return idx, nil
	}

	preparedIdx := idx
	err := filepath.Walk(idx.Root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if path == idx.Root {
			return nil
		}
		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		relPath, err := filepath.Rel(idx.Root, path)
		if err != nil {
			return nil
		}
		relPath = filepath.Clean(relPath)
		if !anyPathGlobMatches(includes, relPath) {
			return nil
		}
		if _, ok := preparedIdx.Entries[relPath]; ok {
			return nil
		}

		data, indexable, err := index.LoadFileData(idx.Root, relPath)
		if err != nil {
			return err
		}
		if !indexable {
			return nil
		}

		if preparedIdx == idx {
			clonedEntries := make(map[string]index.FileData, len(idx.Entries)+1)
			for existingPath, existingData := range idx.Entries {
				clonedEntries[existingPath] = existingData
			}
			preparedIdx = &index.Index{
				Root:    idx.Root,
				Entries: clonedEntries,
			}
		}
		preparedIdx.Entries[relPath] = data
		return nil
	})
	if err != nil {
		return nil, err
	}

	return preparedIdx, nil
}

func (service *Service) resolvePaths(idx *index.Index, scope Scope, matchers pathFilterMatchers) (*scopeFilter, error) {
	filter := &scopeFilter{
		isSingleFile: len(scope.Targets) == 1,
		paths:        make(map[string]struct{}),
	}

	if len(scope.Targets) == 0 {
		for path := range idx.Entries {
			filter.paths[path] = struct{}{}
		}
		return applyPathGlobs(filter, matchers.includes, matchers.excludes, matchers.ignore), nil
	}

	missingPaths := make([]string, 0)
	for _, target := range scope.Targets {
		relPath := normalizeRelativePath(target, idx.Root)
		absPath := filepath.Join(idx.Root, relPath)

		info, statErr := os.Stat(absPath)
		if statErr == nil {
			if info.IsDir() {
				filter.isSingleFile = false
				prefix := relPath
				if prefix == "." {
					prefix = ""
				}

				for entryPath := range idx.Entries {
					if prefix == "" || entryPath == prefix || strings.HasPrefix(entryPath, prefix+string(filepath.Separator)) {
						filter.paths[entryPath] = struct{}{}
					}
				}
				continue
			}

			if _, ok := idx.Entries[relPath]; !ok {
				return nil, service.fileLookupError(relPath, idx.Root)
			}
			filter.paths[relPath] = struct{}{}
			continue
		}

		if !hasGlobMeta(relPath) {
			missingPaths = append(missingPaths, relPath)
			continue
		}

		matcher, matchErr := compilePathGlob(relPath)
		if matchErr != nil {
			return nil, matchErr
		}

		matched := false
		for entryPath := range idx.Entries {
			if !matcher.Match(displayPath(entryPath)) {
				continue
			}
			filter.paths[entryPath] = struct{}{}
			matched = true
		}
		if !matched {
			return nil, fmt.Errorf("gx: no indexed files match %s", displayScopePath(relPath))
		}
		filter.isSingleFile = false
	}

	if len(missingPaths) > 0 {
		return nil, missingPathsError(missingPaths)
	}

	return applyPathGlobs(filter, matchers.includes, matchers.excludes, matchers.ignore), nil
}

func (filter *scopeFilter) contains(path string) bool {
	if filter == nil {
		return true
	}
	_, ok := filter.paths[path]
	return ok
}

func applyPathGlobs(filter *scopeFilter, includes []compiledGlob, excludes []compiledGlob, ignoreMatcher *index.IgnoreMatcher) *scopeFilter {
	if filter == nil {
		return nil
	}
	if len(includes) == 0 && len(excludes) == 0 && ignoreMatcher == nil {
		return filter
	}

	filtered := &scopeFilter{
		isSingleFile: filter.isSingleFile,
		paths:        make(map[string]struct{}),
	}
	for path := range filter.paths {
		included := anyPathGlobMatches(includes, path)
		if len(includes) > 0 && !included {
			continue
		}
		if ignoreMatcher != nil && ignoreMatcher.Matches(path, false) && !included {
			continue
		}
		if anyPathGlobMatches(excludes, path) {
			continue
		}
		filtered.paths[path] = struct{}{}
	}
	if len(filtered.paths) != 1 {
		filtered.isSingleFile = false
	}
	return filtered
}

func (service *Service) fileLookupError(relPath string, root string) error {
	absPath := filepath.Join(root, relPath)
	if info, err := os.Stat(absPath); err == nil && !info.IsDir() {
		if language.DetectLanguage(absPath) == "" && strings.TrimPrefix(filepath.Ext(absPath), ".") != "" {
			return fmt.Errorf("gx: unsupported file type: .%s", strings.TrimPrefix(filepath.Ext(absPath), "."))
		}

		source, readErr := os.ReadFile(absPath)
		if readErr == nil && language.DetectLanguageFromSource(absPath, source) == "" {
			return fmt.Errorf("gx: unsupported file type: .%s", strings.TrimPrefix(filepath.Ext(absPath), "."))
		}
	}
	return fmt.Errorf("gx: file not in index: %s", displayPath(relPath))
}

func missingPathsError(paths []string) error {
	if len(paths) == 1 {
		return fmt.Errorf("gx: path not found: %s", displayScopePath(paths[0]))
	}

	displayPaths := make([]string, 0, len(paths))
	for _, path := range paths {
		displayPaths = append(displayPaths, displayScopePath(path))
	}
	return fmt.Errorf("gx: paths not found: %s", strings.Join(displayPaths, ", "))
}
