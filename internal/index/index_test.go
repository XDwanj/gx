package index

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gx/internal/lang"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	concurrentLoadOrBuildWorkers = 12
	concurrentStoreWriters       = 4
	concurrentStoreReaders       = 8
	concurrentStoreIterations    = 20
)

func TestCachePathUsesGxNamespace(t *testing.T) {
	cachePath := CachePathFor(filepath.Join(string(filepath.Separator), "tmp", "project"))
	if !strings.Contains(cachePath, filepath.Join("gx", "indexes")) {
		t.Fatalf("expected gx cache namespace, got %q", cachePath)
	}
	if strings.Contains(cachePath, filepath.Join("cx", "indexes")) {
		t.Fatalf("unexpected cx cache namespace in %q", cachePath)
	}
	if filepath.Ext(cachePath) != cacheFileExtension {
		t.Fatalf("expected sqlite cache extension %q, got %q", cacheFileExtension, cachePath)
	}
}

func TestWalkSkipsDirectoriesWithGxIgnore(t *testing.T) {
	root := t.TempDir()
	skippedDir := filepath.Join(root, "skip-me")
	keptFile := filepath.Join(root, "keep.go")
	skippedFile := filepath.Join(skippedDir, "ignored.go")

	if err := os.MkdirAll(skippedDir, 0o755); err != nil {
		t.Fatalf("mkdir skipped dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skippedDir, ".gx-ignore"), []byte{}, 0o644); err != nil {
		t.Fatalf("write .gx-ignore: %v", err)
	}
	if err := os.WriteFile(skippedFile, []byte("package skipped\n"), 0o644); err != nil {
		t.Fatalf("write skipped file: %v", err)
	}
	if err := os.WriteFile(keptFile, []byte("package kept\n"), 0o644); err != nil {
		t.Fatalf("write kept file: %v", err)
	}

	visited := map[string]bool{}
	if err := walk(root, func(candidate fileCandidate) error {
		visited[candidate.RelPath] = true
		return nil
	}); err != nil {
		t.Fatalf("walk: %v", err)
	}

	if visited[filepath.Join("skip-me", "ignored.go")] {
		t.Fatalf("expected .gx-ignore directory to be skipped, visited=%v", visited)
	}
	if !visited["keep.go"] {
		t.Fatalf("expected keep.go to be visited, visited=%v", visited)
	}
}

func TestWalkDoesNotTreatCxIgnoreAsSpecial(t *testing.T) {
	root := t.TempDir()
	legacyDir := filepath.Join(root, "legacy")
	legacyFile := filepath.Join(legacyDir, "still-visible.go")

	if err := os.MkdirAll(legacyDir, 0o755); err != nil {
		t.Fatalf("mkdir legacy dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(legacyDir, ".cx-ignore"), []byte{}, 0o644); err != nil {
		t.Fatalf("write .cx-ignore: %v", err)
	}
	if err := os.WriteFile(legacyFile, []byte("package legacy\n"), 0o644); err != nil {
		t.Fatalf("write legacy file: %v", err)
	}

	visited := map[string]bool{}
	if err := walk(root, func(candidate fileCandidate) error {
		visited[candidate.RelPath] = true
		return nil
	}); err != nil {
		t.Fatalf("walk: %v", err)
	}

	if !visited[filepath.Join("legacy", "still-visible.go")] {
		t.Fatalf("expected .cx-ignore to have no effect, visited=%v", visited)
	}
}

func TestNeedsUpdateRebuildsEmptyCacheWhenInstalledLanguageFilesExist(t *testing.T) {
	cacheHome := t.TempDir()
	t.Setenv("HOME", cacheHome)

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}
	if err := lang.Add(os.Stdout, os.Stderr, []string{"go"}); err != nil {
		t.Fatalf("install go grammar: %v", err)
	}

	if !needsUpdate(root, map[string]FileData{}) {
		t.Fatalf("expected empty cache to rebuild once installed-language files exist")
	}
}

func TestNeedsUpdateKeepsEmptyCacheWhenNoInstalledLanguageFilesExist(t *testing.T) {
	cacheHome := t.TempDir()
	t.Setenv("HOME", cacheHome)

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}

	if needsUpdate(root, map[string]FileData{}) {
		t.Fatalf("expected empty cache to remain unchanged when matching grammar is not installed")
	}
}

