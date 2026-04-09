package index

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/XDwanj/gx/internal/language"

	ignore "github.com/sabhiram/go-gitignore"

	langpkg "github.com/XDwanj/gx/internal/lang"
)

const IndexVersion = 7

const (
	verboseProgressEveryFiles = 25
	verboseSlowFileThreshold  = 250 * time.Millisecond
	sqliteDriverName          = "sqlite"
	sqliteBusyTimeoutMillis   = 5_000
	cacheFileExtension        = ".sqlite3"
	gitDirName                = ".git"
	gitIgnoreFileName         = ".gitignore"
	dotIgnoreFileName         = ".ignore"
	gxIgnoreFileName          = ".gx-ignore"
)

type Index struct {
	Root    string
	Entries map[string]FileData
}

type LoadOptions struct {
	Verbose bool
	Stderr  io.Writer
}

type debugLogger struct {
	enabled bool
	stderr  io.Writer
}

type fileCandidate struct {
	AbsPath string
	RelPath string
	Info    os.FileInfo
}

type walkState struct {
	patterns []string
	matcher  *ignore.GitIgnore
}

func LoadOrBuild(root string) (*Index, error) {
	return LoadOrBuildWithOptions(root, LoadOptions{})
}

func LoadOrBuildWithOptions(root string, options LoadOptions) (*Index, error) {
	cleanRoot := filepath.Clean(root)
	cachePath := CachePathFor(cleanRoot)
	logger := newDebugLogger(options)
	logger.Printf("index root=%s", cleanRoot)
	logger.Printf("index cache=%s", cachePath)

	entries, err := loadEntries(cachePath)
	if err == nil && !needsUpdate(cleanRoot, entries) {
		logger.Printf("using cached index with %d entries", len(entries))
		return &Index{Root: cleanRoot, Entries: entries}, nil
	}
	if err != nil {
		logger.Printf("cache load miss or rebuild required: %v", err)
	} else {
		logger.Printf("cached index stale; rebuilding from %d entries", len(entries))
	}

	idx := &Index{
		Root:    cleanRoot,
		Entries: entries,
	}
	if idx.Entries == nil {
		idx.Entries = map[string]FileData{}
	}

	if len(idx.Entries) == 0 {
		logger.Printf("starting full index crawl")
		if err := idx.fullCrawl(logger); err != nil {
			return nil, err
		}
		logger.Printf("finished full index crawl with %d entries", len(idx.Entries))
	} else {
		logger.Printf("starting incremental index update from %d entries", len(idx.Entries))
		if err := idx.incrementalUpdate(logger); err != nil {
			return nil, err
		}
		logger.Printf("finished incremental index update with %d entries", len(idx.Entries))
	}

	if err := saveStore(cleanRoot, idx.Entries); err != nil {
		return nil, err
	}
	logger.Printf("saved index cache with %d entries", len(idx.Entries))
	return idx, nil
}

func CachePathFor(root string) string {
	return filepath.Join(cacheDir(), "indexes", sanitizeRoot(root)+cacheFileExtension)
}

func cacheDir() string {
	base, err := os.UserCacheDir()
	if err != nil {
		return filepath.Join(".cache", "gx")
	}
	return filepath.Join(base, "gx")
}

func NewFileEntry(modifiedAt time.Time, languageName string) FileEntry {
	return FileEntry{
		MTimeSecs:  modifiedAt.Unix(),
		MTimeNanos: int64(modifiedAt.Nanosecond()),
		Language:   languageName,
	}
}

func (entry FileEntry) MTime() time.Time {
	return time.Unix(entry.MTimeSecs, entry.MTimeNanos)
}

func needsUpdate(root string, entries map[string]FileData) bool {
	if len(entries) == 0 {
		return hasInstalledIndexableFiles(root)
	}

	indexedLanguages := map[string]bool{}
	for _, data := range entries {
		indexedLanguages[data.Meta.Language] = true
	}

	matchedCount := 0
	_ = walk(root, func(candidate fileCandidate) error {
		languageName := language.DetectLanguage(candidate.AbsPath)
		if languageName == "" {
			return nil
		}

		data, ok := entries[candidate.RelPath]
		if !ok {
			if indexedLanguages[languageName] {
				matchedCount = -1
				return fmt.Errorf("new file")
			}
			return nil
		}

		if !sameModTime(data.Meta.MTime(), candidate.Info.ModTime()) {
			matchedCount = -1
			return fmt.Errorf("mtime changed")
		}

		matchedCount++
		return nil
	})

	if matchedCount == -1 {
		return true
	}
	return matchedCount != len(entries)
}

