package query

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/XDwanj/gx/internal/index"
	"github.com/XDwanj/gx/internal/language"
	"github.com/XDwanj/gx/internal/output"
)

const (
	directoryOverviewMaxSymbols = 10
	symbolPriorityPrimary       = iota
	symbolPrioritySecondary
	symbolPriorityFallback
)

const (
	definitionPriorityTypes = iota
	definitionPriorityCallables
	definitionPriorityFallback
)

type Runtime struct {
	Root    string
	JSON    bool
	Verbose bool
	Query   *Service
}

func NewRuntime(root string, json bool, verbose bool) *Runtime {
	service := &Service{
		root:    filepath.Clean(root),
		stdout:  os.Stdout,
		stderr:  os.Stderr,
		json:    json,
		verbose: verbose,
	}
	return &Runtime{
		Root:    filepath.Clean(root),
		JSON:    json,
		Verbose: verbose,
		Query:   service,
	}
}

type Service struct {
	root    string
	stdout  io.Writer
	stderr  io.Writer
	json    bool
	verbose bool
}

type SymbolRow struct {
	File      string `json:"file"`
	Line      int    `json:"line"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Signature string `json:"signature"`
}

type DefinitionResult struct {
	File      string `json:"file"`
	Line      int    `json:"line"`
	Truncated bool   `json:"truncated,omitempty"`
	Lines     int    `json:"lines,omitempty"`
	Body      string `json:"body"`
}

type ReferenceRow struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Caller  string `json:"caller,omitempty"`
	Context string `json:"context"`
}

type CalleeRow struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Caller  string `json:"caller"`
	Callee  string `json:"callee"`
	Context string `json:"context"`
}

type UniqueCallerRow struct {
	File   string `json:"file"`
	Caller string `json:"caller"`
	Line   int    `json:"line"`
}

type definitionMatch struct {
	path   string
	symbol index.Symbol
}

type scopeFilter struct {
	isSingleFile bool
	paths        map[string]struct{}
}

type DirOverviewRow struct {
	File    string `json:"file"`
	Symbols string `json:"symbols"`
}

type DirOverviewFullRow struct {
	File      string `json:"file"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Signature string `json:"signature"`
}

type OverviewSection struct {
	Target     string `json:"target"`
	TargetKind string `json:"target_kind"`
	Rows       any    `json:"rows"`
}

type directoryOverviewEntry struct {
	path string
	data index.FileData
}

type directoryOverviewData struct {
	relDir      string
	directFiles []directoryOverviewEntry
	subdirs     map[string][2]int
}

type overviewTargetKind string

const (
	overviewTargetDirectory overviewTargetKind = "directory"
	overviewTargetFile      overviewTargetKind = "file"
	overviewTargetMarkdown  overviewTargetKind = "markdown"
)

type PageRequest struct {
	Limit  int
	Offset int
}

type pageInfo struct {
	Total      int
	Offset     int
	Returned   int
	NextOffset int
	Truncated  bool
	OutOfRange bool
}

func (service *Service) Symbols(idx *index.Index, paths []string, nameGlob *string, kind *index.SymbolKind, page PageRequest) error {
	rows, err := service.symbolRows(idx, paths, nameGlob, kind)
	if err != nil {
		return err
	}

	if len(rows) == 0 {
		_, _ = fmt.Fprintln(service.stderr, "gx: no matches")
		return nil
	}

	sort.Slice(rows, func(left int, right int) bool {
		if rows[left].File == rows[right].File {
			if rows[left].Line == rows[right].Line {
				return rows[left].Name < rows[right].Name
			}
			return rows[left].Line < rows[right].Line
		}
		return rows[left].File < rows[right].File
	})

	rows, pageState := paginateRows(rows, page)
	service.writePageHint(pageState)

	if service.json {
		return output.PrintJSON(service.stdout, rows)
	}
	return output.PrintTOON(service.stdout, rows)
}

func (service *Service) Overview(idx *index.Index, paths []string, full bool, page PageRequest) error {
	if len(paths) <= 1 {
		if len(paths) == 0 {
			return service.Symbols(idx, paths, nil, nil, page)
		}
		return service.singleOverview(idx, paths[0], full, page)
	}

	sections := make([]OverviewSection, 0, len(paths))
	for _, target := range paths {
		section, empty, err := service.overviewSectionForTarget(idx, target, full, page)
		if err != nil {
			return err
		}
		if empty {
			continue
		}
		sections = append(sections, section)
	}

	if len(sections) == 0 {
		_, _ = fmt.Fprintln(service.stderr, "gx: no matches")
		return nil
	}

	if service.json {
		return output.PrintJSON(service.stdout, sections)
	}
	return service.printOverviewSections(sections)
}