func TestLoadOrBuildVerboseLogsStages(t *testing.T) {
	cacheHome := t.TempDir()
	t.Setenv("HOME", cacheHome)

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}
	if err := lang.Add(os.Stdout, os.Stderr, []string{"go"}); err != nil {
		t.Fatalf("install go grammar: %v", err)
	}

	var stderr bytes.Buffer
	idx, err := LoadOrBuildWithOptions(root, LoadOptions{
		Verbose: true,
		Stderr:  &stderr,
	})
	if err != nil {
		t.Fatalf("load verbose index: %v", err)
	}
	if len(idx.Entries) != 1 {
		t.Fatalf("expected 1 indexed file, got %d", len(idx.Entries))
	}

	output := stderr.String()
	if !strings.Contains(output, "gx: debug: index root=") {
		t.Fatalf("expected root log, got %q", output)
	}
	if !strings.Contains(output, "gx: debug: starting full index crawl") {
		t.Fatalf("expected crawl log, got %q", output)
	}
	if !strings.Contains(output, "gx: debug: indexing main.go (go)") {
		t.Fatalf("expected file log, got %q", output)
	}
	if !strings.Contains(output, "gx: debug: saved index cache with 1 entries") {
		t.Fatalf("expected save log, got %q", output)
	}

	header, err := os.ReadFile(CachePathFor(root))
	if err != nil {
		t.Fatalf("read sqlite cache: %v", err)
	}
	if len(header) < len("SQLite format 3") || string(header[:len("SQLite format 3")]) != "SQLite format 3" {
		t.Fatalf("expected sqlite file header, got %q", string(header))
	}
}

func TestSaveStoreRoundTripsEntries(t *testing.T) {
	cacheHome := t.TempDir()
	t.Setenv("HOME", cacheHome)

	root := filepath.Join(string(filepath.Separator), "tmp", "project")
	entries := map[string]FileData{
		"main.go": {
			Meta: NewFileEntry(time.Unix(1712000000, 123), "go"),
			Symbols: []Symbol{
				{Name: "main", Kind: SymbolKindFn, Signature: "func main()", ByteStart: 0, ByteEnd: 12},
			},
		},
	}

	if err := saveStore(root, entries); err != nil {
		t.Fatalf("save store: %v", err)
	}

	got, err := loadEntries(CachePathFor(root))
	if err != nil {
		t.Fatalf("load store: %v", err)
	}

	if len(got) != len(entries) {
		t.Fatalf("expected %d entries, got %d", len(entries), len(got))
	}

	data, ok := got["main.go"]
	if !ok {
		t.Fatalf("expected main.go entry, got %v", got)
	}
	if data.Meta.Language != "go" {
		t.Fatalf("expected go language, got %q", data.Meta.Language)
	}
	if len(data.Symbols) != 1 || data.Symbols[0].Name != "main" {
		t.Fatalf("unexpected symbols: %+v", data.Symbols)
	}
}

func TestLoadOrBuildHandlesConcurrentCalls(t *testing.T) {
	cacheHome := t.TempDir()
	t.Setenv("HOME", cacheHome)

	root := t.TempDir()
	files := map[string]string{
		"main.go":   "package main\nfunc main() {}\n",
		"helper.go": "package main\nfunc helper() {}\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	if err := lang.Add(os.Stdout, os.Stderr, []string{"go"}); err != nil {
		t.Fatalf("install go grammar: %v", err)
	}

	start := make(chan struct{})
	errs := make(chan error, concurrentLoadOrBuildWorkers)
	entryCounts := make(chan int, concurrentLoadOrBuildWorkers)

	var wg sync.WaitGroup
	for worker := 0; worker < concurrentLoadOrBuildWorkers; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start

			idx, err := LoadOrBuild(root)
			if err != nil {
				errs <- err
				return
			}
			entryCounts <- len(idx.Entries)
		}()
	}

	close(start)
	wg.Wait()
	close(errs)
	close(entryCounts)

	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent LoadOrBuild failed: %v", err)
		}
	}

	for count := range entryCounts {
		if count != len(files) {
			t.Fatalf("expected %d indexed entries, got %d", len(files), count)
		}
	}

	entries, err := loadEntries(CachePathFor(root))
	if err != nil {
		t.Fatalf("load cached entries after concurrent builds: %v", err)
	}
	if len(entries) != len(files) {
		t.Fatalf("expected persisted %d entries, got %d", len(files), len(entries))
	}
}