func hasInstalledIndexableFiles(root string) bool {
	found := false
	_ = walk(root, func(candidate fileCandidate) error {
		languageName := language.DetectLanguage(candidate.AbsPath)
		if languageName == "" || !langpkg.IsInstalled(languageName) {
			return nil
		}
		found = true
		return fmt.Errorf("installed language file present")
	})
	return found
}

func (idx *Index) fullCrawl(logger *debugLogger) error {
	missingLanguages := map[string]int{}
	processedCount := 0

	return walk(idx.Root, func(candidate fileCandidate) error {
		languageName := language.DetectLanguage(candidate.AbsPath)
		if languageName == "" {
			return nil
		}

		processedCount++
		logger.Printf("indexing %s (%s)", candidate.RelPath, languageName)

		source, err := os.ReadFile(candidate.AbsPath)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "gx: warning: failed to read %s: %v\n", candidate.AbsPath, err)
			return nil
		}

		startedAt := time.Now()
		symbols, err := language.ParseAndExtract(languageName, source, candidate.AbsPath)
		if err != nil {
			if language.IsNotInstalled(err) {
				missingLanguages[languageName]++
				logger.Printf("skipping %s: %v", candidate.RelPath, err)
				return nil
			}
			logger.Printf("failed to parse %s: %v", candidate.RelPath, err)
			return nil
		}

		idx.Entries[candidate.RelPath] = FileData{
			Meta:    NewFileEntry(candidate.Info.ModTime(), languageName),
			Symbols: toIndexSymbols(symbols),
		}
		logIndexedFile(logger, candidate.RelPath, time.Since(startedAt), len(symbols))
		logIndexProgress(logger, processedCount, len(idx.Entries), "full crawl")
		return nil
	})
}

func (idx *Index) incrementalUpdate(logger *debugLogger) error {
	onDisk := map[string]fileCandidate{}
	missingLanguages := map[string]bool{}
	processedCount := 0

	if err := walk(idx.Root, func(candidate fileCandidate) error {
		languageName := language.DetectLanguage(candidate.AbsPath)
		if languageName == "" {
			return nil
		}
		onDisk[candidate.RelPath] = candidate
		return nil
	}); err != nil {
		return err
	}

	for path := range idx.Entries {
		if _, ok := onDisk[path]; !ok {
			delete(idx.Entries, path)
		}
	}

	for relPath, candidate := range onDisk {
		languageName := language.DetectLanguage(candidate.AbsPath)
		if languageName == "" {
			continue
		}
		data, ok := idx.Entries[relPath]
		if ok && sameModTime(data.Meta.MTime(), candidate.Info.ModTime()) {
			continue
		}

		source, err := os.ReadFile(candidate.AbsPath)
		if err != nil {
			logger.Printf("failed to read %s: %v", relPath, err)
			continue
		}

		processedCount++
		logger.Printf("re-indexing %s (%s)", relPath, languageName)
		startedAt := time.Now()
		symbols, err := language.ParseAndExtract(languageName, source, candidate.AbsPath)
		if err != nil {
			if language.IsNotInstalled(err) {
				missingLanguages[languageName] = true
			}
			logger.Printf("failed to parse %s: %v", relPath, err)
			continue
		}

		idx.Entries[relPath] = FileData{
			Meta:    NewFileEntry(candidate.Info.ModTime(), languageName),
			Symbols: toIndexSymbols(symbols),
		}
		logIndexedFile(logger, relPath, time.Since(startedAt), len(symbols))
		logIndexProgress(logger, processedCount, len(idx.Entries), "incremental update")
	}

	for languageName := range missingLanguages {
		_, _ = fmt.Fprintf(os.Stderr, "gx: skipping .%s files — enable with: gx lang enable %s\n", language.PrimaryExtension(languageName), languageName)
	}
	return nil
}

func newDebugLogger(options LoadOptions) *debugLogger {
	stderr := options.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}
	return &debugLogger{
		enabled: options.Verbose,
		stderr:  stderr,
	}
}

func (logger *debugLogger) Printf(format string, args ...any) {
	if !logger.enabled {
		return
	}
	_, _ = fmt.Fprintf(logger.stderr, "gx: debug: "+format+"\n", args...)
}

func logIndexedFile(logger *debugLogger, relPath string, duration time.Duration, symbolCount int) {
	if duration < verboseSlowFileThreshold {
		return
	}
	logger.Printf("indexed %s in %s (%d symbols)", relPath, duration.Round(time.Millisecond), symbolCount)
}