func (service *Service) singleOverview(idx *index.Index, target string, full bool, page PageRequest) error {
	switch classifyOverviewTarget(target) {
	case overviewTargetDirectory:
		return service.DirectoryOverview(idx, target, full, page)
	case overviewTargetMarkdown:
		return service.MarkdownOverview(target)
	default:
		return service.Symbols(idx, []string{target}, nil, nil, PageRequest{})
	}
}

func (service *Service) overviewSectionForTarget(idx *index.Index, target string, full bool, page PageRequest) (OverviewSection, bool, error) {
	targetLabel := displayPath(normalizeRelativePath(target, service.root))

	switch classifyOverviewTarget(target) {
	case overviewTargetDirectory:
		if idx == nil {
			return OverviewSection{}, false, fmt.Errorf("gx: overview index is required for directories")
		}
		if full {
			rows, err := service.directoryOverviewFullRows(idx, target)
			if err != nil {
				return OverviewSection{}, false, err
			}
			rows, pageState := paginateRows(rows, page)
			service.writeScopedPageHint(targetLabel, pageState)
			return OverviewSection{
				Target:     targetLabel,
				TargetKind: string(overviewTargetDirectory),
				Rows:       rows,
			}, len(rows) == 0, nil
		}

		rows, err := service.directoryOverviewRows(idx, target)
		if err != nil {
			return OverviewSection{}, false, err
		}
		rows, pageState := paginateRows(rows, page)
		service.writeScopedPageHint(targetLabel, pageState)
		return OverviewSection{
			Target:     targetLabel,
			TargetKind: string(overviewTargetDirectory),
			Rows:       rows,
		}, len(rows) == 0, nil
	case overviewTargetMarkdown:
		rows, err := service.markdownOverviewRows(target)
		if err != nil {
			return OverviewSection{}, false, err
		}
		if len(rows) == 0 {
			_, _ = fmt.Fprintf(service.stderr, "gx: no headings found in %s\n", targetLabel)
			return OverviewSection{}, true, nil
		}
		return OverviewSection{
			Target:     targetLabel,
			TargetKind: string(overviewTargetMarkdown),
			Rows:       rows,
		}, false, nil
	default:
		if idx == nil {
			return OverviewSection{}, false, fmt.Errorf("gx: overview index is required for files")
		}
		rows, err := service.symbolRows(idx, []string{target}, nil, nil)
		if err != nil {
			return OverviewSection{}, false, err
		}
		if len(rows) == 0 {
			_, _ = fmt.Fprintf(service.stderr, "gx: no matches in %s\n", targetLabel)
			return OverviewSection{}, true, nil
		}
		return OverviewSection{
			Target:     targetLabel,
			TargetKind: string(overviewTargetFile),
			Rows:       rows,
		}, false, nil
	}
}

func (service *Service) printOverviewSections(sections []OverviewSection) error {
	for indexValue, section := range sections {
		if indexValue > 0 {
			if _, err := fmt.Fprintln(service.stdout); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(service.stdout, "target: %s\ntarget_kind: %s\n", section.Target, section.TargetKind); err != nil {
			return err
		}

		var rendered bytes.Buffer
		if err := output.PrintTOON(&rendered, section.Rows); err != nil {
			return err
		}
		if _, err := io.WriteString(service.stdout, rendered.String()); err != nil {
			return err
		}
	}
	return nil
}

func classifyOverviewTarget(path string) overviewTargetKind {
	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		return overviewTargetDirectory
	}
	if err == nil && IsMarkdownPath(path) {
		return overviewTargetMarkdown
	}
	return overviewTargetFile
}

