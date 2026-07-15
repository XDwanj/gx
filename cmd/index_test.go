package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/XDwanj/gx/internal/app"
	"github.com/XDwanj/gx/internal/index"
)

func TestIndexCommandBuildsProjectIndex(t *testing.T) {
	cacheHome := t.TempDir()
	t.Setenv("HOME", cacheHome)
	ensureCommandLanguages(t, "go")
	root := commandProject(t, map[string]string{
		"main.go": "package main\n\nfunc main() {}\n",
	})

	stdout, stderr, exitCode := executeRootForTest(t, "--json", "-C", root, "index")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", exitCode, stderr)
	}
	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}

	var result app.IndexResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("decode index result: %v", err)
	}
	resolvedRoot, err := app.ResolveRoot(root)
	if err != nil {
		t.Fatalf("resolve root: %v", err)
	}
	if result.Root != resolvedRoot {
		t.Fatalf("expected root %q, got %q", resolvedRoot, result.Root)
	}
	if result.Cache != index.CachePathFor(resolvedRoot) {
		t.Fatalf("expected cache path %q, got %q", index.CachePathFor(resolvedRoot), result.Cache)
	}
	if result.IndexedFiles != 1 {
		t.Fatalf("expected one indexed file, got %d", result.IndexedFiles)
	}
	if _, err := os.Stat(result.Cache); err != nil {
		t.Fatalf("expected index cache to exist: %v", err)
	}
}

func TestIndexCommandPrintsTOONResult(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ensureCommandLanguages(t, "go")
	root := commandProject(t, map[string]string{
		"main.go": "package main\n",
	})

	stdout, stderr, exitCode := executeRootForTest(t, "-C", root, "index")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", exitCode, stderr)
	}
	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	for _, expected := range []string{"root", "cache", "indexed_files", filepath.Clean(root)} {
		if !strings.Contains(stdout, expected) {
			t.Fatalf("expected output to contain %q, got %q", expected, stdout)
		}
	}
}

func TestIndexCommandRejectsPaths(t *testing.T) {
	stdout, stderr, exitCode := executeRootForTest(t, "index", ".")
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if stdout != "" {
		t.Fatalf("expected empty stdout, got %q", stdout)
	}
	if !strings.Contains(stderr, "unknown command") {
		t.Fatalf("expected no-args error, got %q", stderr)
	}
}
