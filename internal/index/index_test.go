package index

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeGoFile(t *testing.T, root string, name string, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadOrBuildIndexesGoFiles(t *testing.T) {
	root := t.TempDir()
	writeGoFile(t, root, "main.go", "package main\n\nfunc Hello() {}\n")
	idx, err := LoadOrBuild(root)
	if err != nil {
		t.Fatalf("LoadOrBuild: %v", err)
	}
	data, ok := idx.Entries["main.go"]
	if !ok {
		t.Fatalf("main.go missing from entries: %+v", idx.Entries)
	}
	if len(data.Symbols) != 1 || data.Symbols[0].Name != "Hello" {
		t.Fatalf("symbols %+v", data.Symbols)
	}
}

func TestLoadOrBuildIncrementalUpdate(t *testing.T) {
	root := t.TempDir()
	writeGoFile(t, root, "main.go", "package main\n\nfunc Hello() {}\n")
	if _, err := LoadOrBuild(root); err != nil {
		t.Fatalf("LoadOrBuild: %v", err)
	}
	writeGoFile(t, root, "main.go", "package main\n\nfunc Hello() {}\nfunc World() {}\n")
	future := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(filepath.Join(root, "main.go"), future, future); err != nil {
		t.Fatal(err)
	}
	idx, err := LoadOrBuild(root)
	if err != nil {
		t.Fatalf("LoadOrBuild after change: %v", err)
	}
	if len(idx.Entries["main.go"].Symbols) != 2 {
		t.Fatalf("changed file should be re-indexed, symbols %+v", idx.Entries["main.go"].Symbols)
	}
}

func TestLoadOrBuildRespectsGitignore(t *testing.T) {
	root := t.TempDir()
	writeGoFile(t, root, "main.go", "package main\n\nfunc Hello() {}\n")
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("ignored.go\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeGoFile(t, root, "ignored.go", "package main\n\nfunc Ignored() {}\n")
	idx, err := LoadOrBuild(root)
	if err != nil {
		t.Fatalf("LoadOrBuild: %v", err)
	}
	if _, ok := idx.Entries["ignored.go"]; ok {
		t.Fatal("gitignored file should not be indexed")
	}
	if _, ok := idx.Entries["main.go"]; !ok {
		t.Fatal("main.go should be indexed")
	}
}
