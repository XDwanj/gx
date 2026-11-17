package query

import (
	"github.com/XDwanj/gx/internal/index"
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

const (
	aiSnippetRadius           = 4
	aiTargetBodyMaxRunes      = 8_000
	aiCandidateBodyMaxRunes   = 4_000
	aiCandidateSnippetMaxRune = 2_000
)

// Intent describes which symbol the user is looking for.
// A non-empty DefinedIn points at the file holding the intended
// declaration and enables AI disambiguation for the query.
type Intent struct {
	Name      string
	Kind      *index.SymbolKind
	DefinedIn string
}

// Scope describes where the query runs.
type Scope struct {
	Targets []string
	Include []string
	Exclude []string
}

// Page describes how many rows to return. Limit 0 returns all rows.
type Page struct {
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

type resultPage struct {
	Total      int  `json:"total" toon:"total"`
	Offset     int  `json:"offset" toon:"offset"`
	NextOffset int  `json:"next_offset,omitempty" toon:"next_offset,omitempty"`
	OutOfRange bool `json:"out_of_range,omitempty" toon:"out_of_range,omitempty"`
}

type resultEnvelope struct {
	Rows any        `json:"rows" toon:"rows"`
	Page resultPage `json:"page" toon:"page"`
}

type SymbolRow struct {
	File      string `json:"file" toon:"file"`
	Line      int    `json:"line" toon:"line"`
	Name      string `json:"name" toon:"name"`
	Kind      string `json:"kind" toon:"kind"`
	Signature string `json:"signature,omitempty" toon:"signature,omitempty"`
}

type DefinitionResult struct {
	File      string `json:"file" toon:"file"`
	Line      int    `json:"line" toon:"line"`
	Body      string `json:"body" toon:"body"`
	Truncated bool   `json:"truncated,omitempty" toon:"truncated,omitempty"`
	Lines     int    `json:"lines,omitempty" toon:"lines,omitempty"`
}

type ReferenceRow struct {
	File    string `json:"file" toon:"file"`
	Line    int    `json:"line" toon:"line"`
	Caller  string `json:"caller" toon:"caller"`
	Context string `json:"context" toon:"context"`
}

type UniqueCallerRow struct {
	File   string `json:"file" toon:"file"`
	Caller string `json:"caller" toon:"caller"`
	Line   int    `json:"line" toon:"line"`
}

type DirOverviewRow struct {
	File    string `json:"file" toon:"file"`
	Symbols string `json:"symbols" toon:"symbols"`
}

type DirOverviewFullRow struct {
	File      string `json:"file" toon:"file"`
	Name      string `json:"name" toon:"name"`
	Kind      string `json:"kind" toon:"kind"`
	Signature string `json:"signature" toon:"signature"`
}

type OverviewSection struct {
	Target     string     `json:"target" toon:"target"`
	TargetKind string     `json:"target_kind" toon:"target_kind"`
	Rows       any        `json:"rows" toon:"rows"`
	Page       resultPage `json:"page" toon:"page"`
}

type symbolTextRow struct {
	File      string `json:"file" toon:"file"`
	Name      string `json:"name" toon:"name"`
	Kind      string `json:"kind" toon:"kind"`
	Signature string `json:"signature,omitempty" toon:"signature,omitempty"`
}

type referenceTextRow struct {
	File    string `json:"file" toon:"file"`
	Caller  string `json:"caller" toon:"caller"`
	Context string `json:"context" toon:"context"`
}

type uniqueCallerTextRow struct {
	File   string `json:"file" toon:"file"`
	Caller string `json:"caller" toon:"caller"`
}

type definitionMatch struct {
	path   string
	symbol index.Symbol
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
