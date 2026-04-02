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
	scope := "src/main.rs"

	if err := service.Symbols(idx, &scope, nil, nil); err != nil {
		t.Fatalf("symbols query: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "{name,kind,signature}:") {
		t.Fatalf("unexpected output: %s", output)
	}
	if !strings.Contains(output, "main,fn") {
		t.Fatalf("missing main symbol: %s", output)
	}
	if strings.Contains(output, "src/main.rs") {
		t.Fatalf("single-file scope should omit file column: %s", output)
	}
}

func TestSymbolsDirectoryScopeOutput(t *testing.T) {
	ensureInstalled(t, "rust")
	root := tempProject(t, map[string]string{
		"src/main.rs":    "fn main() {}\n",
		"src/helper.rs":  "fn helper() {}\n",
		"other/extra.rs": "fn extra() {}\n",
	})

	idx, err := index.LoadOrBuild(root)
	if err != nil {
		t.Fatalf("load index: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	service := &Service{root: root, stdout: &stdout, stderr: &stderr}
	scope := "src"

	if err := service.Symbols(idx, &scope, nil, nil); err != nil {
		t.Fatalf("symbols query: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "{file,name,kind,signature}:") {
		t.Fatalf("directory scope should include file column: %s", output)
	}
	if !strings.Contains(output, "src/main.rs,main,fn") {
		t.Fatalf("missing src/main.rs symbol: %s", output)
	}
	if !strings.Contains(output, "src/helper.rs,helper,fn") {
		t.Fatalf("missing src/helper.rs symbol: %s", output)
	}
	if strings.Contains(output, "other/extra.rs") {
		t.Fatalf("directory scope should exclude files outside scope: %s", output)
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

func TestDefinitionSupportsGlobName(t *testing.T) {
	ensureInstalled(t, "rust")
	root := tempProject(t, map[string]string{
		"src/main.rs": "fn build_runtime() {}\nfn helper() {}\n",
	})

	idx, err := index.LoadOrBuild(root)
	if err != nil {
		t.Fatalf("load index: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	service := &Service{root: root, stdout: &stdout, stderr: &stderr}

	if err := service.Definition(idx, "build*", nil, nil, 200); err != nil {
		t.Fatalf("definition query: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "fn build_runtime()") {
		t.Fatalf("missing glob-matched function body: %s", output)
	}
}

func TestDefinitionScopeFiltersResults(t *testing.T) {
	ensureInstalled(t, "rust")
	root := tempProject(t, map[string]string{
		"src/main.rs":   "fn build_runtime() {}\n",
		"other/main.rs": "fn build_runtime() {}\n",
	})

	idx, err := index.LoadOrBuild(root)
	if err != nil {
		t.Fatalf("load index: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	service := &Service{root: root, stdout: &stdout, stderr: &stderr}
	scope := "src"

	if err := service.Definition(idx, "build*", &scope, nil, 200); err != nil {
		t.Fatalf("definition query: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "file: src/main.rs") {
		t.Fatalf("missing scoped definition: %s", output)
	}
	if strings.Contains(output, "file: other/main.rs") {
		t.Fatalf("scope should exclude definitions outside directory: %s", output)
	}
}

func TestReferencesSupportsGlobName(t *testing.T) {
	ensureInstalled(t, "rust")
	root := tempProject(t, map[string]string{
		"src/main.rs": "fn build_runtime() {}\nfn build_helper() {}\nfn main() { build_runtime(); build_helper(); }\n",
	})

	idx, err := index.LoadOrBuild(root)
	if err != nil {
		t.Fatalf("load index: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	service := &Service{root: root, stdout: &stdout, stderr: &stderr}

	if err := service.References(idx, "build*", nil, false); err != nil {
		t.Fatalf("references query: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "build_runtime") {
		t.Fatalf("missing build_runtime reference: %s", output)
	}
	if !strings.Contains(output, "build_helper") {
		t.Fatalf("missing build_helper reference: %s", output)
	}
}

func TestReferencesScopeFiltersResults(t *testing.T) {
	ensureInstalled(t, "rust")
	root := tempProject(t, map[string]string{
		"src/main.rs":   "fn build_runtime() {}\nfn main() { build_runtime(); }\n",
		"other/main.rs": "fn build_runtime() {}\nfn main() { build_runtime(); }\n",
	})

	idx, err := index.LoadOrBuild(root)
	if err != nil {
		t.Fatalf("load index: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	service := &Service{root: root, stdout: &stdout, stderr: &stderr}
	scope := "src/main.rs"

	if err := service.References(idx, "build*", &scope, false); err != nil {
		t.Fatalf("references query: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "src/main.rs") {
		t.Fatalf("missing scoped reference: %s", output)
	}
	if strings.Contains(output, "other/main.rs") {
		t.Fatalf("scope should exclude references outside file: %s", output)
	}
}

func TestMissingScopeReturnsError(t *testing.T) {
	ensureInstalled(t, "rust")
	root := tempProject(t, map[string]string{
		"src/main.rs": "fn main() {}\n",
	})

	idx, err := index.LoadOrBuild(root)
	if err != nil {
		t.Fatalf("load index: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	service := &Service{root: root, stdout: &stdout, stderr: &stderr}
	scope := "missing"

	err = service.Symbols(idx, &scope, nil, nil)
	if err == nil {
		t.Fatal("expected missing scope to return an error")
	}
	if !strings.Contains(err.Error(), "gx: scope not found: missing") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMarkdownOverviewOutput(t *testing.T) {
	root := tempProject(t, map[string]string{
		"README.md": "# Title\n\n## Section\n\n### Deep Dive\n\n#### Level Four\n\n##### Level Five\n\n###### Level Six\n",
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	service := &Service{root: root, stdout: &stdout, stderr: &stderr}

	if err := service.MarkdownOverview("README.md"); err != nil {
		t.Fatalf("markdown overview: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "{level,heading}:") {
		t.Fatalf("unexpected output: %s", output)
	}
	for _, expected := range []string{
		"1,Title",
		"2,Section",
		"3,Deep Dive",
		"4,Level Four",
		"5,Level Five",
		"6,Level Six",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("missing heading %q in output %s", expected, output)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestMarkdownOverviewIgnoresFencedCodeAndSupportsSetext(t *testing.T) {
	root := tempProject(t, map[string]string{
		"guide.markdown": "Document Title\n===============\n\n```md\n# not a heading\n## still not a heading\n```\n\nSection\n-------\n\n### Real Heading\n",
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	service := &Service{root: root, stdout: &stdout, stderr: &stderr}

	if err := service.MarkdownOverview("guide.markdown"); err != nil {
		t.Fatalf("markdown overview: %v", err)
	}

	output := stdout.String()
	for _, expected := range []string{
		"1,Document Title",
		"2,Section",
		"3,Real Heading",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("missing heading %q in output %s", expected, output)
		}
	}
	if strings.Contains(output, "not a heading") {
		t.Fatalf("fenced code heading should be ignored: %s", output)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}
