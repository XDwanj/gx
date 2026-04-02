package query

import (
	"bytes"
	"fmt"
	"gx/internal/index"
	"gx/internal/language"
	"gx/internal/output"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	directoryOverviewMaxSymbols = 10
	symbolPriorityPrimary       = iota
	symbolPrioritySecondary
	symbolPriorityFallback
)

type Runtime struct {
	Root  string
	JSON  bool
	Query *Service
}

func NewRuntime(root string, json bool) *Runtime {
	service := &Service{
		root:   filepath.Clean(root),
		stdout: os.Stdout,
		stderr: os.Stderr,
		json:   json,
	}
	return &Runtime{
		Root:  filepath.Clean(root),
		JSON:  json,
		Query: service,
	}
}

type Service struct {
	root   string
	stdout io.Writer
	stderr io.Writer
	json   bool
}

type SymbolRow struct {
	File      string `json:"file,omitempty"`
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

type UniqueCallerRow struct {
	File   string `json:"file"`
	Caller string `json:"caller"`
	Line   int    `json:"line"`
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

func (service *Service) Symbols(idx *index.Index, file *string, nameGlob *string, kind *index.SymbolKind) error {
	relPath := ""
	if file != nil {
		relPath = normalizeRelativePath(*file, idx.Root)
		if _, ok := idx.Entries[relPath]; !ok {
			return service.fileLookupError(relPath, idx.Root)
		}
	}

	rows := make([]SymbolRow, 0)
	for path, data := range idx.Entries {
		if relPath != "" && path != relPath {
			continue
		}
		for _, symbol := range data.Symbols {
			if nameGlob != nil {
				ok, err := globMatch(*nameGlob, symbol.Name)
				if err != nil || !ok {
					continue
				}
			}
			if kind != nil && symbol.Kind != *kind {
				continue
			}
			row := SymbolRow{
				Name:      symbol.Name,
				Kind:      string(symbol.Kind),
				Signature: symbol.Signature,
			}
			if relPath == "" {
				row.File = displayPath(path)
			}
			rows = append(rows, row)
		}
	}

	if len(rows) == 0 {
		_, _ = fmt.Fprintln(service.stderr, "gx: no matches")
		return nil
	}

	sort.Slice(rows, func(left int, right int) bool {
		if rows[left].File == rows[right].File {
			return rows[left].Name < rows[right].Name
		}
		return rows[left].File < rows[right].File
	})

	if service.json {
		return output.PrintJSON(service.stdout, rows)
	}
	return output.PrintTOON(service.stdout, rows)
}

func (service *Service) Definition(idx *index.Index, name string, from *string, kind *index.SymbolKind, maxLines int) error {
	fromRel := ""
	if from != nil {
		fromRel = normalizeRelativePath(*from, idx.Root)
	}

	type definitionMatch struct {
		path   string
		symbol index.Symbol
	}

	matches := make([]definitionMatch, 0)
	for path, data := range idx.Entries {
		for _, symbol := range data.Symbols {
			if symbol.Name != name {
				continue
			}
			if kind != nil && symbol.Kind != *kind {
				continue
			}
			matches = append(matches, definitionMatch{path: path, symbol: symbol})
		}
	}

	if fromRel != "" {
		filtered := make([]definitionMatch, 0)
		for _, match := range matches {
			if match.path == fromRel {
				filtered = append(filtered, match)
			}
		}
		if len(filtered) > 0 {
			matches = filtered
		}
	}

	if len(matches) == 0 {
		_, _ = fmt.Fprintln(service.stderr, "gx: no matches")
		return nil
	}

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
			if _, err := fmt.Fprintf(service.stdout, "\ntruncated: %d lines total", result.Lines); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(service.stdout, "\n---\n%s\n", result.Body); err != nil {
			return err
		}
	}
	return nil
}

func (service *Service) References(idx *index.Index, name string, file *string, unique bool) error {
	relPath := ""
	if file != nil {
		relPath = normalizeRelativePath(*file, idx.Root)
		if _, ok := idx.Entries[relPath]; !ok {
			return service.fileLookupError(relPath, idx.Root)
		}
	}

	rows := make([]ReferenceRow, 0)
	nameBytes := []byte(name)
	for path, data := range idx.Entries {
		if relPath != "" && path != relPath {
			continue
		}

		source, err := os.ReadFile(filepath.Join(idx.Root, path))
		if err != nil || !bytes.Contains(source, nameBytes) {
			continue
		}

		references, err := language.FindReferences(data.Meta.Language, source, filepath.Join(idx.Root, path), name)
		if err != nil {
			if language.IsNotInstalled(err) {
				return err
			}
			continue
		}

		lines := splitLines(source)
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
		if service.json {
			return output.PrintJSON(service.stdout, uniqueRows)
		}
		return output.PrintTOON(service.stdout, uniqueRows)
	}

	if service.json {
		return output.PrintJSON(service.stdout, deduped)
	}
	return output.PrintTOON(service.stdout, deduped)
}

func (service *Service) DirectoryOverview(idx *index.Index, dir string, full bool) error {
	relDir := normalizeRelativePath(dir, idx.Root)
	if relDir == "." {
		relDir = ""
	}

	type entry struct {
		path string
		data index.FileData
	}
	allEntries := make([]entry, 0)
	for path, data := range idx.Entries {
		if relDir != "" && !strings.HasPrefix(path, relDir) {
			continue
		}
		if isTestFile(path) {
			continue
		}
		allEntries = append(allEntries, entry{path: path, data: data})
	}

	if len(allEntries) == 0 {
		return fmt.Errorf("gx: no indexed files under %s", displayPath(relDir))
	}

	directFiles := make([]entry, 0)
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

	if full {
		rows := make([]DirOverviewFullRow, 0)
		subdirNames := sortedKeys(subdirs)
		for _, name := range subdirNames {
			stats := subdirs[name]
			rows = append(rows, DirOverviewFullRow{
				File:      formatSubdir(relDir, name),
				Name:      fmt.Sprintf("(%d files, %d symbols)", stats[0], stats[1]),
				Kind:      "",
				Signature: "",
			})
		}

		for _, fileEntry := range directFiles {
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

		if service.json {
			return output.PrintJSON(service.stdout, rows)
		}
		return output.PrintTOON(service.stdout, rows)
	}

	rows := make([]DirOverviewRow, 0)
	for _, name := range sortedKeys(subdirs) {
		stats := subdirs[name]
		rows = append(rows, DirOverviewRow{
			File:    formatSubdir(relDir, name),
			Symbols: fmt.Sprintf("(%d files, %d symbols)", stats[0], stats[1]),
		})
	}

	for _, fileEntry := range directFiles {
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

	if service.json {
		return output.PrintJSON(service.stdout, rows)
	}
	return output.PrintTOON(service.stdout, rows)
}

func (service *Service) fileLookupError(relPath string, root string) error {
	absPath := filepath.Join(root, relPath)
	if _, err := os.Stat(absPath); err == nil && language.DetectLanguage(absPath) == "" {
		return fmt.Errorf("gx: unsupported file type: .%s", strings.TrimPrefix(filepath.Ext(absPath), "."))
	}
	return fmt.Errorf("gx: file not in index: %s", displayPath(relPath))
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
	case index.SymbolKindStruct, index.SymbolKindEnum, index.SymbolKindTrait, index.SymbolKindInterface, index.SymbolKindClass:
		return symbolPriorityPrimary
	case index.SymbolKindFn, index.SymbolKindConst, index.SymbolKindType, index.SymbolKindModule, index.SymbolKindEvent:
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
