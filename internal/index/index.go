package index

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gx/internal/language"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	ignore "github.com/sabhiram/go-gitignore"

	langpkg "gx/internal/lang"
)

const IndexVersion = 5

const (
	verboseProgressEveryFiles = 25
	verboseSlowFileThreshold  = 250 * time.Millisecond
	sqliteDriverName          = "sqlite"
	sqliteBusyTimeoutMillis   = 5_000
	cacheFileExtension        = ".sqlite3"
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

func loadEntries(path string) (map[string]FileData, error) {
	if !fileExists(path) {
		return nil, os.ErrNotExist
	}

	db, err := openStore(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()

	version, _, err := loadMeta(db)
	if err != nil {
		return nil, err
	}
	if version != IndexVersion {
		return nil, fmt.Errorf("version mismatch")
	}

	rows, err := db.Query(`
		SELECT path, mtime_secs, mtime_nanos, language, symbols_json
		FROM file_entries
	`)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	entries := map[string]FileData{}
	for rows.Next() {
		var (
			path       string
			mtimeSecs  int64
			mtimeNanos int64
			language   string
			symbolsRaw []byte
			symbols    []Symbol
		)
		if err := rows.Scan(&path, &mtimeSecs, &mtimeNanos, &language, &symbolsRaw); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(symbolsRaw, &symbols); err != nil {
			return nil, err
		}
		entries[path] = FileData{
			Meta: FileEntry{
				MTimeSecs:  mtimeSecs,
				MTimeNanos: mtimeNanos,
				Language:   language,
			},
			Symbols: symbols,
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func saveStore(root string, entries map[string]FileData) error {
	path := CachePathFor(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	db, err := openStore(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = db.Close()
	}()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, execErr := tx.Exec(`DELETE FROM store_meta`); execErr != nil {
		return execErr
	}
	if _, execErr := tx.Exec(
		`INSERT INTO store_meta (id, version, root) VALUES (1, ?, ?)`,
		IndexVersion,
		root,
	); execErr != nil {
		return execErr
	}
	if _, execErr := tx.Exec(`DELETE FROM file_entries`); execErr != nil {
		return execErr
	}

	statement, err := tx.Prepare(`
		INSERT INTO file_entries (path, mtime_secs, mtime_nanos, language, symbols_json)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer func() {
		_ = statement.Close()
	}()

	for path, data := range entries {
		symbolsJSON, err := json.Marshal(data.Symbols)
		if err != nil {
			return fmt.Errorf("gx: failed to encode index: %w", err)
		}
		if _, err := statement.Exec(
			path,
			data.Meta.MTimeSecs,
			data.Meta.MTimeNanos,
			data.Meta.Language,
			symbolsJSON,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func openStore(path string) (*sql.DB, error) {
	db, err := sql.Open(sqliteDriverName, path)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if _, err := db.Exec(fmt.Sprintf("PRAGMA busy_timeout = %d", sqliteBusyTimeoutMillis)); err != nil {
		_ = db.Close()
		return nil, err
	}

	if err := initStore(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func initStore(db *sql.DB) error {
	schema := []string{
		`
			CREATE TABLE IF NOT EXISTS store_meta (
				id INTEGER PRIMARY KEY CHECK (id = 1),
				version INTEGER NOT NULL,
				root TEXT NOT NULL
			)
		`,
		`
			CREATE TABLE IF NOT EXISTS file_entries (
				path TEXT PRIMARY KEY,
				mtime_secs INTEGER NOT NULL,
				mtime_nanos INTEGER NOT NULL,
				language TEXT NOT NULL,
				symbols_json BLOB NOT NULL
			)
		`,
	}
	for _, statement := range schema {
		if _, err := db.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}

func loadMeta(db *sql.DB) (int, string, error) {
	var (
		version int
		root    string
	)
	err := db.QueryRow(`
		SELECT version, root
		FROM store_meta
		WHERE id = 1
	`).Scan(&version, &root)
	if err != nil {
		return 0, "", err
	}
	return version, root, nil
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
		_, _ = fmt.Fprintf(os.Stderr, "gx: skipping .%s files — install with: gx lang add %s\n", language.PrimaryExtension(languageName), languageName)
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
	gitignoreMatcher := compileIgnore(filepath.Join(root, ".gitignore"))

	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		name := info.Name()
		if name == ".git" {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() && fileExists(filepath.Join(path, ".gx-ignore")) {
			return filepath.SkipDir
		}

		relPath, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil
		}
		relPath = filepath.Clean(relPath)
		if relPath == "." {
			return nil
		}

		if gitignoreMatcher != nil && gitignoreMatcher.MatchesPath(strings.ReplaceAll(relPath, "\\", "/")) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			return nil
		}

		return visit(fileCandidate{
			AbsPath: path,
			RelPath: relPath,
			Info:    info,
		})
	})
}

func compileIgnore(path string) *ignore.GitIgnore {
	matcher, err := ignore.CompileIgnoreFile(path)
	if err != nil {
		return nil
	}
	return matcher
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
			ByteStart: symbol.ByteStart,
			ByteEnd:   symbol.ByteEnd,
			IsTest:    symbol.IsTest,
		})
	}
	return result
}
