package index

import (
	"encoding/json"
	"fmt"
	"gx/internal/language"
	"os"
	"path/filepath"
	"strings"
	"time"

	ignore "github.com/sabhiram/go-gitignore"

	langpkg "gx/internal/lang"
)

const IndexVersion = 5

type Store struct {
	Version int                 `json:"version"`
	Root    string              `json:"root"`
	Entries map[string]FileData `json:"entries"`
}

type Index struct {
	Root    string
	Entries map[string]FileData
}

type fileCandidate struct {
	AbsPath string
	RelPath string
	Info    os.FileInfo
}

func LoadOrBuild(root string) (*Index, error) {
	cleanRoot := filepath.Clean(root)
	cachePath := CachePathFor(cleanRoot)
	entries, err := loadEntries(cachePath)
	if err == nil && !needsUpdate(cleanRoot, entries) {
		return &Index{Root: cleanRoot, Entries: entries}, nil
	}

	idx := &Index{
		Root:    cleanRoot,
		Entries: entries,
	}
	if idx.Entries == nil {
		idx.Entries = map[string]FileData{}
	}

	if len(idx.Entries) == 0 {
		if err := idx.fullCrawl(); err != nil {
			return nil, err
		}
	} else {
		if err := idx.incrementalUpdate(); err != nil {
			return nil, err
		}
	}

	if err := saveStore(cleanRoot, idx.Entries); err != nil {
		return nil, err
	}
	return idx, nil
}

func CachePathFor(root string) string {
	return filepath.Join(cacheDir(), "indexes", sanitizeRoot(root)+".json")
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
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	store := &Store{}
	if err := json.Unmarshal(bytes, store); err != nil {
		return nil, err
	}
	if store.Version != IndexVersion {
		return nil, fmt.Errorf("version mismatch")
	}
	return store.Entries, nil
}

func saveStore(root string, entries map[string]FileData) error {
	path := CachePathFor(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	payload := Store{
		Version: IndexVersion,
		Root:    root,
		Entries: entries,
	}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("gx: failed to encode index: %w", err)
	}
	return os.WriteFile(path, bytes, 0o644)
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

func (idx *Index) fullCrawl() error {
	missingLanguages := map[string]int{}

	return walk(idx.Root, func(candidate fileCandidate) error {
		languageName := language.DetectLanguage(candidate.AbsPath)
		if languageName == "" {
			return nil
		}

		source, err := os.ReadFile(candidate.AbsPath)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "gx: warning: failed to read %s: %v\n", candidate.AbsPath, err)
			return nil
		}

		symbols, err := language.ParseAndExtract(languageName, source, candidate.AbsPath)
		if err != nil {
			if language.IsNotInstalled(err) {
				missingLanguages[languageName]++
				return nil
			}
			return nil
		}

		idx.Entries[candidate.RelPath] = FileData{
			Meta:    NewFileEntry(candidate.Info.ModTime(), languageName),
			Symbols: toIndexSymbols(symbols),
		}
		return nil
	})
}

func (idx *Index) incrementalUpdate() error {
	onDisk := map[string]fileCandidate{}
	missingLanguages := map[string]bool{}

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
			continue
		}

		symbols, err := language.ParseAndExtract(languageName, source, candidate.AbsPath)
		if err != nil {
			if language.IsNotInstalled(err) {
				missingLanguages[languageName] = true
			}
			continue
		}

		idx.Entries[relPath] = FileData{
			Meta:    NewFileEntry(candidate.Info.ModTime(), languageName),
			Symbols: toIndexSymbols(symbols),
		}
	}

	for languageName := range missingLanguages {
		_, _ = fmt.Fprintf(os.Stderr, "gx: skipping .%s files — install with: gx lang add %s\n", language.PrimaryExtension(languageName), languageName)
	}
	return nil
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