func logIndexProgress(logger *debugLogger, processedCount int, entryCount int, stage string) {
	if processedCount == 0 || processedCount%verboseProgressEveryFiles != 0 {
		return
	}
	logger.Printf("%s progress: processed %d files, %d indexed entries", stage, processedCount, entryCount)
}

func walk(root string, visit func(fileCandidate) error) error {
	cleanRoot := filepath.Clean(root)
	states := map[string]walkState{}

	return filepath.Walk(cleanRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if path == cleanRoot {
			states[path] = buildWalkState(cleanRoot, cleanRoot, walkState{})
			return nil
		}

		name := info.Name()
		if name == gitDirName {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() && fileExists(filepath.Join(path, gxIgnoreFileName)) {
			return filepath.SkipDir
		}

		relPath, relErr := filepath.Rel(cleanRoot, path)
		if relErr != nil {
			return nil
		}
		relPath = filepath.Clean(relPath)

		parentState, ok := states[filepath.Dir(path)]
		if !ok {
			return nil
		}

		if parentState.matches(relPath, info.IsDir()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			states[path] = buildWalkState(cleanRoot, path, parentState)
			return nil
		}

		return visit(fileCandidate{
			AbsPath: path,
			RelPath: relPath,
			Info:    info,
		})
	})
}

func buildWalkState(root string, dir string, parent walkState) walkState {
	nextPatterns := parent.patterns
	scope := ignoreScope(root, dir)

	nextPatterns = appendIgnoreFilePatterns(nextPatterns, scope, filepath.Join(dir, gitIgnoreFileName))
	nextPatterns = appendIgnoreFilePatterns(nextPatterns, scope, filepath.Join(dir, dotIgnoreFileName))

	if len(nextPatterns) == len(parent.patterns) {
		return parent
	}

	statePatterns := append([]string(nil), nextPatterns...)
	return walkState{
		patterns: statePatterns,
		matcher:  ignore.CompileIgnoreLines(statePatterns...),
	}
}

func appendIgnoreFilePatterns(patterns []string, scope string, path string) []string {
	lines, err := os.ReadFile(path)
	if err != nil {
		return patterns
	}

	result := patterns
	for _, line := range strings.Split(string(lines), "\n") {
		result = append(result, rewriteScopedIgnoreLine(scope, line))
	}
	return result
}

func rewriteScopedIgnoreLine(scope string, line string) string {
	if scope == "" {
		return line
	}

	trimmed := strings.TrimRight(line, "\r")
	trimmed = strings.TrimSpace(trimmed)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return line
	}

	negated := strings.HasPrefix(trimmed, "!")
	pattern := trimmed
	if negated {
		pattern = trimmed[1:]
	}

	if pattern == "" {
		return line
	}

	if strings.Contains(pattern, "/") {
		pattern = "/" + joinIgnoreScope(scope, strings.TrimPrefix(pattern, "/"))
	} else {
		pattern = "/" + joinIgnoreScope(scope, "**", pattern)
	}

	if negated {
		return "!" + pattern
	}
	return pattern
}

func joinIgnoreScope(parts ...string) string {
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		filtered = append(filtered, strings.Trim(part, "/"))
	}
	return strings.Join(filtered, "/")
}

func ignoreScope(root string, dir string) string {
	if dir == root {
		return ""
	}

	relPath, err := filepath.Rel(root, dir)
	if err != nil {
		return ""
	}
	return strings.ReplaceAll(filepath.Clean(relPath), "\\", "/")
}

func (state walkState) matches(relPath string, isDir bool) bool {
	if state.matcher == nil {
		return false
	}

	candidate := strings.ReplaceAll(relPath, "\\", "/")
	if isDir && !strings.HasSuffix(candidate, "/") {
		candidate += "/"
	}
	return state.matcher.MatchesPath(candidate)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func sameModTime(left time.Time, right time.Time) bool {
	return left.Unix() == right.Unix() && left.Nanosecond() == right.Nanosecond()
}

func sanitizeRoot(root string) string {
	clean := filepath.Clean(root)
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_")
	return replacer.Replace(clean)
}

func toIndexSymbols(symbols []language.Symbol) []Symbol {
	result := make([]Symbol, 0, len(symbols))
	for _, symbol := range symbols {
		result = append(result, Symbol{
			Name:      symbol.Name,
			Kind:      SymbolKind(symbol.Kind),
			Signature: symbol.Signature,
			Line:      symbol.Line,
			ByteStart: symbol.ByteStart,
			ByteEnd:   symbol.ByteEnd,
			IsTest:    symbol.IsTest,
		})
	}
	return result
}
