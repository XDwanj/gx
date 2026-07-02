package query

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/XDwanj/gx/internal/aifilter"
	"github.com/XDwanj/gx/internal/index"
	"github.com/XDwanj/gx/internal/language"
	"github.com/XDwanj/gx/internal/output"
)

const (
	directoryOverviewMaxSymbols = 10
	treeAIConcurrencyLimit      = 256
	symbolPriorityPrimary       = iota
	symbolPrioritySecondary
	symbolPriorityFallback
)

var treeAIRequestLimiter = make(chan struct{}, treeAIConcurrencyLimit)

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
	root       string
	stdout     io.Writer
	stderr     io.Writer
	json       bool
	verbose    bool
	aiSelector aifilter.Selector
	debugMu    sync.Mutex
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

type symbolTextRow struct {
	File      string `json:"file"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Signature string `json:"signature"`
}

type referenceTextRow struct {
	File    string `json:"file"`
	Caller  string `json:"caller,omitempty"`
	Context string `json:"context"`
}

type calleeTextRow struct {
	File    string `json:"file"`
	Caller  string `json:"caller"`
	Callee  string `json:"callee"`
	Context string `json:"context"`
}

type uniqueCallerTextRow struct {
	File   string `json:"file"`
	Caller string `json:"caller"`
}

type treeTextRow struct {
	Tree      string `json:"tree"`
	Depth     int    `json:"depth"`
	Path      string `json:"path"`
	File      string `json:"file"`
	Name      string `json:"name"`
	Signature string `json:"signature,omitempty"`
	Edge      string `json:"edge,omitempty"`
	Context   string `json:"context,omitempty"`
	Cycle     bool   `json:"cycle,omitempty"`
}

type definitionMatch struct {
	path   string
	symbol index.Symbol
}

type treeNode struct {
	match  definitionMatch
	target aifilter.Target
}

type treeEdge struct {
	match   definitionMatch
	target  aifilter.Target
	edge    string
	context string
}

type treeCallerCandidate struct {
	match   definitionMatch
	line    int
	context string
}

type treeWalkState struct {
	idx      *index.Index
	filter   *scopeFilter
	maxDepth int
}

type treeWalkResult struct {
	rows []TreeRow
	err  error
}

type treeEdgeResult struct {
	edges []treeEdge
	err   error
}

type scopeFilter struct {
	isSingleFile bool
	paths        map[string]struct{}
}

type pathFilterMatchers struct {
	includes []compiledGlob
	excludes []compiledGlob
	ignore   *index.IgnoreMatcher
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

type PathQuery struct {
	Targets []string
	Include []string
	Exclude []string
}

type AIOptions struct {
	DefineIn string
}

type SymbolsOptions struct {
	Paths    PathQuery
	NameGlob *string
	Kind     *index.SymbolKind
	Page     PageRequest
	AI       AIOptions
}

type DefinitionOptions struct {
	Paths    PathQuery
	NameGlob string
	Kind     *index.SymbolKind
	MaxLines int
	Page     PageRequest
	AI       AIOptions
}

type ReferencesOptions struct {
	Paths    PathQuery
	NameGlob string
	Unique   bool
	Page     PageRequest
	AI       AIOptions
}

type CalleesOptions struct {
	Paths    PathQuery
	NameGlob string
	Page     PageRequest
	AI       AIOptions
}

const (
	TreeDirectionIn   = "in"
	TreeDirectionOut  = "out"
	TreeDirectionBoth = "both"
)

type TreeOptions struct {
	Paths     PathQuery
	NameGlob  string
	Direction string
	Depth     int
	AI        AIOptions
}

type TreeRow struct {
	Tree      string `json:"tree"`
	Depth     int    `json:"depth"`
	Path      string `json:"path"`
	File      string `json:"file"`
	Line      int    `json:"line"`
	Name      string `json:"name"`
	Signature string `json:"signature,omitempty"`
	Edge      string `json:"edge,omitempty"`
	Context   string `json:"context,omitempty"`
	Cycle     bool   `json:"cycle,omitempty"`
}

type TreeResult struct {
	In  *TreeOutputNode `json:"in,omitempty" toon:"in,omitempty"`
	Out *TreeOutputNode `json:"out,omitempty" toon:"out,omitempty"`
}

type TreeOutputNode struct {
	File   string           `json:"file" toon:"file"`
	Symbol string           `json:"symbol" toon:"symbol"`
	Cycle  bool             `json:"cycle,omitempty" toon:"cycle,omitempty"`
	In     []TreeOutputNode `json:"in,omitempty" toon:"in,omitempty"`
	Out    []TreeOutputNode `json:"out,omitempty" toon:"out,omitempty"`
}

type treeBuildNode struct {
	row      TreeRow
	children []*treeBuildNode
}

type pageInfo struct {
	Total      int
	Offset     int
	Returned   int
	NextOffset int
	Truncated  bool
	OutOfRange bool
}

func (service *Service) Symbols(idx *index.Index, options SymbolsOptions) error {
	rows, resolvedIdx, err := service.symbolRowsInContext(idx, options.Paths, options.NameGlob, options.Kind)
	if err != nil {
		return err
	}
	if options.AI.DefineIn != "" && options.NameGlob == nil {
		return fmt.Errorf("gx: --define-in requires --name")
	}

	if len(rows) == 0 {
		if service.json {
			return output.PrintJSON(service.stdout, []SymbolRow{})
		}
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

	if options.AI.DefineIn != "" {
		target, err := service.resolveAITarget(resolvedIdx, options.AI.DefineIn, *options.NameGlob, options.Kind)
		if err != nil {
			return err
		}
		rows, err = service.filterSymbolRowsWithAI(resolvedIdx, "symbols", *options.NameGlob, target, rows)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			if service.json {
				return output.PrintJSON(service.stdout, []SymbolRow{})
			}
			_, _ = fmt.Fprintln(service.stderr, "gx: no matches")
			return nil
		}
	}

	rows, pageState := paginateRows(rows, options.Page)
	service.writePageHint(pageState)

	if service.json {
		return output.PrintJSON(service.stdout, rows)
	}
	return output.PrintTOON(service.stdout, humanizeRows(rows))
}

func (service *Service) Overview(idx *index.Index, paths []string, full bool, page PageRequest) error {
	if len(paths) <= 1 {
		if len(paths) == 0 {
			return service.Symbols(idx, SymbolsOptions{Paths: PathQuery{Targets: paths}, Page: page})
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
		return service.Symbols(idx, SymbolsOptions{
			Paths: PathQuery{Targets: []string{target}},
			Page:  PageRequest{},
		})
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
		rows, err := service.symbolRows(idx, PathQuery{Targets: []string{target}}, nil, nil)
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
		if err := output.PrintTOON(&rendered, humanizeRows(section.Rows)); err != nil {
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

func (service *Service) symbolRows(idx *index.Index, paths PathQuery, nameGlob *string, kind *index.SymbolKind) ([]SymbolRow, error) {
	rows, _, err := service.symbolRowsInContext(idx, paths, nameGlob, kind)
	return rows, err
}

func (service *Service) symbolRowsInContext(idx *index.Index, paths PathQuery, nameGlob *string, kind *index.SymbolKind) ([]SymbolRow, *index.Index, error) {
	resolvedIdx, filter, err := service.resolveQueryContext(idx, paths)
	if err != nil {
		return nil, nil, err
	}

	rows := make([]SymbolRow, 0)
	for path, data := range resolvedIdx.Entries {
		if filter != nil && !filter.contains(path) {
			continue
		}
		for _, symbol := range data.Symbols {
			if nameGlob != nil {
				ok, err := globMatch(*nameGlob, symbol.Name)
				if err != nil {
					return nil, nil, err
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

	return rows, resolvedIdx, nil
}

func (service *Service) Definition(idx *index.Index, options DefinitionOptions) error {
	resolvedIdx, filter, err := service.resolveQueryContext(idx, options.Paths)
	if err != nil {
		return err
	}

	matches, err := findDefinitionMatches(resolvedIdx, options.NameGlob, options.Kind)
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
		if service.json {
			return output.PrintJSON(service.stdout, []DefinitionResult{})
		}
		_, _ = fmt.Fprintln(service.stderr, "gx: no matches")
		return nil
	}

	sortDefinitionMatches(matches)
	if options.AI.DefineIn != "" {
		target, err := service.resolveAITarget(resolvedIdx, options.AI.DefineIn, options.NameGlob, options.Kind)
		if err != nil {
			return err
		}
		matches, err = service.filterDefsWithAI(resolvedIdx, "definition", options.NameGlob, target, matches)
		if err != nil {
			return err
		}
		if len(matches) == 0 {
			if service.json {
				return output.PrintJSON(service.stdout, []DefinitionResult{})
			}
			_, _ = fmt.Fprintln(service.stderr, "gx: no matches")
			return nil
		}
	}
	matches, pageState := paginateRows(matches, options.Page)
	service.writePageHint(pageState)

	results := make([]DefinitionResult, 0, len(matches))
	for _, match := range matches {
		body, startLine, ok := readBody(resolvedIdx.Root, match.path, match.symbol)
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
		if lineCount > options.MaxLines && options.MaxLines > 0 {
			lines := strings.Split(body, "\n")
			result.Truncated = true
			result.Lines = lineCount
			result.Body = strings.Join(lines[:options.MaxLines], "\n")
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
		if _, err := fmt.Fprintf(service.stdout, "file: %s", formatLocation(result.File, result.Line)); err != nil {
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

func (service *Service) References(idx *index.Index, options ReferencesOptions) error {
	resolvedIdx, filter, err := service.resolveQueryContext(idx, options.Paths)
	if err != nil {
		return err
	}

	matchedNames, err := findReferenceNames(resolvedIdx, options.NameGlob, filter)
	if err != nil {
		return err
	}
	if len(matchedNames) == 0 {
		if service.json {
			return output.PrintJSON(service.stdout, []ReferenceRow{})
		}
		_, _ = fmt.Fprintln(service.stderr, "gx: no matches")
		return nil
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
		if service.json {
			return output.PrintJSON(service.stdout, []ReferenceRow{})
		}
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

	if options.AI.DefineIn != "" {
		target, err := service.resolveAITarget(resolvedIdx, options.AI.DefineIn, options.NameGlob, nil)
		if err != nil {
			return err
		}
		deduped, err = service.filterReferenceRowsWithAI(resolvedIdx, options.NameGlob, target, deduped)
		if err != nil {
			return err
		}
		if len(deduped) == 0 {
			if service.json {
				return output.PrintJSON(service.stdout, []ReferenceRow{})
			}
			_, _ = fmt.Fprintln(service.stderr, "gx: no matches")
			return nil
		}
	}

	if options.Unique {
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
			if service.json {
				return output.PrintJSON(service.stdout, []UniqueCallerRow{})
			}
			_, _ = fmt.Fprintln(service.stderr, "gx: no callers found")
			return nil
		}
		uniqueRows, pageState := paginateRows(uniqueRows, options.Page)
		service.writePageHint(pageState)
		if service.json {
			return output.PrintJSON(service.stdout, uniqueRows)
		}
		return output.PrintTOON(service.stdout, humanizeRows(uniqueRows))
	}

	deduped, pageState := paginateRows(deduped, options.Page)
	service.writePageHint(pageState)
	if service.json {
		return output.PrintJSON(service.stdout, deduped)
	}
	return output.PrintTOON(service.stdout, humanizeRows(deduped))
}

func (service *Service) Callees(idx *index.Index, options CalleesOptions) error {
	resolvedIdx, filter, err := service.resolveQueryContext(idx, options.Paths)
	if err != nil {
		return err
	}

	callableKind := index.SymbolKindFunc
	matches, err := findDefinitionMatches(resolvedIdx, options.NameGlob, &callableKind)
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
		if service.json {
			return output.PrintJSON(service.stdout, []CalleeRow{})
		}
		_, _ = fmt.Fprintln(service.stderr, "gx: no matches")
		return nil
	}
	if options.AI.DefineIn != "" {
		target, targetErr := service.resolveAITarget(resolvedIdx, options.AI.DefineIn, options.NameGlob, &callableKind)
		if targetErr != nil {
			return targetErr
		}
		matches, err = service.filterDefsWithAI(resolvedIdx, "callees", options.NameGlob, target, matches)
		if err != nil {
			return err
		}
		if len(matches) == 0 {
			if service.json {
				return output.PrintJSON(service.stdout, []CalleeRow{})
			}
			_, _ = fmt.Fprintln(service.stderr, "gx: no matches")
			return nil
		}
	}

	rows := make([]CalleeRow, 0)
	sourceCache := make(map[string][]byte)
	linesCache := make(map[string][]string)
	for _, match := range matches {
		data, ok := resolvedIdx.Entries[match.path]
		if !ok {
			continue
		}
		if !language.SupportsCallees(data.Meta.Language) {
			return fmt.Errorf("gx: callees not supported for language: %s", data.Meta.Language)
		}

		source, ok := sourceCache[match.path]
		if !ok {
			source, err = os.ReadFile(filepath.Join(resolvedIdx.Root, match.path))
			if err != nil {
				return err
			}
			sourceCache[match.path] = source
			linesCache[match.path] = splitLines(source)
		}

		callees, err := language.FindCallees(
			data.Meta.Language,
			source,
			filepath.Join(resolvedIdx.Root, match.path),
			match.symbol.ByteStart,
			match.symbol.ByteEnd,
		)
		if err != nil {
			return err
		}

		lines := linesCache[match.path]
		for _, callee := range callees {
			callMatches := findScopedCallableMatchesByCall(resolvedIdx, match.path, callee.Name, filter)
			if len(callMatches) == 0 {
				continue
			}

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
		if service.json {
			return output.PrintJSON(service.stdout, []CalleeRow{})
		}
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

	rows, pageState := paginateRows(rows, options.Page)
	service.writePageHint(pageState)
	if service.json {
		return output.PrintJSON(service.stdout, rows)
	}
	return output.PrintTOON(service.stdout, humanizeRows(rows))
}

func (service *Service) Tree(idx *index.Index, options TreeOptions) error {
	if options.AI.DefineIn == "" {
		return fmt.Errorf("gx: tree requires --define-in")
	}
	if options.Depth < 0 {
		return fmt.Errorf("gx: --depth must be greater than or equal to 0")
	}

	resolvedIdx, filter, err := service.resolveQueryContext(idx, options.Paths)
	if err != nil {
		return err
	}

	rootMatch, err := service.resolveAIMatch(resolvedIdx, options.AI.DefineIn, options.NameGlob)
	if err != nil {
		return err
	}
	if filter != nil && !filter.contains(rootMatch.path) {
		return fmt.Errorf("gx: --define-in target is outside query paths: %s", displayPath(rootMatch.path))
	}

	root := treeNode{
		match:  rootMatch,
		target: treeTargetForMatch(resolvedIdx, rootMatch),
	}
	rootKey := treeNodeKey(rootMatch)
	rows := make([]TreeRow, 0)
	state := treeWalkState{
		idx:      resolvedIdx,
		filter:   filter,
		maxDepth: options.Depth,
	}
	service.debugf(
		"tree root=%s file=%s depth=%d direction=%s ai_limit=%d",
		root.match.symbol.Name,
		displayPath(root.match.path),
		options.Depth,
		options.Direction,
		treeAIConcurrencyLimit,
	)

	switch options.Direction {
	case "", TreeDirectionBoth:
		inRows, outRows, err := service.walkBothTrees(state, root, rootKey)
		if err != nil {
			return err
		}
		rows = append(rows, inRows...)
		rows = append(rows, outRows...)
	case TreeDirectionIn:
		rows = append(rows, treeRow(TreeDirectionIn, 0, "0", root, "", "", false))
		childRows, err := service.walkInTree(state, root, "0", 0, map[string]bool{rootKey: true})
		if err != nil {
			return err
		}
		rows = append(rows, childRows...)
	case TreeDirectionOut:
		rows = append(rows, treeRow(TreeDirectionOut, 0, "0", root, "", "", false))
		childRows, err := service.walkOutTree(state, root, "0", 0, map[string]bool{rootKey: true})
		if err != nil {
			return err
		}
		rows = append(rows, childRows...)
	default:
		return fmt.Errorf("gx: --direction must be in, out, or both")
	}

	result := buildTreeResult(rows)
	if service.json {
		return output.PrintJSON(service.stdout, result)
	}
	return printTreeResult(service.stdout, result)
}

func (service *Service) walkBothTrees(state treeWalkState, root treeNode, rootKey string) ([]TreeRow, []TreeRow, error) {
	inCh := make(chan treeWalkResult, 1)
	outCh := make(chan treeWalkResult, 1)

	go func() {
		rows := make([]TreeRow, 0, 1)
		rows = append(rows, treeRow(TreeDirectionIn, 0, "0", root, "", "", false))
		childRows, err := service.walkInTree(state, root, "0", 0, map[string]bool{rootKey: true})
		rows = append(rows, childRows...)
		inCh <- treeWalkResult{rows: rows, err: err}
	}()
	go func() {
		rows := make([]TreeRow, 0, 1)
		rows = append(rows, treeRow(TreeDirectionOut, 0, "0", root, "", "", false))
		childRows, err := service.walkOutTree(state, root, "0", 0, map[string]bool{rootKey: true})
		rows = append(rows, childRows...)
		outCh <- treeWalkResult{rows: rows, err: err}
	}()

	inResult := <-inCh
	outResult := <-outCh
	if inResult.err != nil {
		return nil, nil, inResult.err
	}
	if outResult.err != nil {
		return nil, nil, outResult.err
	}
	return inResult.rows, outResult.rows, nil
}

func (service *Service) walkOutTree(state treeWalkState, parent treeNode, parentPath string, depth int, ancestors map[string]bool) ([]TreeRow, error) {
	if depth >= state.maxDepth {
		return nil, nil
	}

	edges, err := service.treeOutEdges(state.idx, state.filter, parent)
	if err != nil {
		return nil, err
	}
	service.debugf("tree out expand depth=%d path=%s node=%s edges=%d", depth, parentPath, parent.match.symbol.Name, len(edges))

	results := make([]treeWalkResult, len(edges))
	var wg sync.WaitGroup
	for edgeIndex, edge := range edges {
		childPath := fmt.Sprintf("%s.%d", parentPath, edgeIndex+1)
		child := treeNode{match: edge.match, target: edge.target}
		key := treeNodeKey(edge.match)
		cycle := ancestors[key]
		row := treeRow(TreeDirectionOut, depth+1, childPath, child, edge.edge, edge.context, cycle)
		if cycle {
			results[edgeIndex] = treeWalkResult{rows: []TreeRow{row}}
			continue
		}

		childAncestors := cloneTreeAncestors(ancestors)
		childAncestors[key] = true
		wg.Add(1)
		go func(indexValue int, childNode treeNode, path string, row TreeRow, nextAncestors map[string]bool) {
			defer wg.Done()
			childRows, walkErr := service.walkOutTree(state, childNode, path, depth+1, nextAncestors)
			rows := append([]TreeRow{row}, childRows...)
			results[indexValue] = treeWalkResult{rows: rows, err: walkErr}
		}(edgeIndex, child, childPath, row, childAncestors)
	}
	wg.Wait()

	rows := make([]TreeRow, 0)
	for _, result := range results {
		if result.err != nil {
			return nil, result.err
		}
		rows = append(rows, result.rows...)
	}
	return rows, nil
}

func (service *Service) walkInTree(state treeWalkState, parent treeNode, parentPath string, depth int, ancestors map[string]bool) ([]TreeRow, error) {
	if depth >= state.maxDepth {
		return nil, nil
	}

	edges, err := service.treeInEdges(state.idx, state.filter, parent)
	if err != nil {
		return nil, err
	}
	service.debugf("tree in expand depth=%d path=%s node=%s edges=%d", depth, parentPath, parent.match.symbol.Name, len(edges))

	results := make([]treeWalkResult, len(edges))
	var wg sync.WaitGroup
	for edgeIndex, edge := range edges {
		childPath := fmt.Sprintf("%s.%d", parentPath, edgeIndex+1)
		child := treeNode{match: edge.match, target: edge.target}
		key := treeNodeKey(edge.match)
		cycle := ancestors[key]
		row := treeRow(TreeDirectionIn, depth+1, childPath, child, edge.edge, edge.context, cycle)
		if cycle {
			results[edgeIndex] = treeWalkResult{rows: []TreeRow{row}}
			continue
		}

		childAncestors := cloneTreeAncestors(ancestors)
		childAncestors[key] = true
		wg.Add(1)
		go func(indexValue int, childNode treeNode, path string, row TreeRow, nextAncestors map[string]bool) {
			defer wg.Done()
			childRows, walkErr := service.walkInTree(state, childNode, path, depth+1, nextAncestors)
			rows := append([]TreeRow{row}, childRows...)
			results[indexValue] = treeWalkResult{rows: rows, err: walkErr}
		}(edgeIndex, child, childPath, row, childAncestors)
	}
	wg.Wait()

	rows := make([]TreeRow, 0)
	for _, result := range results {
		if result.err != nil {
			return nil, result.err
		}
		rows = append(rows, result.rows...)
	}
	return rows, nil
}

func (service *Service) treeOutEdges(idx *index.Index, filter *scopeFilter, parent treeNode) ([]treeEdge, error) {
	data, ok := idx.Entries[parent.match.path]
	if !ok {
		return nil, nil
	}
	if !language.SupportsCallees(data.Meta.Language) {
		return nil, fmt.Errorf("gx: callees not supported for language: %s", data.Meta.Language)
	}

	source, err := os.ReadFile(filepath.Join(idx.Root, parent.match.path))
	if err != nil {
		return nil, err
	}
	callees, err := language.FindCallees(
		data.Meta.Language,
		source,
		filepath.Join(idx.Root, parent.match.path),
		parent.match.symbol.ByteStart,
		parent.match.symbol.ByteEnd,
	)
	if err != nil {
		return nil, err
	}

	lines := splitLines(source)
	results := make([]treeEdgeResult, len(callees))
	var wg sync.WaitGroup
	for calleeIndex, callee := range callees {
		context := lineContext(lines, callee.Line)
		matches := findScopedCallableMatchesByCall(idx, parent.match.path, callee.Name, filter)
		if len(matches) == 0 {
			continue
		}

		wg.Add(1)
		go func(indexValue int, calleeName string, callContext string, callMatches []definitionMatch) {
			defer wg.Done()
			filteredMatches, filterErr := service.filterTreeOutAI(idx, parent.target, calleeName, callContext, callMatches)
			if filterErr != nil {
				results[indexValue] = treeEdgeResult{err: filterErr}
				return
			}

			callEdges := make([]treeEdge, 0, len(filteredMatches))
			for _, match := range filteredMatches {
				callEdges = append(callEdges, treeEdge{
					match:   match,
					target:  treeTargetForMatch(idx, match),
					edge:    calleeName,
					context: callContext,
				})
			}
			results[indexValue] = treeEdgeResult{edges: callEdges}
		}(calleeIndex, callee.Name, context, matches)
	}
	wg.Wait()

	edges := make([]treeEdge, 0)
	seen := map[string]bool{}
	for _, result := range results {
		if result.err != nil {
			return nil, result.err
		}
		for _, edge := range result.edges {
			key := treeNodeKey(edge.match)
			if seen[key] {
				continue
			}
			seen[key] = true
			edges = append(edges, edge)
		}
	}
	return edges, nil
}

func (service *Service) treeInEdges(idx *index.Index, filter *scopeFilter, parent treeNode) ([]treeEdge, error) {
	candidates, err := service.treeInCandidates(idx, filter, parent)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, nil
	}

	candidates, err = service.filterTreeInWithAI(idx, parent.target, candidates)
	if err != nil {
		return nil, err
	}

	edges := make([]treeEdge, 0)
	seen := map[string]bool{}
	for _, candidate := range candidates {
		key := treeNodeKey(candidate.match)
		if seen[key] {
			continue
		}
		seen[key] = true
		edges = append(edges, treeEdge{
			match:   candidate.match,
			target:  treeTargetForMatch(idx, candidate.match),
			edge:    parent.match.symbol.Name,
			context: candidate.context,
		})
	}
	return edges, nil
}

func (service *Service) treeInCandidates(idx *index.Index, filter *scopeFilter, parent treeNode) ([]treeCallerCandidate, error) {
	candidates := make([]treeCallerCandidate, 0)
	nameBytes := [][]byte{[]byte(parent.match.symbol.Name)}
	for path, data := range idx.Entries {
		if filter != nil && !filter.contains(path) {
			continue
		}
		if !sameDefinitionScope(parent.match.path, path) {
			continue
		}

		source, err := os.ReadFile(filepath.Join(idx.Root, path))
		if err != nil || !containsAnyName(source, nameBytes) {
			continue
		}

		references, findErr := language.FindReferencesForNames(
			data.Meta.Language,
			source,
			filepath.Join(idx.Root, path),
			[]string{parent.match.symbol.Name},
		)
		if findErr != nil {
			if language.IsNotInstalled(findErr) {
				return nil, findErr
			}
			continue
		}

		lines := splitLines(source)
		for _, reference := range references {
			caller, ok := findEnclosingFuncSymbol(data.Symbols, reference.ByteOffset)
			if !ok {
				continue
			}
			if path == parent.match.path && caller.ByteStart == parent.match.symbol.ByteStart && reference.Line == parent.match.symbol.Line {
				continue
			}
			candidates = append(candidates, treeCallerCandidate{
				match:   definitionMatch{path: path, symbol: caller},
				line:    reference.Line,
				context: lineContext(lines, reference.Line),
			})
		}
	}

	sort.Slice(candidates, func(left int, right int) bool {
		if candidates[left].match.path == candidates[right].match.path {
			return candidates[left].line < candidates[right].line
		}
		return candidates[left].match.path < candidates[right].match.path
	})
	return candidates, nil
}

func findCallableMatchesByNames(idx *index.Index, names []string, filter *scopeFilter) []definitionMatch {
	nameSet := make(map[string]bool, len(names))
	for _, name := range names {
		if name != "" {
			nameSet[name] = true
		}
	}

	matches := make([]definitionMatch, 0)
	for path, data := range idx.Entries {
		if filter != nil && !filter.contains(path) {
			continue
		}
		for _, symbol := range data.Symbols {
			if symbol.Kind != index.SymbolKindFunc {
				continue
			}
			if !nameSet[symbol.Name] {
				continue
			}
			matches = append(matches, definitionMatch{path: path, symbol: symbol})
		}
	}
	sortDefinitionMatches(matches)
	return matches
}

func findScopedCallableMatchesByCall(idx *index.Index, callerPath string, name string, filter *scopeFilter) []definitionMatch {
	matches := findCallableMatchesByNames(idx, callNameCandidates(name), filter)
	scoped := make([]definitionMatch, 0, len(matches))
	for _, match := range matches {
		if sameDefinitionScope(callerPath, match.path) {
			scoped = append(scoped, match)
		}
	}
	return scoped
}

func callNameCandidates(name string) []string {
	candidates := make([]string, 0, 3)
	addCallNameCandidate := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		for _, candidate := range candidates {
			if candidate == value {
				return
			}
		}
		candidates = append(candidates, value)
	}

	addCallNameCandidate(name)
	if indexValue := strings.LastIndex(name, "::"); indexValue >= 0 {
		addCallNameCandidate(name[indexValue+2:])
	}
	if indexValue := strings.LastIndex(name, "."); indexValue >= 0 {
		addCallNameCandidate(name[indexValue+1:])
	}
	if indexValue := strings.LastIndex(name, ":"); indexValue >= 0 {
		addCallNameCandidate(name[indexValue+1:])
	}
	return candidates
}

func sameDefinitionScope(leftPath string, rightPath string) bool {
	return filepath.Dir(leftPath) == filepath.Dir(rightPath)
}

func cloneTreeAncestors(ancestors map[string]bool) map[string]bool {
	cloned := make(map[string]bool, len(ancestors)+1)
	for key, value := range ancestors {
		cloned[key] = value
	}
	return cloned
}

func treeNodeKey(match definitionMatch) string {
	return fmt.Sprintf("%s:%d", match.path, match.symbol.ByteStart)
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

func formatLocation(path string, line int) string {
	if line <= 0 {
		return path
	}
	return fmt.Sprintf("%s:%d", path, line)
}

func humanizeRows(rows any) any {
	switch typed := rows.(type) {
	case []SymbolRow:
		human := make([]symbolTextRow, 0, len(typed))
		for _, row := range typed {
			human = append(human, symbolTextRow{
				File:      formatLocation(row.File, row.Line),
				Name:      row.Name,
				Kind:      row.Kind,
				Signature: row.Signature,
			})
		}
		return human
	case []ReferenceRow:
		human := make([]referenceTextRow, 0, len(typed))
		for _, row := range typed {
			human = append(human, referenceTextRow{
				File:    formatLocation(row.File, row.Line),
				Caller:  row.Caller,
				Context: row.Context,
			})
		}
		return human
	case []CalleeRow:
		human := make([]calleeTextRow, 0, len(typed))
		for _, row := range typed {
			human = append(human, calleeTextRow{
				File:    formatLocation(row.File, row.Line),
				Caller:  row.Caller,
				Callee:  row.Callee,
				Context: row.Context,
			})
		}
		return human
	case []UniqueCallerRow:
		human := make([]uniqueCallerTextRow, 0, len(typed))
		for _, row := range typed {
			human = append(human, uniqueCallerTextRow{
				File:   formatLocation(row.File, row.Line),
				Caller: row.Caller,
			})
		}
		return human
	case []TreeRow:
		human := make([]treeTextRow, 0, len(typed))
		for _, row := range typed {
			human = append(human, treeTextRow{
				Tree:      row.Tree,
				Depth:     row.Depth,
				Path:      row.Path,
				File:      formatLocation(row.File, row.Line),
				Name:      row.Name,
				Signature: row.Signature,
				Edge:      row.Edge,
				Context:   row.Context,
				Cycle:     row.Cycle,
			})
		}
		return human
	default:
		return rows
	}
}

const (
	aiSnippetRadius           = 4
	aiTargetBodyMaxRunes      = 8_000
	aiCandidateBodyMaxRunes   = 4_000
	aiCandidateSnippetMaxRune = 2_000
)

func (service *Service) resolveAITarget(idx *index.Index, defineIn string, nameGlob string, kind *index.SymbolKind) (aifilter.Target, error) {
	relPath := normalizeRelativePath(defineIn, idx.Root)
	data, ok := idx.Entries[relPath]
	if !ok {
		return aifilter.Target{}, service.fileLookupError(relPath, idx.Root)
	}

	matches := make([]index.Symbol, 0)
	for _, symbol := range data.Symbols {
		matched, err := globMatch(nameGlob, symbol.Name)
		if err != nil {
			return aifilter.Target{}, err
		}
		if !matched {
			continue
		}
		if kind != nil && symbol.Kind != *kind {
			continue
		}
		matches = append(matches, symbol)
	}
	if len(matches) == 0 {
		return aifilter.Target{}, fmt.Errorf("gx: --define-in %s has no symbol matching %q", displayScopePath(relPath), nameGlob)
	}
	if len(matches) > 1 {
		return aifilter.Target{}, fmt.Errorf("gx: --define-in %s matches multiple symbols for %q; narrow --name", displayScopePath(relPath), nameGlob)
	}

	body, line, ok := readBody(idx.Root, relPath, matches[0])
	if !ok {
		body = ""
		line = matches[0].Line
	}
	return aifilter.Target{
		File:      displayPath(relPath),
		Line:      line,
		Name:      matches[0].Name,
		Kind:      string(matches[0].Kind),
		Signature: matches[0].Signature,
		Body:      truncatePromptText(body, aiTargetBodyMaxRunes),
	}, nil
}

func (service *Service) resolveAIMatch(idx *index.Index, defineIn string, nameGlob string) (definitionMatch, error) {
	callableKind := index.SymbolKindFunc
	relPath := normalizeRelativePath(defineIn, idx.Root)
	data, ok := idx.Entries[relPath]
	if !ok {
		return definitionMatch{}, service.fileLookupError(relPath, idx.Root)
	}

	matches := make([]index.Symbol, 0)
	for _, symbol := range data.Symbols {
		matched, err := globMatch(nameGlob, symbol.Name)
		if err != nil {
			return definitionMatch{}, err
		}
		if !matched {
			continue
		}
		if symbol.Kind != callableKind {
			continue
		}
		matches = append(matches, symbol)
	}
	if len(matches) == 0 {
		return definitionMatch{}, fmt.Errorf("gx: --define-in %s has no function matching %q", displayScopePath(relPath), nameGlob)
	}
	if len(matches) > 1 {
		return definitionMatch{}, fmt.Errorf("gx: --define-in %s matches multiple functions for %q; narrow --name", displayScopePath(relPath), nameGlob)
	}
	return definitionMatch{path: relPath, symbol: matches[0]}, nil
}

func treeTargetForMatch(idx *index.Index, match definitionMatch) aifilter.Target {
	body, line, ok := readBody(idx.Root, match.path, match.symbol)
	if !ok {
		body = ""
		line = match.symbol.Line
	}
	return aifilter.Target{
		File:      displayPath(match.path),
		Line:      line,
		Name:      match.symbol.Name,
		Kind:      string(match.symbol.Kind),
		Signature: match.symbol.Signature,
		Body:      truncatePromptText(body, aiTargetBodyMaxRunes),
	}
}

func treeRow(tree string, depth int, path string, node treeNode, edge string, context string, cycle bool) TreeRow {
	return TreeRow{
		Tree:      tree,
		Depth:     depth,
		Path:      path,
		File:      displayPath(node.match.path),
		Line:      node.match.symbol.Line,
		Name:      node.match.symbol.Name,
		Signature: node.match.symbol.Signature,
		Edge:      edge,
		Context:   context,
		Cycle:     cycle,
	}
}

func buildTreeResult(rows []TreeRow) TreeResult {
	result := TreeResult{}
	if root := buildTreeRoot(rows, TreeDirectionIn); root != nil {
		result.In = root
	}
	if root := buildTreeRoot(rows, TreeDirectionOut); root != nil {
		result.Out = root
	}
	return result
}

func printTreeResult(writer io.Writer, result TreeResult) error {
	if result.In != nil {
		if _, err := fmt.Fprintln(writer, "in:"); err != nil {
			return err
		}
		if err := printTreeNode(writer, *result.In, TreeDirectionIn, "  ", false); err != nil {
			return err
		}
	}
	if result.Out != nil {
		if result.In != nil {
			if _, err := fmt.Fprintln(writer); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(writer, "out:"); err != nil {
			return err
		}
		if err := printTreeNode(writer, *result.Out, TreeDirectionOut, "  ", false); err != nil {
			return err
		}
	}
	return nil
}

func printTreeNode(writer io.Writer, node TreeOutputNode, direction string, indent string, listItem bool) error {
	prefix := indent
	if listItem {
		if _, err := fmt.Fprintf(writer, "%s- file: %s\n", indent, node.File); err != nil {
			return err
		}
		prefix = indent + "  "
	} else if _, err := fmt.Fprintf(writer, "%sfile: %s\n", indent, node.File); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(writer, "%ssymbol: %s\n", prefix, node.Symbol); err != nil {
		return err
	}
	if node.Cycle {
		if _, err := fmt.Fprintf(writer, "%scycle: true\n", prefix); err != nil {
			return err
		}
	}

	children := node.Out
	if direction == TreeDirectionIn {
		children = node.In
	}
	if len(children) == 0 {
		return nil
	}

	if _, err := fmt.Fprintf(writer, "%s%s:\n", prefix, direction); err != nil {
		return err
	}
	for _, child := range children {
		if err := printTreeNode(writer, child, direction, prefix+"  ", true); err != nil {
			return err
		}
	}
	return nil
}

func buildTreeRoot(rows []TreeRow, direction string) *TreeOutputNode {
	nodes := make(map[string]*treeBuildNode)
	for _, row := range rows {
		if row.Tree != direction {
			continue
		}

		node := &treeBuildNode{row: row}
		nodes[row.Path] = node
		if row.Path == "0" {
			continue
		}

		parent, ok := nodes[parentTreePath(row.Path)]
		if ok {
			parent.children = append(parent.children, node)
		}
	}

	root, ok := nodes["0"]
	if !ok {
		return nil
	}
	output := treeOutputNode(root, direction)
	return &output
}

func treeOutputNode(node *treeBuildNode, direction string) TreeOutputNode {
	children := make([]TreeOutputNode, 0, len(node.children))
	for _, child := range node.children {
		children = append(children, treeOutputNode(child, direction))
	}

	outputNode := TreeOutputNode{
		File:   formatLocation(node.row.File, node.row.Line),
		Symbol: node.row.Name,
		Cycle:  node.row.Cycle,
	}
	switch direction {
	case TreeDirectionIn:
		outputNode.In = children
	case TreeDirectionOut:
		outputNode.Out = children
	}
	return outputNode
}

func parentTreePath(path string) string {
	indexValue := strings.LastIndex(path, ".")
	if indexValue < 0 {
		return ""
	}
	return path[:indexValue]
}

func (service *Service) filterSymbolRowsWithAI(idx *index.Index, command string, name string, target aifilter.Target, rows []SymbolRow) ([]SymbolRow, error) {
	candidates := make([]aifilter.Candidate, 0, len(rows))
	for indexValue, row := range rows {
		snippet, err := snippetForLine(idx.Root, row.File, row.Line)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, aifilter.Candidate{
			ID:        candidateID(indexValue),
			File:      row.File,
			Line:      row.Line,
			Name:      row.Name,
			Kind:      row.Kind,
			Signature: row.Signature,
			Snippet:   snippet,
		})
	}

	selected, err := service.selectCandidateIDs(idx.Root, command, name, target, candidates)
	if err != nil {
		return nil, err
	}
	filtered := make([]SymbolRow, 0, len(rows))
	for indexValue, row := range rows {
		if selected[candidateID(indexValue)] {
			filtered = append(filtered, row)
		}
	}
	return filtered, nil
}

func (service *Service) filterDefsWithAI(idx *index.Index, command, name string, target aifilter.Target, matches []definitionMatch) ([]definitionMatch, error) {
	candidates := make([]aifilter.Candidate, 0, len(matches))
	for indexValue, match := range matches {
		body, line, ok := readBody(idx.Root, match.path, match.symbol)
		if !ok {
			body = ""
			line = match.symbol.Line
		}
		candidates = append(candidates, aifilter.Candidate{
			ID:        candidateID(indexValue),
			File:      displayPath(match.path),
			Line:      line,
			Name:      match.symbol.Name,
			Kind:      string(match.symbol.Kind),
			Signature: match.symbol.Signature,
			Body:      truncatePromptText(body, aiCandidateBodyMaxRunes),
		})
	}

	selected, err := service.selectCandidateIDs(idx.Root, command, name, target, candidates)
	if err != nil {
		return nil, err
	}
	filtered := make([]definitionMatch, 0, len(matches))
	for indexValue, match := range matches {
		if selected[candidateID(indexValue)] {
			filtered = append(filtered, match)
		}
	}
	return filtered, nil
}

func (service *Service) filterReferenceRowsWithAI(idx *index.Index, name string, target aifilter.Target, rows []ReferenceRow) ([]ReferenceRow, error) {
	candidates := make([]aifilter.Candidate, 0, len(rows))
	for indexValue, row := range rows {
		snippet, err := snippetForLine(idx.Root, row.File, row.Line)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, aifilter.Candidate{
			ID:      candidateID(indexValue),
			File:    row.File,
			Line:    row.Line,
			Name:    name,
			Caller:  row.Caller,
			Context: row.Context,
			Snippet: snippet,
		})
	}

	selected, err := service.selectCandidateIDs(idx.Root, "references", name, target, candidates)
	if err != nil {
		return nil, err
	}
	filtered := make([]ReferenceRow, 0, len(rows))
	for indexValue, row := range rows {
		if selected[candidateID(indexValue)] {
			filtered = append(filtered, row)
		}
	}
	return filtered, nil
}

func (service *Service) filterTreeOutAI(idx *index.Index, target aifilter.Target, name, ctx string, matches []definitionMatch) ([]definitionMatch, error) {
	candidates := make([]aifilter.Candidate, 0, len(matches))
	for indexValue, match := range matches {
		body, line, ok := readBody(idx.Root, match.path, match.symbol)
		if !ok {
			body = ""
			line = match.symbol.Line
		}
		candidates = append(candidates, aifilter.Candidate{
			ID:        candidateID(indexValue),
			File:      displayPath(match.path),
			Line:      line,
			Name:      match.symbol.Name,
			Kind:      string(match.symbol.Kind),
			Callee:    name,
			Signature: match.symbol.Signature,
			Context:   ctx,
			Body:      truncatePromptText(body, aiCandidateBodyMaxRunes),
		})
	}

	selected, err := service.selectCandidateIDs(idx.Root, "tree-out", name, target, candidates)
	if err != nil {
		return nil, err
	}
	filtered := make([]definitionMatch, 0, len(matches))
	for indexValue, match := range matches {
		if selected[candidateID(indexValue)] {
			filtered = append(filtered, match)
		}
	}
	return filtered, nil
}

func (service *Service) filterTreeInWithAI(idx *index.Index, target aifilter.Target, candidates []treeCallerCandidate) ([]treeCallerCandidate, error) {
	aiCandidates := make([]aifilter.Candidate, 0, len(candidates))
	for indexValue, candidate := range candidates {
		snippet, err := snippetForLine(idx.Root, displayPath(candidate.match.path), candidate.line)
		if err != nil {
			return nil, err
		}
		aiCandidates = append(aiCandidates, aifilter.Candidate{
			ID:      candidateID(indexValue),
			File:    displayPath(candidate.match.path),
			Line:    candidate.line,
			Name:    target.Name,
			Caller:  candidate.match.symbol.Name,
			Context: candidate.context,
			Snippet: snippet,
		})
	}

	selected, err := service.selectCandidateIDs(idx.Root, "tree-in", target.Name, target, aiCandidates)
	if err != nil {
		return nil, err
	}
	filtered := make([]treeCallerCandidate, 0, len(candidates))
	for indexValue, candidate := range candidates {
		if selected[candidateID(indexValue)] {
			filtered = append(filtered, candidate)
		}
	}
	return filtered, nil
}

func (service *Service) selectCandidateIDs(root, command, name string, target aifilter.Target, candidates []aifilter.Candidate) (map[string]bool, error) {
	selector := service.aiSelector
	if selector == nil {
		var err error
		selector, err = aifilter.NewClientFromEnv()
		if err != nil {
			return nil, err
		}
	}
	request := aifilter.SelectionRequest{
		Command:    command,
		Name:       name,
		Target:     target,
		Candidates: candidates,
	}
	providerID := selector.ProviderID()
	key, err := aifilter.CacheKey(providerID, request)
	if err != nil {
		return nil, err
	}
	cache, err := aifilter.OpenCache(root)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = cache.Close()
	}()

	selectedIDs, found, err := cache.Get(key)
	if err != nil {
		return nil, err
	}
	if !found {
		service.debugf("ai cache miss command=%s name=%s target=%s candidates=%d limit=%d", command, name, target.Name, len(candidates), treeAIConcurrencyLimit)
		treeAIRequestLimiter <- struct{}{}
		service.debugf("ai request start command=%s name=%s target=%s candidates=%d", command, name, target.Name, len(candidates))
		defer func() {
			<-treeAIRequestLimiter
		}()
		selectedIDs, err = selector.Select(context.Background(), request)
		if err != nil {
			return nil, err
		}
		service.debugf("ai request done command=%s name=%s target=%s selected=%d", command, name, target.Name, len(selectedIDs))
		if err := cache.Put(key, providerID, selectedIDs); err != nil {
			return nil, err
		}
	} else {
		service.debugf("ai cache hit command=%s name=%s target=%s candidates=%d selected=%d", command, name, target.Name, len(candidates), len(selectedIDs))
	}

	return validateSelectedIDs(candidates, selectedIDs)
}

func (service *Service) debugf(format string, args ...any) {
	if !service.verbose {
		return
	}
	service.debugMu.Lock()
	defer service.debugMu.Unlock()
	_, _ = fmt.Fprintf(service.stderr, "gx: debug: "+format+"\n", args...)
}

func validateSelectedIDs(candidates []aifilter.Candidate, selectedIDs []string) (map[string]bool, error) {
	valid := make(map[string]bool, len(candidates))
	for _, candidate := range candidates {
		valid[candidate.ID] = true
	}

	selected := make(map[string]bool, len(selectedIDs))
	for _, id := range selectedIDs {
		if !valid[id] {
			return nil, fmt.Errorf("gx: AI selected unknown candidate id: %s", id)
		}
		selected[id] = true
	}
	return selected, nil
}

func candidateID(indexValue int) string {
	return fmt.Sprintf("c%d", indexValue)
}

func snippetForLine(root string, file string, line int) (string, error) {
	source, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file)))
	if err != nil {
		return "", err
	}
	lines := splitLines(source)
	if line <= 0 || line > len(lines) {
		return "", nil
	}
	start := line - aiSnippetRadius - 1
	if start < 0 {
		start = 0
	}
	end := line + aiSnippetRadius
	if end > len(lines) {
		end = len(lines)
	}
	return truncatePromptText(strings.Join(lines[start:end], "\n"), aiCandidateSnippetMaxRune), nil
}

func truncatePromptText(text string, limit int) string {
	runes := []rune(text)
	if limit <= 0 || len(runes) <= limit {
		return text
	}
	return string(runes[:limit]) + "\n... truncated ..."
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

func (service *Service) resolveQueryContext(idx *index.Index, query PathQuery) (*index.Index, *scopeFilter, error) {
	preparedIdx, matchers, err := preparePathQueryIndex(idx, query)
	if err != nil {
		return nil, nil, err
	}
	filter, err := service.resolvePaths(preparedIdx, query, matchers)
	if err != nil {
		return nil, nil, err
	}
	return preparedIdx, filter, nil
}

func preparePathQueryIndex(idx *index.Index, query PathQuery) (*index.Index, pathFilterMatchers, error) {
	includeMatchers, err := compilePathGlobs(query.Include, idx.Root)
	if err != nil {
		return nil, pathFilterMatchers{}, err
	}
	excludeMatchers, err := compilePathGlobs(query.Exclude, idx.Root)
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

func (service *Service) resolvePaths(idx *index.Index, query PathQuery, matchers pathFilterMatchers) (*scopeFilter, error) {
	filter := &scopeFilter{
		isSingleFile: len(query.Targets) == 1,
		paths:        make(map[string]struct{}),
	}

	if len(query.Targets) == 0 {
		for path := range idx.Entries {
			filter.paths[path] = struct{}{}
		}
		return applyPathGlobs(filter, matchers.includes, matchers.excludes, matchers.ignore), nil
	}

	missingPaths := make([]string, 0)
	for _, target := range query.Targets {
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

func findEnclosingFuncSymbol(symbols []index.Symbol, byteOffset uint) (index.Symbol, bool) {
	var result index.Symbol
	found := false
	bestWidth := ^uint(0)
	for _, symbol := range symbols {
		if symbol.Kind != index.SymbolKindFunc {
			continue
		}
		if symbol.ByteStart > byteOffset || byteOffset >= symbol.ByteEnd {
			continue
		}
		width := symbol.ByteEnd - symbol.ByteStart
		if width >= bestWidth {
			continue
		}
		bestWidth = width
		result = symbol
		found = true
	}
	return result, found
}

func lineContext(lines []string, line int) string {
	if line <= 0 || line > len(lines) {
		return ""
	}
	return strings.TrimSpace(lines[line-1])
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