func (service *Service) symbolRows(idx *index.Index, paths []string, nameGlob *string, kind *index.SymbolKind) ([]SymbolRow, error) {
	filter, err := service.resolvePaths(idx, paths)
	if err != nil {
		return nil, err
	}

	rows := make([]SymbolRow, 0)
	for path, data := range idx.Entries {
		if filter != nil && !filter.contains(path) {
			continue
		}
		for _, symbol := range data.Symbols {
			if nameGlob != nil {
				ok, err := globMatch(*nameGlob, symbol.Name)
				if err != nil {
					return nil, err
				}
				if !ok {
					continue
				}
			}
			if kind != nil && symbol.Kind != *kind {
				continue
			}
			rows = append(rows, SymbolRow{
				File:      displayPath(path),
				Line:      symbol.Line,
				Name:      symbol.Name,
				Kind:      string(symbol.Kind),
				Signature: symbol.Signature,
			})
		}
	}

	sort.Slice(rows, func(left int, right int) bool {
		if rows[left].File == rows[right].File {
			if rows[left].Line == rows[right].Line {
				return rows[left].Name < rows[right].Name
			}
			return rows[left].Line < rows[right].Line
		}
		return rows[left].File < rows[right].File
	})

	return rows, nil
}

func (service *Service) Definition(idx *index.Index, nameGlob string, paths []string, kind *index.SymbolKind, maxLines int, page PageRequest) error {
	filter, err := service.resolvePaths(idx, paths)
	if err != nil {
		return err
	}

	matches, err := findDefinitionMatches(idx, nameGlob, kind)
	if err != nil {
		return err
	}

	if filter != nil {
		filtered := make([]definitionMatch, 0)
		for _, match := range matches {
			if filter.contains(match.path) {
				filtered = append(filtered, match)
			}
		}
		matches = filtered
	}

	if len(matches) == 0 {
		_, _ = fmt.Fprintln(service.stderr, "gx: no matches")
		return nil
	}

	sortDefinitionMatches(matches)
	matches, pageState := paginateRows(matches, page)
	service.writePageHint(pageState)

	results := make([]DefinitionResult, 0, len(matches))
	for _, match := range matches {
		body, startLine, ok := readBody(idx.Root, match.path, match.symbol)
		if !ok {
			body = ""
			startLine = 0
		}
		lineCount := 0
		if body != "" {
			lineCount = strings.Count(body, "\n") + 1
		}
		result := DefinitionResult{
			File: displayPath(match.path),
			Line: startLine,
			Body: body,
		}
		if lineCount > maxLines && maxLines > 0 {
			lines := strings.Split(body, "\n")
			result.Truncated = true
			result.Lines = lineCount
			result.Body = strings.Join(lines[:maxLines], "\n")
		}
		results = append(results, result)
	}

	if service.json {
		return output.PrintJSON(service.stdout, results)
	}

	for indexValue, result := range results {
		if indexValue > 0 {
			_, _ = fmt.Fprintln(service.stdout)
		}
		if _, err := fmt.Fprintf(service.stdout, "file: %s\nline: %d", result.File, result.Line); err != nil {
			return err
		}
		if result.Lines > 0 {
			shownLines := strings.Count(result.Body, "\n") + 1
			if _, err := fmt.Fprintf(
				service.stdout,
				"\ntruncated: showing first %d of %d lines; rerun with --max-lines %d to view the full body",
				shownLines,
				result.Lines,
				result.Lines,
			); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(service.stdout, "\n---\n%s\n", result.Body); err != nil {
			return err
		}
	}
	return nil
}

func (service *Service) References(idx *index.Index, nameGlob string, paths []string, unique bool, page PageRequest) error {
	filter, err := service.resolvePaths(idx, paths)
	if err != nil {
		return err
	}

	matchedNames, err := findReferenceNames(idx, nameGlob, filter)
	if err != nil {
		return err
	}
	if len(matchedNames) == 0 {
		_, _ = fmt.Fprintln(service.stderr, "gx: no matches")
		return nil
	}

	rows := make([]ReferenceRow, 0)
	nameBytes := make([][]byte, 0, len(matchedNames))
	for _, matchedName := range matchedNames {
		nameBytes = append(nameBytes, []byte(matchedName))
	}
	for path, data := range idx.Entries {
		if filter != nil && !filter.contains(path) {
			continue
		}

		source, err := os.ReadFile(filepath.Join(idx.Root, path))
		if err != nil || !containsAnyName(source, nameBytes) {
			continue
		}

		lines := splitLines(source)
		references, findErr := language.FindReferencesForNames(
			data.Meta.Language,
			source,
			filepath.Join(idx.Root, path),
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
		_, _ = fmt.Fprintln(service.stderr, "gx: no matches")
		return nil
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
			return nil
		}
		uniqueRows, pageState := paginateRows(uniqueRows, page)
		service.writePageHint(pageState)
		if service.json {
			return output.PrintJSON(service.stdout, uniqueRows)
		}
		return output.PrintTOON(service.stdout, uniqueRows)
	}

	deduped, pageState := paginateRows(deduped, page)
	service.writePageHint(pageState)
	if service.json {
		return output.PrintJSON(service.stdout, deduped)
	}
	return output.PrintTOON(service.stdout, deduped)
}

func (service *Service) Callees(idx *index.Index, nameGlob string, paths []string, page PageRequest) error {
	filter, err := service.resolvePaths(idx, paths)
	if err != nil {
		return err
	}

	callableKind := index.SymbolKindFunc
	matches, err := findDefinitionMatches(idx, nameGlob, &callableKind)
	if err != nil {
		return err
	}
	if filter != nil {
		filtered := make([]definitionMatch, 0, len(matches))
		for _, match := range matches {
			if filter.contains(match.path) {
				filtered = append(filtered, match)
			}
		}
		matches = filtered
	}
	if len(matches) == 0 {
		_, _ = fmt.Fprintln(service.stderr, "gx: no matches")
		return nil
	}

	rows := make([]CalleeRow, 0)
	sourceCache := make(map[string][]byte)
	linesCache := make(map[string][]string)
	for _, match := range matches {
		data, ok := idx.Entries[match.path]
		if !ok {
			continue
		}
		if !language.SupportsCallees(data.Meta.Language) {
			return fmt.Errorf("gx: callees not supported for language: %s", data.Meta.Language)
		}

		source, ok := sourceCache[match.path]
		if !ok {
			source, err = os.ReadFile(filepath.Join(idx.Root, match.path))
			if err != nil {
				return err
			}
			sourceCache[match.path] = source
			linesCache[match.path] = splitLines(source)
		}

		callees, err := language.FindCallees(
			data.Meta.Language,
			source,
			filepath.Join(idx.Root, match.path),
			match.symbol.ByteStart,
			match.symbol.ByteEnd,
		)
		if err != nil {
			return err
		}

		lines := linesCache[match.path]
		for _, callee := range callees {
			context := ""
			if callee.Line-1 >= 0 && callee.Line-1 < len(lines) {
				context = strings.TrimSpace(lines[callee.Line-1])
			}
			rows = append(rows, CalleeRow{
				File:    displayPath(match.path),
				Line:    callee.Line,
				Caller:  match.symbol.Name,
				Callee:  callee.Name,
				Context: context,
			})
		}
	}

	if len(rows) == 0 {
		_, _ = fmt.Fprintln(service.stderr, "gx: no matches")
		return nil
	}

	sort.Slice(rows, func(left int, right int) bool {
		if rows[left].File == rows[right].File {
			if rows[left].Line == rows[right].Line {
				if rows[left].Caller == rows[right].Caller {
					return rows[left].Callee < rows[right].Callee
				}
				return rows[left].Caller < rows[right].Caller
			}
			return rows[left].Line < rows[right].Line
		}
		return rows[left].File < rows[right].File
	})

	rows, pageState := paginateRows(rows, page)
	service.writePageHint(pageState)
	if service.json {
		return output.PrintJSON(service.stdout, rows)
	}
	return output.PrintTOON(service.stdout, rows)
}

func findDefinitionMatches(idx *index.Index, nameGlob string, kind *index.SymbolKind) ([]definitionMatch, error) {
	matches := make([]definitionMatch, 0)
	for path, data := range idx.Entries {
		for _, symbol := range data.Symbols {
			ok, err := globMatch(nameGlob, symbol.Name)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
			if kind != nil && symbol.Kind != *kind {
				continue
			}
			matches = append(matches, definitionMatch{path: path, symbol: symbol})
		}
	}
	return matches, nil
}

func sortDefinitionMatches(matches []definitionMatch) {
	sort.Slice(matches, func(left int, right int) bool {
		leftPriority := definitionSymbolPriority(matches[left].symbol.Kind)
		rightPriority := definitionSymbolPriority(matches[right].symbol.Kind)
		if leftPriority != rightPriority {
			return leftPriority < rightPriority
		}
		if matches[left].path == matches[right].path {
			if matches[left].symbol.Line == matches[right].symbol.Line {
				return matches[left].symbol.Name < matches[right].symbol.Name
			}
			return matches[left].symbol.Line < matches[right].symbol.Line
		}
		return matches[left].path < matches[right].path
	})
}

func definitionSymbolPriority(kind index.SymbolKind) int {
	switch kind {
	case index.SymbolKindStruct,
		index.SymbolKindEnum,
		index.SymbolKindInterface,
		index.SymbolKindClass,
		index.SymbolKindModule,
		index.SymbolKindType,
		index.SymbolKindConst:
		return definitionPriorityTypes
	case index.SymbolKindFunc:
		return definitionPriorityCallables
	default:
		return definitionPriorityFallback
	}
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

func (service *Service) DirectoryOverview(idx *index.Index, dir string, full bool, page PageRequest) error {
	if full {
		rows, err := service.directoryOverviewFullRows(idx, dir)
		if err != nil {
			return err
		}
		rows, pageState := paginateRows(rows, page)
		service.writePageHint(pageState)
		if service.json {
			return output.PrintJSON(service.stdout, rows)
		}
		return output.PrintTOON(service.stdout, rows)
	}

	rows, err := service.directoryOverviewRows(idx, dir)
	if err != nil {
		return err
	}
	rows, pageState := paginateRows(rows, page)
	service.writePageHint(pageState)
	if service.json {
		return output.PrintJSON(service.stdout, rows)
	}
	return output.PrintTOON(service.stdout, rows)
}

func (service *Service) directoryOverviewRows(idx *index.Index, dir string) ([]DirOverviewRow, error) {
	data, err := buildDirectoryOverviewData(idx, dir)
	if err != nil {
		return nil, err
	}

	rows := make([]DirOverviewRow, 0)
	for _, name := range sortedKeys(data.subdirs) {
		stats := data.subdirs[name]
		rows = append(rows, DirOverviewRow{
			File:    formatSubdir(data.relDir, name),
			Symbols: fmt.Sprintf("(%d files, %d symbols)", stats[0], stats[1]),
		})
	}

	for _, fileEntry := range data.directFiles {
		symbols := prepareSymbols(fileEntry.data.Symbols)
		if len(symbols) == 0 {
			continue
		}
		seen := map[string]bool{}
		names := make([]string, 0)
		for _, symbol := range symbols[:minInt(len(symbols), directoryOverviewMaxSymbols)] {
			if seen[symbol.Name] {
				continue
			}
			seen[symbol.Name] = true
			names = append(names, symbol.Name)
		}
		suffix := ""
		if len(symbols) > len(names) {
			suffix = fmt.Sprintf(", ... (+%d more)", len(symbols)-len(names))
		}
		rows = append(rows, DirOverviewRow{
			File:    displayPath(fileEntry.path),
			Symbols: strings.Join(names, ", ") + suffix,
		})
	}

	return rows, nil
}

func (service *Service) directoryOverviewFullRows(idx *index.Index, dir string) ([]DirOverviewFullRow, error) {
	data, err := buildDirectoryOverviewData(idx, dir)
	if err != nil {
		return nil, err
	}

	rows := make([]DirOverviewFullRow, 0)
	for _, name := range sortedKeys(data.subdirs) {
		stats := data.subdirs[name]
		rows = append(rows, DirOverviewFullRow{
			File:      formatSubdir(data.relDir, name),
			Name:      fmt.Sprintf("(%d files, %d symbols)", stats[0], stats[1]),
			Kind:      "",
			Signature: "",
		})
	}

	for _, fileEntry := range data.directFiles {
		symbols := prepareSymbols(fileEntry.data.Symbols)
		if len(symbols) == 0 {
			continue
		}
		total := len(symbols)
		for _, symbol := range symbols[:minInt(total, directoryOverviewMaxSymbols)] {
			rows = append(rows, DirOverviewFullRow{
				File:      displayPath(fileEntry.path),
				Name:      symbol.Name,
				Kind:      string(symbol.Kind),
				Signature: symbol.Signature,
			})
		}
		if total > directoryOverviewMaxSymbols {
			rows = append(rows, DirOverviewFullRow{
				File: displayPath(fileEntry.path),
				Name: fmt.Sprintf("... (+%d more)", total-directoryOverviewMaxSymbols),
			})
		}
	}

	return rows, nil
}

func buildDirectoryOverviewData(idx *index.Index, dir string) (directoryOverviewData, error) {
	relDir := normalizeRelativePath(dir, idx.Root)
	if relDir == "." {
		relDir = ""
	}

	allEntries := make([]directoryOverviewEntry, 0)
	for path, data := range idx.Entries {
		if relDir != "" && !strings.HasPrefix(path, relDir) {
			continue
		}
		if isTestFile(path) {
			continue
		}
		allEntries = append(allEntries, directoryOverviewEntry{path: path, data: data})
	}

	if len(allEntries) == 0 {
		return directoryOverviewData{}, fmt.Errorf("gx: no indexed files under %s", displayPath(relDir))
	}

	directFiles := make([]directoryOverviewEntry, 0)
	subdirs := map[string][2]int{}
	for _, item := range allEntries {
		child, ok := childComponent(item.path, relDir)
		if !ok {
			continue
		}
		nonTestCount := 0
		for _, symbol := range item.data.Symbols {
			if !symbol.IsTest {
				nonTestCount++
			}
		}

		if !strings.Contains(child, string(filepath.Separator)) && filepath.Ext(child) != "" {
			directFiles = append(directFiles, item)
			continue
		}
		stats := subdirs[child]
		stats[0]++
		stats[1] += nonTestCount
		subdirs[child] = stats
	}

	sort.Slice(directFiles, func(left int, right int) bool {
		return directFiles[left].path < directFiles[right].path
	})

	return directoryOverviewData{
		relDir:      relDir,
		directFiles: directFiles,
		subdirs:     subdirs,
	}, nil
}

func paginateRows[T any](rows []T, request PageRequest) ([]T, pageInfo) {
	info := pageInfo{
		Total:  len(rows),
		Offset: request.Offset,
	}
	if len(rows) == 0 {
		return rows, info
	}
	if request.Offset >= len(rows) {
		return rows[:0], pageInfo{
			Total:      len(rows),
			Offset:     request.Offset,
			OutOfRange: request.Offset > 0,
		}
	}

	paged := rows
	if request.Offset > 0 {
		paged = paged[request.Offset:]
	}
	if request.Limit > 0 && len(paged) > request.Limit {
		paged = paged[:request.Limit]
		info.Truncated = true
		info.NextOffset = request.Offset + len(paged)
	}
	info.Returned = len(paged)
	return paged, info
}

func (service *Service) writePageHint(info pageInfo) {
	service.writeScopedPageHint("", info)
}

func (service *Service) writeScopedPageHint(scope string, info pageInfo) {
	prefix := ""
	if scope != "" {
		prefix = scope + " "
	}

	switch {
	case info.OutOfRange:
		_, _ = fmt.Fprintf(service.stderr, "gx: %soffset %d exceeds %d results\n", prefix, info.Offset, info.Total)
	case info.Truncated:
		start := info.Offset + 1
		end := info.Offset + info.Returned
		_, _ = fmt.Fprintf(
			service.stderr,
			"gx: %sshowing %d-%d of %d; narrow query, use --offset %d, or --all\n",
			prefix,
			start,
			end,
			info.Total,
			info.NextOffset,
		)
	}
}

func (service *Service) fileLookupError(relPath string, root string) error {
	absPath := filepath.Join(root, relPath)
	if _, err := os.Stat(absPath); err == nil && language.DetectLanguage(absPath) == "" {
		return fmt.Errorf("gx: unsupported file type: .%s", strings.TrimPrefix(filepath.Ext(absPath), "."))
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

func (service *Service) resolvePaths(idx *index.Index, paths []string) (*scopeFilter, error) {
	if len(paths) == 0 {
		return nil, nil
	}

	missingPaths := make([]string, 0)
	for _, path := range paths {
		relPath := normalizeRelativePath(path, idx.Root)
		absPath := filepath.Join(idx.Root, relPath)
		if _, err := os.Stat(absPath); err != nil {
			missingPaths = append(missingPaths, relPath)
		}
	}
	if len(missingPaths) > 0 {
		return nil, missingPathsError(missingPaths)
	}

	filter := &scopeFilter{
		isSingleFile: len(paths) == 1,
		paths:        make(map[string]struct{}),
	}

	for _, path := range paths {
		relPath := normalizeRelativePath(path, idx.Root)
		absPath := filepath.Join(idx.Root, relPath)
		info, err := os.Stat(absPath)
		if err != nil {
			return nil, missingPathsError([]string{relPath})
		}

		if info.IsDir() {
			filter.isSingleFile = false
			prefix := relPath
			if prefix == "." {
				prefix = ""
			}

			matched := false
			for entryPath := range idx.Entries {
				if prefix == "" || entryPath == prefix || strings.HasPrefix(entryPath, prefix+string(filepath.Separator)) {
					filter.paths[entryPath] = struct{}{}
					matched = true
				}
			}
			if !matched {
				return nil, fmt.Errorf("gx: no indexed files under %s", displayScopePath(relPath))
			}
			continue
		}

		if _, ok := idx.Entries[relPath]; !ok {
			return nil, service.fileLookupError(relPath, idx.Root)
		}
		filter.paths[relPath] = struct{}{}
	}

	return filter, nil
}

func (filter *scopeFilter) contains(path string) bool {
	if filter == nil {
		return true
	}
	_, ok := filter.paths[path]
	return ok
}

func readBody(root string, file string, symbol index.Symbol) (string, int, bool) {
	source, err := os.ReadFile(filepath.Join(root, file))
	if err != nil {
		return "", 0, false
	}
	if int(symbol.ByteEnd) > len(source) {
		return "", 0, false
	}
	start := int(symbol.ByteStart)
	line := bytes.Count(source[:start], []byte{'\n'}) + 1
	return string(source[symbol.ByteStart:symbol.ByteEnd]), line, true
}

func findEnclosingSymbol(symbols []index.Symbol, byteOffset uint) string {
	result := ""
	bestWidth := ^uint(0)
	for _, symbol := range symbols {
		if symbol.ByteStart <= byteOffset && byteOffset < symbol.ByteEnd {
			width := symbol.ByteEnd - symbol.ByteStart
			if width < bestWidth {
				bestWidth = width
				result = symbol.Name
			}
		}
	}
	return result
}

func splitLines(source []byte) []string {
	text := strings.ReplaceAll(string(source), "\r\n", "\n")
	return strings.Split(text, "\n")
}

func prepareSymbols(symbols []index.Symbol) []index.Symbol {
	result := make([]index.Symbol, 0)
	for _, symbol := range symbols {
		if !symbol.IsTest {
			result = append(result, symbol)
		}
	}
	sort.Slice(result, func(left int, right int) bool {
		leftPriority := symbolPriority(result[left].Kind)
		rightPriority := symbolPriority(result[right].Kind)
		if leftPriority == rightPriority {
			return result[left].Name < result[right].Name
		}
		return leftPriority < rightPriority
	})
	return result
}

func symbolPriority(kind index.SymbolKind) int {
	switch kind {
	case index.SymbolKindStruct, index.SymbolKindEnum, index.SymbolKindInterface, index.SymbolKindClass:
		return symbolPriorityPrimary
	case index.SymbolKindFunc, index.SymbolKindConst, index.SymbolKindType, index.SymbolKindModule:
		return symbolPrioritySecondary
	default:
		return symbolPriorityFallback
	}
}

func isTestFile(path string) bool {
	for _, part := range strings.Split(displayPath(path), "/") {
		if part == "tests" || part == "test" || part == "__tests__" {
			return true
		}
	}
	name := filepath.Base(path)
	switch {
	case strings.HasSuffix(name, "_test.go"):
		return true
	case strings.HasSuffix(name, ".test.ts"), strings.HasSuffix(name, ".test.tsx"), strings.HasSuffix(name, ".test.js"), strings.HasSuffix(name, ".test.jsx"):
		return true
	case strings.HasSuffix(name, ".spec.ts"), strings.HasSuffix(name, ".spec.tsx"), strings.HasSuffix(name, ".spec.js"), strings.HasSuffix(name, ".spec.jsx"):
		return true
	case strings.HasPrefix(name, "test_") && strings.HasSuffix(name, ".py"):
		return true
	case strings.HasSuffix(name, "_spec.rb"):
		return true
	default:
		return false
	}
}

func childComponent(path string, dir string) (string, bool) {
	relative := path
	if dir != "" {
		var err error
		relative, err = filepath.Rel(dir, path)
		if err != nil {
			return "", false
		}
	}
	parts := strings.Split(displayPath(relative), "/")
	if len(parts) == 0 {
		return "", false
	}
	if len(parts) > 1 {
		return parts[0], true
	}
	return relative, true
}

func formatSubdir(relDir string, dirName string) string {
	if relDir == "" {
		return displayPath(dirName) + "/"
	}
	return displayPath(filepath.Join(relDir, dirName)) + "/"
}

func sortedKeys(values map[string][2]int) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func minInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}
