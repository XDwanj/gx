package index

import (
	"gx/internal/lang"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCachePathUsesGxNamespace(t *testing.T) {
	cachePath := CachePathFor(filepath.Join(string(filepath.Separator), "tmp", "project"))
	if !strings.Contains(cachePath, filepath.Join("gx", "indexes")) {
		t.Fatalf("expected gx cache namespace, got %q", cachePath)
	}
	if strings.Contains(cachePath, filepath.Join("cx", "indexes")) {
		t.Fatalf("unexpected cx cache namespace in %q", cachePath)
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
