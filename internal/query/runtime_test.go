package query

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/XDwanj/gx/internal/index"
	"github.com/XDwanj/gx/internal/lang"
)

func ensureInstalled(t testing.TB, languages ...string) {
	t.Helper()
	if err := lang.Add(io.Discard, io.Discard, languages); err != nil {
		t.Fatalf("install grammars: %v", err)
	}
}

func tempProject(t testing.TB, files map[string]string) string {
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

func symbolsOptions(paths []string, nameGlob *string, kind *index.SymbolKind, page PageRequest) SymbolsOptions {
	return SymbolsOptions{
		Paths:    PathQuery{Targets: paths},
		NameGlob: nameGlob,
		Kind:     kind,
		Page:     page,
	}
}

func definitionOptions(paths []string, nameGlob string, kind *index.SymbolKind, maxLines int, page PageRequest) DefinitionOptions {
	return DefinitionOptions{
		Paths:    PathQuery{Targets: paths},
		NameGlob: nameGlob,
		Kind:     kind,
		MaxLines: maxLines,
		Page:     page,
	}
}

func referencesOptions(paths []string, nameGlob string, unique bool, page PageRequest) ReferencesOptions {
	return ReferencesOptions{
		Paths:    PathQuery{Targets: paths},
		NameGlob: nameGlob,
		Unique:   unique,
		Page:     page,
	}
}

func calleesOptions(paths []string, nameGlob string, page PageRequest) CalleesOptions {
	return CalleesOptions{
		Paths:    PathQuery{Targets: paths},
		NameGlob: nameGlob,
		Page:     page,
	}
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

	if err := service.Symbols(idx, symbolsOptions(paths, nil, nil, PageRequest{})); err != nil {
		t.Fatalf("symbols query: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "{file,line,name,kind,signature}:") {
		t.Fatalf("unexpected output: %s", output)
	}
	if !strings.Contains(output, "src/main.rs,1,main,func") {
		t.Fatalf("missing main symbol: %s", output)
	}
	if !strings.Contains(output, "src/main.rs,2,helper,func") {
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

	if err := service.Symbols(idx, symbolsOptions(paths, nil, nil, PageRequest{})); err != nil {
		t.Fatalf("symbols query: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "{file,line,name,kind,signature}:") {
		t.Fatalf("directory scope should include file and line fields: %s", output)
	}
	if !strings.Contains(output, "src/main.rs,1,main,func") {
		t.Fatalf("missing src/main.rs symbol: %s", output)
	}
	if !strings.Contains(output, "src/helper.rs,1,helper,func") {
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

	if err := service.Symbols(idx, symbolsOptions(paths, nil, nil, PageRequest{})); err != nil {
		t.Fatalf("symbols query: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "{file,line,name,kind,signature}:") {
		t.Fatalf("multiple paths should include file and line fields: %s", output)
	}
	if !strings.Contains(output, "src/main.rs,1,main,func") {
		t.Fatalf("missing src/main.rs symbol: %s", output)
	}
	if !strings.Contains(output, "pkg/helper.rs,1,helper,func") {
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

	if err := service.Symbols(idx, symbolsOptions([]string{"src/main.rs"}, nil, nil, PageRequest{})); err != nil {
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

func TestSymbolsJSONNoMatchesOutputsEmptyArray(t *testing.T) {
	ensureInstalled(t, "rust")
	root := tempProject(t, map[string]string{
		"src/main.rs": "fn main() {}\n",
	})

	idx, err := index.LoadOrBuild(root)
	if err != nil {
		t.Fatalf("load index: %v", err)
	}

	name := "Hidden"
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	service := &Service{root: root, stdout: &stdout, stderr: &stderr, json: true}

	if err := service.Symbols(idx, symbolsOptions([]string{"src/main.rs"}, &name, nil, PageRequest{})); err != nil {
		t.Fatalf("symbols query: %v", err)
	}

	if stdout.String() != "[]\n" {
		t.Fatalf("expected empty json array for no matches, got %q", stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("expected empty stderr for json no matches, got %q", stderr.String())
	}
}

func TestSymbolsSupportsPipeSeparatedAlternatives(t *testing.T) {
	ensureInstalled(t, "go")
	root := tempProject(t, map[string]string{
		"main.go": "package main\n\nfunc WechatPay() {}\nfunc AliPay() {}\nfunc StripePay() {}\n",
	})

	idx, err := index.LoadOrBuild(root)
	if err != nil {
		t.Fatalf("load index: %v", err)
	}

	pattern := "WechatPay|AliPay"
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	service := &Service{root: root, stdout: &stdout, stderr: &stderr}
	if err := service.Symbols(idx, symbolsOptions([]string{"main.go"}, &pattern, nil, PageRequest{})); err != nil {
		t.Fatalf("symbols query: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "main.go,3,WechatPay,func") {
		t.Fatalf("missing WechatPay symbol: %s", output)
	}
	if !strings.Contains(output, "main.go,4,AliPay,func") {
		t.Fatalf("missing AliPay symbol: %s", output)
	}
	if strings.Contains(output, "StripePay") {
		t.Fatalf("unexpected symbol matched by alternation: %s", output)
	}
}

func TestSymbolsRejectsPipePatternWithEmptyAlternative(t *testing.T) {
	ensureInstalled(t, "go")
	root := tempProject(t, map[string]string{
		"main.go": "package main\n\nfunc WechatPay() {}\nfunc AliPay() {}\n",
	})

	idx, err := index.LoadOrBuild(root)
	if err != nil {
		t.Fatalf("load index: %v", err)
	}

	pattern := "WechatPay||AliPay"
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	service := &Service{root: root, stdout: &stdout, stderr: &stderr}
	err = service.Symbols(idx, symbolsOptions([]string{"main.go"}, &pattern, nil, PageRequest{}))
	if err == nil {
		t.Fatal("expected invalid pattern error")
	}
	if !strings.Contains(err.Error(), "empty alternative") {
		t.Fatalf("unexpected error: %v", err)
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

	if err := service.Definition(idx, definitionOptions(nil, "main", nil, 200, PageRequest{})); err != nil {
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

	if err := service.Definition(idx, definitionOptions(nil, "build*", nil, 200, PageRequest{})); err != nil {
		t.Fatalf("definition query: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "fn build_runtime()") {
		t.Fatalf("missing glob-matched function body: %s", output)
	}
}

func TestDefinitionSupportsPipeSeparatedAlternatives(t *testing.T) {
	ensureInstalled(t, "go")
	root := tempProject(t, map[string]string{
		"main.go": "package main\n\nfunc WechatPay() {}\nfunc AliPay() {}\nfunc StripePay() {}\n",
	})

	idx, err := index.LoadOrBuild(root)
	if err != nil {
		t.Fatalf("load index: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	service := &Service{root: root, stdout: &stdout, stderr: &stderr}
	if err := service.Definition(idx, definitionOptions([]string{"main.go"}, "WechatPay|AliPay", nil, 200, PageRequest{})); err != nil {
		t.Fatalf("definition query: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "func WechatPay()") {
		t.Fatalf("missing WechatPay definition: %s", output)
	}
	if !strings.Contains(output, "func AliPay()") {
		t.Fatalf("missing AliPay definition: %s", output)
	}
	if strings.Contains(output, "func StripePay()") {
		t.Fatalf("unexpected StripePay definition: %s", output)
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

	if err := service.Definition(idx, definitionOptions(paths, "build*", nil, 200, PageRequest{})); err != nil {
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

	if err := service.Definition(idx, definitionOptions(nil, "build*", nil, 200, PageRequest{})); err != nil {
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

	if err := service.Definition(idx, definitionOptions(nil, "build*", nil, 200, PageRequest{})); err != nil {
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

	if err := service.Symbols(idx, symbolsOptions([]string{"src/main.rs"}, nil, nil, PageRequest{Limit: 2})); err != nil {
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

	if err := service.Definition(idx, definitionOptions(nil, "build*", nil, 200, PageRequest{Limit: 1, Offset: 1})); err != nil {
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

func TestDefinitionMaxLinesTruncationMentionsHowToContinue(t *testing.T) {
	ensureInstalled(t, "go")
	root := tempProject(t, map[string]string{
		"main.go": "package main\n\ntype buildType struct {\n\tAlpha int\n\tBeta int\n\tGamma int\n}\n",
	})

	idx, err := index.LoadOrBuild(root)
	if err != nil {
		t.Fatalf("load index: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	service := &Service{root: root, stdout: &stdout, stderr: &stderr}

	if err := service.Definition(idx, definitionOptions(nil, "buildType", nil, 3, PageRequest{})); err != nil {
		t.Fatalf("definition query: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "truncated: showing first 3 of 5 lines") {
		t.Fatalf("expected truncation marker, got %s", output)
	}
	if strings.Contains(output, "Gamma int") {
		t.Fatalf("expected --max-lines truncation to omit trailing lines: %s", output)
	}
	if !strings.Contains(output, "--max-lines 5") {
		t.Fatalf("expected truncation hint to mention --max-lines so unfinished output is explicit, got %s", output)
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

	if err := service.References(idx, referencesOptions(nil, "build*", false, PageRequest{})); err != nil {
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

func TestReferencesSupportsPipeSeparatedAlternatives(t *testing.T) {
	ensureInstalled(t, "go")
	root := tempProject(t, map[string]string{
		"main.go": "package main\n\nfunc WechatPay() {}\nfunc AliPay() {}\nfunc StripePay() {}\n\nfunc main() {\n\tWechatPay()\n\tAliPay()\n\tStripePay()\n}\n",
	})

	idx, err := index.LoadOrBuild(root)
	if err != nil {
		t.Fatalf("load index: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	service := &Service{root: root, stdout: &stdout, stderr: &stderr}
	if err := service.References(idx, referencesOptions([]string{"main.go"}, "WechatPay|AliPay", false, PageRequest{})); err != nil {
		t.Fatalf("references query: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "WechatPay") {
		t.Fatalf("missing WechatPay references: %s", output)
	}
	if !strings.Contains(output, "AliPay") {
		t.Fatalf("missing AliPay references: %s", output)
	}
	if strings.Contains(output, "StripePay") {
		t.Fatalf("unexpected StripePay references: %s", output)
	}
}

func TestReferencesPaginationWritesHint(t *testing.T) {
	ensureInstalled(t, "rust")
	root := tempProject(t, map[string]string{
		"src/main.rs": "fn build_runtime() {}\nfn one() { build_runtime(); }\nfn two() { build_runtime(); }\nfn three() { build_runtime(); }\n",
	})

	idx, err := index.LoadOrBuild(root)
	if err != nil {
		t.Fatalf("load index: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	service := &Service{root: root, stdout: &stdout, stderr: &stderr}

	if err := service.References(idx, referencesOptions(nil, "build*", false, PageRequest{Limit: 2})); err != nil {
		t.Fatalf("references query: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "build_runtime") {
		t.Fatalf("missing paged references output: %s", output)
	}
	if !strings.Contains(stderr.String(), "gx: showing 1-2 of ") {
		t.Fatalf("expected references pagination hint, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "use --offset 2, or --all") {
		t.Fatalf("expected references pagination hint to suggest next offset, got %q", stderr.String())
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

	if err := service.References(idx, referencesOptions(paths, "build*", false, PageRequest{})); err != nil {
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

func TestReferencesFindsExternalSymbolUsages(t *testing.T) {
	ensureInstalled(t, "go")
	root := tempProject(t, map[string]string{
		"main.go": "package main\n\nimport helper \"example.com/helper\"\n\nfunc main() {\n\thelper.SplitCutset(\"a,b\", \",\")\n}\n",
	})

	idx, err := index.LoadOrBuild(root)
	if err != nil {
		t.Fatalf("load index: %v", err)
	}

	matchedNames, err := findMatchingSymbolNames(idx, "SplitCutset")
	if err != nil {
		t.Fatalf("match local declarations: %v", err)
	}
	if len(matchedNames) != 0 {
		t.Fatalf("expected no local declarations, got %v", matchedNames)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	service := &Service{root: root, stdout: &stdout, stderr: &stderr}

	if err := service.References(idx, referencesOptions([]string{"main.go"}, "SplitCutset", false, PageRequest{})); err != nil {
		t.Fatalf("references query: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "main.go,6") {
		t.Fatalf("missing external reference row: %s", output)
	}
	if !strings.Contains(output, "SplitCutset") {
		t.Fatalf("missing external reference context: %s", output)
	}
	if strings.Contains(stderr.String(), "gx: no matches") {
		t.Fatalf("unexpected no matches message: %q", stderr.String())
	}
}

func BenchmarkReferencesAlternation(b *testing.B) {
	ensureInstalled(b, "go")

	var source strings.Builder
	source.WriteString("package main\n\n")
	for nameIndex := range 12 {
		_, _ = source.WriteString("func RefTarget")
		_, _ = source.WriteString(strconv.Itoa(nameIndex))
		_, _ = source.WriteString("() {}\n")
	}
	_, _ = source.WriteString("\nfunc main() {\n")
	for repeatIndex := range 80 {
		for nameIndex := range 12 {
			_, _ = source.WriteString("\tRefTarget")
			_, _ = source.WriteString(strconv.Itoa(nameIndex))
			_, _ = source.WriteString("()\n")
		}
		if repeatIndex%10 == 0 {
			_, _ = source.WriteString("\tprintln(\"marker\")\n")
		}
	}
	_, _ = source.WriteString("}\n")

	root := tempProject(b, map[string]string{
		"main.go": source.String(),
	})

	idx, err := index.LoadOrBuild(root)
	if err != nil {
		b.Fatalf("load index: %v", err)
	}

	service := &Service{
		root:   root,
		stdout: io.Discard,
		stderr: io.Discard,
	}
	namePattern := "RefTarget{0,1,2,3,4,5,6,7,8,9,10,11}"

	b.ResetTimer()
	for range b.N {
		if err := service.References(idx, referencesOptions([]string{"main.go"}, namePattern, false, PageRequest{})); err != nil {
			b.Fatalf("references query: %v", err)
		}
	}
}

func TestCalleesReturnsCallSites(t *testing.T) {
	ensureInstalled(t, "go")
	root := tempProject(t, map[string]string{
		"main.go": "package main\n\nimport \"fmt\"\n\nfunc B() {}\nfunc C() {}\n\nfunc A() {\n\tB()\n\tC()\n\tfmt.Println(\"hello\")\n}\n",
	})

	idx, err := index.LoadOrBuild(root)
	if err != nil {
		t.Fatalf("load index: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	service := &Service{root: root, stdout: &stdout, stderr: &stderr}

	if err := service.Callees(idx, calleesOptions(nil, "A", PageRequest{})); err != nil {
		t.Fatalf("callees query: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "{file,line,caller,callee,context}:") {
		t.Fatalf("unexpected output shape: %s", output)
	}
	for _, expected := range []string{
		"main.go,9,A,B,B()",
		"main.go,10,A,C,C()",
		"main.go,11,A,fmt.Println,\"fmt.Println(\\\"hello\\\")\"",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("missing callee row %q in %s", expected, output)
		}
	}
	if strings.TrimSpace(stderr.String()) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestCalleesPaginationWritesHint(t *testing.T) {
	ensureInstalled(t, "go")
	root := tempProject(t, map[string]string{
		"main.go": "package main\n\nfunc one() {}\nfunc two() {}\nfunc three() {}\n\nfunc A() {\n\tone()\n\ttwo()\n\tthree()\n}\n",
	})

	idx, err := index.LoadOrBuild(root)
	if err != nil {
		t.Fatalf("load index: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	service := &Service{root: root, stdout: &stdout, stderr: &stderr}

	if err := service.Callees(idx, calleesOptions(nil, "A", PageRequest{Limit: 2})); err != nil {
		t.Fatalf("callees query: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "one()") || !strings.Contains(output, "two()") {
		t.Fatalf("missing paged callees output: %s", output)
	}
	if strings.Contains(output, "three()") {
		t.Fatalf("expected pagination to exclude third callee: %s", output)
	}
	if !strings.Contains(stderr.String(), "gx: showing 1-2 of 3; narrow query, use --offset 2, or --all") {
		t.Fatalf("expected pagination hint, got %q", stderr.String())
	}
}

func TestCalleesScopeFiltersResults(t *testing.T) {
	ensureInstalled(t, "go")
	root := tempProject(t, map[string]string{
		"src/main.go":   "package main\n\nfunc left() {}\nfunc A() { left() }\n",
		"other/main.go": "package main\n\nfunc right() {}\nfunc A() { right() }\n",
	})

	idx, err := index.LoadOrBuild(root)
	if err != nil {
		t.Fatalf("load index: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	service := &Service{root: root, stdout: &stdout, stderr: &stderr}

	if err := service.Callees(idx, calleesOptions([]string{"src/main.go"}, "A", PageRequest{})); err != nil {
		t.Fatalf("callees query: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "src/main.go") || !strings.Contains(output, "left") {
		t.Fatalf("missing scoped callee result: %s", output)
	}
	if strings.Contains(output, "other/main.go") || strings.Contains(output, "right") {
		t.Fatalf("scope should exclude other file: %s", output)
	}
	if strings.TrimSpace(stderr.String()) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
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

	err = service.Symbols(idx, symbolsOptions(paths, nil, nil, PageRequest{}))
	if err == nil {
		t.Fatal("expected missing path to return an error")
	}
	if !strings.Contains(err.Error(), "gx: path not found: missing") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMissingPathsReturnCombinedError(t *testing.T) {
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
	paths := []string{"missing", "also-missing"}

	err = service.Symbols(idx, symbolsOptions(paths, nil, nil, PageRequest{}))
	if err == nil {
		t.Fatal("expected missing paths to return an error")
	}
	if !strings.Contains(err.Error(), "gx: paths not found: missing, also-missing") {
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
