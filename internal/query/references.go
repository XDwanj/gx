package query

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/XDwanj/gx/internal/index"
	"github.com/XDwanj/gx/internal/language"
)

func (service *Service) References(idx *index.Index, intent Intent, scope Scope, page Page, unique bool) error {
	if err := validateDecideIntent(intent); err != nil {
		return err
	}
	resolvedIdx, filter, err := service.resolveQueryContext(idx, scope)
	if err != nil {
		return err
	}

	matchedNames, err := findReferenceNames(resolvedIdx, intent.Name, filter)
	if err != nil {
		return err
	}
	if len(matchedNames) == 0 {
		return service.writeEmptyPage(page)
	}

	rows := make([]ReferenceRow, 0)
	nameBytes := make([][]byte, 0, len(matchedNames))
	for _, matchedName := range matchedNames {
		nameBytes = append(nameBytes, []byte(matchedName))
	}
	for path, data := range resolvedIdx.Entries {
		if filter != nil && !filter.contains(path) {
			continue
		}

		source, err := os.ReadFile(filepath.Join(resolvedIdx.Root, path))
		if err != nil || !containsAnyName(source, nameBytes) {
			continue
		}

		lines := splitLines(source)
		references, findErr := language.FindReferencesForNames(
			data.Meta.Language,
			source,
			filepath.Join(resolvedIdx.Root, path),
			matchedNames,
		)
		if findErr != nil {
			if language.IsNotInstalled(findErr) {
				return findErr
			}
			continue
		}

		for _, reference := range references {
			context := ""
			if reference.Line-1 >= 0 && reference.Line-1 < len(lines) {
				context = strings.TrimSpace(lines[reference.Line-1])
			}
			rows = append(rows, ReferenceRow{
				File:    displayPath(path),
				Line:    reference.Line,
				Caller:  findEnclosingSymbol(data.Symbols, reference.ByteOffset),
				Context: context,
			})
		}
	}

	if len(rows) == 0 {
		return service.writeEmptyPage(page)
	}

	sort.Slice(rows, func(left int, right int) bool {
		if rows[left].File == rows[right].File {
			return rows[left].Line < rows[right].Line
		}
		return rows[left].File < rows[right].File
	})

	deduped := make([]ReferenceRow, 0, len(rows))
	for _, row := range rows {
		if len(deduped) > 0 {
			last := deduped[len(deduped)-1]
			if last.File == row.File && last.Line == row.Line {
				continue
			}
		}
		deduped = append(deduped, row)
	}

	if intent.DefinedIn != "" {
		target, err := service.resolveTarget(resolvedIdx, intent)
		if err != nil {
			return err
		}
		deduped, err = service.filterReferenceRowsWithDecide(resolvedIdx, intent.Name, target, deduped)
		if err != nil {
			return err
		}
		if len(deduped) == 0 {
			return service.writeEmptyPage(page)
		}
	}

	if unique {
		seen := map[string]bool{}
		uniqueRows := make([]UniqueCallerRow, 0)
		for _, row := range deduped {
			if row.Caller == "" {
				continue
			}
			key := row.File + ":" + row.Caller
			if seen[key] {
				continue
			}
			seen[key] = true
			uniqueRows = append(uniqueRows, UniqueCallerRow{
				File:   row.File,
				Caller: row.Caller,
				Line:   row.Line,
			})
		}
		if len(uniqueRows) == 0 {
			_, _ = fmt.Fprintln(service.stderr, "gx: no callers found")
			return service.writeEmptyPage(page)
		}
		uniqueRows, info := paginateRows(uniqueRows, page)
		return service.writeRows(uniqueRows, info)
	}

	deduped, info := paginateRows(deduped, page)
	return service.writeRows(deduped, info)
}

func findMatchingSymbolNames(idx *index.Index, nameGlob string) ([]string, error) {
	names := make([]string, 0)
	seen := make(map[string]struct{})
	for _, data := range idx.Entries {
		for _, symbol := range data.Symbols {
			ok, err := globMatch(nameGlob, symbol.Name)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
			if _, exists := seen[symbol.Name]; exists {
				continue
			}
			seen[symbol.Name] = struct{}{}
			names = append(names, symbol.Name)
		}
	}
	sort.Strings(names)
	return names, nil
}

func findReferenceNames(idx *index.Index, nameGlob string, filter *scopeFilter) ([]string, error) {
	names, err := findMatchingSymbolNames(idx, nameGlob)
	if err != nil {
		return nil, err
	}
	if len(names) > 0 {
		return names, nil
	}
	return findMatchingReferenceNames(idx, nameGlob, filter)
}

func findMatchingReferenceNames(idx *index.Index, nameGlob string, filter *scopeFilter) ([]string, error) {
	names := make([]string, 0)
	seen := make(map[string]struct{})
	for path, data := range idx.Entries {
		if filter != nil && !filter.contains(path) {
			continue
		}

		source, err := os.ReadFile(filepath.Join(idx.Root, path))
		if err != nil {
			continue
		}

		referenceNames, err := language.FindReferenceNames(data.Meta.Language, source, filepath.Join(idx.Root, path))
		if err != nil {
			if language.IsNotInstalled(err) {
				return nil, err
			}
			continue
		}

		for _, name := range referenceNames {
			ok, err := globMatch(nameGlob, name)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
			if _, exists := seen[name]; exists {
				continue
			}
			seen[name] = struct{}{}
			names = append(names, name)
		}
	}

	sort.Strings(names)
	return names, nil
}

func containsAnyName(source []byte, names [][]byte) bool {
	for _, name := range names {
		if bytes.Contains(source, name) {
			return true
		}
	}
	return false
}
