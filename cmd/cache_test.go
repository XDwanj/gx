package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/XDwanj/gx/internal/app"
	"github.com/XDwanj/gx/internal/index"
)

func TestCacheCleanRemovesSQLiteCache(t *testing.T) {
	cacheHome := t.TempDir()
	t.Setenv("HOME", cacheHome)

	root := t.TempDir()
	resolvedRoot, err := app.ResolveRoot(root)
	if err != nil {
		t.Fatalf("resolve root: %v", err)
	}
	cachePath := index.CachePathFor(resolvedRoot)
	if mkdirErr := os.MkdirAll(filepath.Dir(cachePath), 0o755); mkdirErr != nil {
		t.Fatalf("mkdir cache dir: %v", mkdirErr)
	}
	if writeErr := os.WriteFile(cachePath, []byte("SQLite format 3\x00"), 0o644); writeErr != nil {
		t.Fatalf("write cache file: %v", writeErr)
	}

	previousDirectory := rootFlags.Directory
	rootFlags.Directory = root
	t.Cleanup(func() {
		rootFlags.Directory = previousDirectory
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
