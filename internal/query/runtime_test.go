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
	paths := []string{"src/main.rs"}

	if err := service.Symbols(idx, paths, nil, nil, PageRequest{}); err != nil {
		t.Fatalf("symbols query: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "{file,line,name,kind,signature}:") {
		t.Fatalf("unexpected output: %s", output)
	}
	if !strings.Contains(output, "src/main.rs,1,main,fn") {
		t.Fatalf("missing main symbol: %s", output)
	}
	if !strings.Contains(output, "src/main.rs,2,helper,fn") {
		t.Fatalf("missing helper symbol line: %s", output)
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
	paths := []string{"src"}

	if err := service.Symbols(idx, paths, nil, nil, PageRequest{}); err != nil {
		t.Fatalf("symbols query: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "{file,line,name,kind,signature}:") {
		t.Fatalf("directory scope should include file and line fields: %s", output)
	}
	if !strings.Contains(output, "src/main.rs,1,main,fn") {
		t.Fatalf("missing src/main.rs symbol: %s", output)
	}
	if !strings.Contains(output, "src/helper.rs,1,helper,fn") {
		t.Fatalf("missing src/helper.rs symbol: %s", output)
	}
	if strings.Contains(output, "other/extra.rs") {
		t.Fatalf("directory scope should exclude files outside scope: %s", output)
	}
}

func TestSymbolsMultiplePathsUnionOutput(t *testing.T) {
	ensureInstalled(t, "rust")
	root := tempProject(t, map[string]string{
		"src/main.rs":    "fn main() {}\n",
		"pkg/helper.rs":  "fn helper() {}\n",
		"other/extra.rs": "fn extra() {}\n",
	})

	idx, err := index.LoadOrBuild(root)
	if err != nil {
		t.Fatalf("load index: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	service := &Service{root: root, stdout: &stdout, stderr: &stderr}
	paths := []string{"src/main.rs", "pkg"}

	if err := service.Symbols(idx, paths, nil, nil, PageRequest{}); err != nil {
		t.Fatalf("symbols query: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "{file,line,name,kind,signature}:") {
		t.Fatalf("multiple paths should include file and line fields: %s", output)
	}
	if !strings.Contains(output, "src/main.rs,1,main,fn") {
		t.Fatalf("missing src/main.rs symbol: %s", output)
	}
	if !strings.Contains(output, "pkg/helper.rs,1,helper,fn") {
		t.Fatalf("missing pkg/helper.rs symbol: %s", output)
	}
	if strings.Contains(output, "other/extra.rs") {
		t.Fatalf("multiple paths should exclude files outside filters: %s", output)
	}
}

func TestSymbolsJSONIncludesCoordinates(t *testing.T) {
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
	service := &Service{root: root, stdout: &stdout, stderr: &stderr, json: true}

	if err := service.Symbols(idx, []string{"src/main.rs"}, nil, nil, PageRequest{}); err != nil {
		t.Fatalf("symbols query: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "\"file\": \"src/main.rs\"") {
		t.Fatalf("missing file in json output: %s", output)
	}
	if !strings.Contains(output, "\"line\": 1") {
		t.Fatalf("missing line in json output: %s", output)
	}
	if strings.Contains(output, "\"column\"") {
		t.Fatalf("unexpected column in json output: %s", output)
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

	if err := service.Definition(idx, "main", nil, nil, 200, PageRequest{}); err != nil {
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

	if err := service.Definition(idx, "build*", nil, nil, 200, PageRequest{}); err != nil {
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
	paths := []string{"src"}

	if err := service.Definition(idx, "build*", paths, nil, 200, PageRequest{}); err != nil {
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

func TestDefinitionSortsTypesBeforeFunctionsBeforeMethods(t *testing.T) {
	ensureInstalled(t, "go")
	root := tempProject(t, map[string]string{
		"main.go": "package main\n\ntype buildType struct{}\n\nfunc buildFunc() {}\n\nfunc (b buildType) buildMethod() {}\n",
	})

	idx, err := index.LoadOrBuild(root)
	if err != nil {
		t.Fatalf("load index: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	service := &Service{root: root, stdout: &stdout, stderr: &stderr}

	if err := service.Definition(idx, "build*", nil, nil, 200, PageRequest{}); err != nil {
		t.Fatalf("definition query: %v", err)
	}

	output := stdout.String()
	typeIndex := strings.Index(output, "buildType struct{}")
	functionIndex := strings.Index(output, "func buildFunc() {}")
	methodIndex := strings.Index(output, "func (b buildType) buildMethod() {}")
	if typeIndex == -1 || functionIndex == -1 || methodIndex == -1 {
		t.Fatalf("missing definitions in output: %s", output)
	}
	if typeIndex >= functionIndex || functionIndex >= methodIndex {
		t.Fatalf("unexpected definition order: %s", output)
	}
}

func TestDefinitionSortsSamePriorityByFileAndLine(t *testing.T) {
	ensureInstalled(t, "go")
	root := tempProject(t, map[string]string{
		"z_last.go":  "package main\n\nfunc buildLast() {}\n",
		"a_first.go": "package main\n\nfunc buildFirst() {}\n",
	})

	idx, err := index.LoadOrBuild(root)
	if err != nil {
		t.Fatalf("load index: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	service := &Service{root: root, stdout: &stdout, stderr: &stderr}

	if err := service.Definition(idx, "build*", nil, nil, 200, PageRequest{}); err != nil {
		t.Fatalf("definition query: %v", err)
	}

	output := stdout.String()
	firstFileIndex := strings.Index(output, "file: a_first.go")
	lastFileIndex := strings.Index(output, "file: z_last.go")
	if firstFileIndex == -1 || lastFileIndex == -1 {
		t.Fatalf("missing file headers in output: %s", output)
	}
	if firstFileIndex >= lastFileIndex {
		t.Fatalf("expected a_first.go before z_last.go: %s", output)
	}
}

func TestSymbolsPaginationWritesHintAndLimitsResults(t *testing.T) {
	ensureInstalled(t, "rust")
	root := tempProject(t, map[string]string{
		"src/main.rs": "fn alpha() {}\nfn beta() {}\nfn gamma() {}\n",
	})

	idx, err := index.LoadOrBuild(root)
	if err != nil {
		t.Fatalf("load index: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	service := &Service{root: root, stdout: &stdout, stderr: &stderr}

	if err := service.Symbols(idx, []string{"src/main.rs"}, nil, nil, PageRequest{Limit: 2}); err != nil {
		t.Fatalf("symbols query: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "alpha") || !strings.Contains(output, "beta") {
		t.Fatalf("missing paged symbols: %s", output)
	}
	if strings.Contains(output, "gamma") {
		t.Fatalf("expected symbols pagination to exclude gamma: %s", output)
	}
	if !strings.Contains(stderr.String(), "gx: showing 1-2 of 3; narrow query, use --offset 2, or --all") {
		t.Fatalf("expected pagination hint, got %q", stderr.String())
	}
}

func TestDefinitionPaginationAppliesOffsetAfterPrioritySort(t *testing.T) {
	ensureInstalled(t, "go")
	root := tempProject(t, map[string]string{
		"main.go": "package main\n\ntype buildType struct{}\n\nfunc buildFunc() {}\n\nfunc (b buildType) buildMethod() {}\n",
	})

	idx, err := index.LoadOrBuild(root)
	if err != nil {
		t.Fatalf("load index: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	service := &Service{root: root, stdout: &stdout, stderr: &stderr}

	if err := service.Definition(idx, "build*", nil, nil, 200, PageRequest{Limit: 1, Offset: 1}); err != nil {
		t.Fatalf("definition query: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "func buildFunc() {}") {
		t.Fatalf("expected second sorted definition, got %s", output)
	}
	if strings.Contains(output, "buildType struct{}") || strings.Contains(output, "buildMethod") {
		t.Fatalf("unexpected definition in paged output: %s", output)
	}
	if !strings.Contains(stderr.String(), "gx: showing 2-2 of 3; narrow query, use --offset 2, or --all") {
		t.Fatalf("expected pagination hint, got %q", stderr.String())
	}
}

func TestDirectoryOverviewPaginationWritesHint(t *testing.T) {
	ensureInstalled(t, "rust")
	root := tempProject(t, map[string]string{
		"src/a.rs": "fn alpha() {}\n",
		"src/b.rs": "fn beta() {}\n",
		"src/c.rs": "fn gamma() {}\n",
	})

	idx, err := index.LoadOrBuild(root)
	if err != nil {
		t.Fatalf("load index: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	service := &Service{root: root, stdout: &stdout, stderr: &stderr}

	if err := service.DirectoryOverview(idx, "src", false, PageRequest{Limit: 2}); err != nil {
		t.Fatalf("directory overview query: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "src/a.rs") || !strings.Contains(output, "src/b.rs") {
		t.Fatalf("missing expected directory rows: %s", output)
	}
	if strings.Contains(output, "src/c.rs") {
		t.Fatalf("expected overview pagination to exclude src/c.rs: %s", output)
	}
	if !strings.Contains(stderr.String(), "gx: showing 1-2 of 3; narrow query, use --offset 2, or --all") {
		t.Fatalf("expected pagination hint, got %q", stderr.String())
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

	if err := service.References(idx, "build*", nil, false, PageRequest{}); err != nil {
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
	paths := []string{"src/main.rs"}

	if err := service.References(idx, "build*", paths, false, PageRequest{}); err != nil {
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

func TestMissingPathReturnsError(t *testing.T) {
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
	paths := []string{"missing"}

	err = service.Symbols(idx, paths, nil, nil, PageRequest{})
	if err == nil {
		t.Fatal("expected missing path to return an error")
	}
	if !strings.Contains(err.Error(), "gx: path not found: missing") {
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
