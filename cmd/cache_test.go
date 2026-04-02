package cmd

import (
	"bytes"
	"gx/internal/index"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCacheCleanRemovesSQLiteCache(t *testing.T) {
	cacheHome := t.TempDir()
	t.Setenv("HOME", cacheHome)

	root := t.TempDir()
	cachePath := index.CachePathFor(root)
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		t.Fatalf("mkdir cache dir: %v", err)
	}
	if err := os.WriteFile(cachePath, []byte("SQLite format 3\x00"), 0o644); err != nil {
		t.Fatalf("write cache file: %v", err)
	}

	previousRoot := rootFlags.Root
	rootFlags.Root = root
	t.Cleanup(func() {
		rootFlags.Root = previousRoot
	})

	command := newCacheCmd()
	cleanCmd, _, err := command.Find([]string{"clean"})
	if err != nil {
		t.Fatalf("find clean command: %v", err)
	}

	var stderr bytes.Buffer
	cleanCmd.SetErr(&stderr)
	if err := cleanCmd.RunE(cleanCmd, nil); err != nil {
		t.Fatalf("run clean: %v", err)
	}

	if _, err := os.Stat(cachePath); !os.IsNotExist(err) {
		t.Fatalf("expected cache file to be removed, stat err=%v", err)
	}
	if !strings.Contains(stderr.String(), "gx: removed "+cachePath) {
		t.Fatalf("expected removal message, got %q", stderr.String())
	}
}
