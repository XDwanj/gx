package query

import (
	"bytes"
	"gx/internal/index"
	"gx/internal/lang"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func ensureInstalled(t *testing.T, languages ...string) {
	t.Helper()
	if err := lang.Add(io.Discard, io.Discard, languages); err != nil {
		t.Fatalf("install grammars: %v", err)
	}
}

func tempProject(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	for path, content := range files {
		fullPath := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("mkdir parents: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("write file: %v", err)
		}
	}
	return root
}

func TestSymbolsSingleFileOutput(t *testing.T) {
	ensureInstalled(t, "rust")
	root := tempProject(t, map[string]string{
		"src/main.rs": "fn main() {}\nfn helper() {}\n",
	})

	idx, err := index.LoadOrBuild(root)
	if err != nil {
		t.Fatalf("load index: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	service := &Service{root: root, stdout: &stdout, stderr: &stderr}
	file := "src/main.rs"

	if err := service.Symbols(idx, &file, nil, nil); err != nil {
		t.Fatalf("symbols query: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "{name,kind,signature}:") {
		t.Fatalf("unexpected output: %s", output)
	}
	if !strings.Contains(output, "main,fn") {
		t.Fatalf("missing main symbol: %s", output)
	}
}

func TestDefinitionOutput(t *testing.T) {
	ensureInstalled(t, "rust")
	root := tempProject(t, map[string]string{
		"src/main.rs": "fn main() {\n    helper();\n}\n\nfn helper() {}\n",
	})

	idx, err := index.LoadOrBuild(root)
	if err != nil {
		t.Fatalf("load index: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	service := &Service{root: root, stdout: &stdout, stderr: &stderr}

	if err := service.Definition(idx, "main", nil, nil, 200); err != nil {
		t.Fatalf("definition query: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "file: src/main.rs") {
		t.Fatalf("missing file header: %s", output)
	}
	if !strings.Contains(output, "fn main()") {
		t.Fatalf("missing function body: %s", output)
	}
}