func TestStoreHandlesConcurrentReadersAndWriters(t *testing.T) {
	cacheHome := t.TempDir()
	t.Setenv("HOME", cacheHome)

	root := t.TempDir()
	variantA := map[string]FileData{
		"main.go": {
			Meta: NewFileEntry(time.Unix(1712000000, 1), "go"),
			Symbols: []Symbol{
				{Name: "main", Kind: SymbolKindFn, Signature: "func main()", ByteStart: 0, ByteEnd: 12},
			},
		},
		"helper.go": {
			Meta: NewFileEntry(time.Unix(1712000001, 2), "go"),
			Symbols: []Symbol{
				{Name: "helperAlpha", Kind: SymbolKindFn, Signature: "func helperAlpha()", ByteStart: 0, ByteEnd: 20},
			},
		},
	}
	variantB := map[string]FileData{
		"main.go": {
			Meta: NewFileEntry(time.Unix(1712000010, 3), "go"),
			Symbols: []Symbol{
				{Name: "mainBeta", Kind: SymbolKindFn, Signature: "func mainBeta()", ByteStart: 0, ByteEnd: 16},
			},
		},
		"helper.go": {
			Meta: NewFileEntry(time.Unix(1712000011, 4), "go"),
			Symbols: []Symbol{
				{Name: "helperBeta", Kind: SymbolKindFn, Signature: "func helperBeta()", ByteStart: 0, ByteEnd: 18},
			},
		},
	}

	if err := saveStore(root, variantA); err != nil {
		t.Fatalf("seed store: %v", err)
	}

	variantAKey, err := marshalEntries(variantA)
	if err != nil {
		t.Fatalf("marshal variantA: %v", err)
	}
	variantBKey, err := marshalEntries(variantB)
	if err != nil {
		t.Fatalf("marshal variantB: %v", err)
	}

	allowed := map[string]bool{
		variantAKey: true,
		variantBKey: true,
	}
	cachePath := CachePathFor(root)
	start := make(chan struct{})
	errs := make(chan error, concurrentStoreWriters*concurrentStoreIterations+concurrentStoreReaders*concurrentStoreIterations)

	var wg sync.WaitGroup
	for writer := 0; writer < concurrentStoreWriters; writer++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			<-start

			for iteration := 0; iteration < concurrentStoreIterations; iteration++ {
				entries := variantA
				if (writerID+iteration)%2 == 1 {
					entries = variantB
				}
				if saveErr := saveStore(root, entries); saveErr != nil {
					errs <- saveErr
					return
				}
			}
		}(writer)
	}

	for reader := 0; reader < concurrentStoreReaders; reader++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start

			for iteration := 0; iteration < concurrentStoreIterations; iteration++ {
				entries, loadErr := loadEntries(cachePath)
				if loadErr != nil {
					errs <- loadErr
					return
				}
				key, marshalErr := marshalEntries(entries)
				if marshalErr != nil {
					errs <- marshalErr
					return
				}
				if !allowed[key] {
					errs <- fmt.Errorf("observed partial or unknown entries snapshot: %s", key)
					return
				}
			}
		}()
	}

	close(start)
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent store access failed: %v", err)
		}
	}

	finalEntries, err := loadEntries(cachePath)
	if err != nil {
		t.Fatalf("load final entries: %v", err)
	}
	finalKey, err := marshalEntries(finalEntries)
	if err != nil {
		t.Fatalf("marshal final entries: %v", err)
	}
	if !allowed[finalKey] {
		t.Fatalf("unexpected final entries snapshot: %+v", finalEntries)
	}
}

func marshalEntries(entries map[string]FileData) (string, error) {
	bytes, err := json.Marshal(entries)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
