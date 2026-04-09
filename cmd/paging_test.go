package cmd

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/XDwanj/gx/internal/app"
)

func TestResolvePageRequestUsesCommandDefaultLimit(t *testing.T) {
	previousCmd := rootCmd
	previousFlags := rootFlags
	rootCmd = newRootCmd()
	rootFlags = app.Flags{}
	t.Cleanup(func() {
		rootCmd = previousCmd
		rootFlags = previousFlags
	})

	page, err := resolvePageRequest(newSymbolsCmd(), defaultSymbolsLimit)
	if err != nil {
		t.Fatalf("resolve page request: %v", err)
	}
	if page.Limit != defaultSymbolsLimit || page.Offset != 0 {
		t.Fatalf("unexpected page request: %+v", page)
	}
}

func TestResolvePageRequestUsesExplicitLimit(t *testing.T) {
	previousCmd := rootCmd
	previousFlags := rootFlags
	rootCmd = newRootCmd()
	rootFlags = app.Flags{}
	t.Cleanup(func() {
		rootCmd = previousCmd
		rootFlags = previousFlags
	})

	if err := rootCmd.PersistentFlags().Set("limit", "25"); err != nil {
		t.Fatalf("set limit flag: %v", err)
	}

	page, err := resolvePageRequest(newSymbolsCmd(), defaultSymbolsLimit)
	if err != nil {
		t.Fatalf("resolve page request: %v", err)
	}
	if page.Limit != 25 || page.Offset != 0 {
		t.Fatalf("unexpected page request: %+v", page)
	}
}

func TestResolvePageRequestBypassesLimitWithAll(t *testing.T) {
	previousCmd := rootCmd
	previousFlags := rootFlags
	rootCmd = newRootCmd()
	rootFlags = app.Flags{}
	t.Cleanup(func() {
		rootCmd = previousCmd
		rootFlags = previousFlags
	})

	if err := rootCmd.PersistentFlags().Set("all", "true"); err != nil {
		t.Fatalf("set all flag: %v", err)
	}
	if err := rootCmd.PersistentFlags().Set("offset", "40"); err != nil {
		t.Fatalf("set offset flag: %v", err)
	}

	page, err := resolvePageRequest(newDefinitionCmd(), defaultDefinitionLimit)
	if err != nil {
		t.Fatalf("resolve page request: %v", err)
	}
	if page.Limit != 0 || page.Offset != 40 {
		t.Fatalf("unexpected page request: %+v", page)
	}
}

func TestResolvePageRequestRejectsInvalidValues(t *testing.T) {
	t.Run("limit", func(t *testing.T) {
		previousCmd := rootCmd
		previousFlags := rootFlags
		rootCmd = newRootCmd()
		rootFlags = app.Flags{}
		t.Cleanup(func() {
			rootCmd = previousCmd
			rootFlags = previousFlags
		})

		if err := rootCmd.PersistentFlags().Set("limit", "0"); err != nil {
			t.Fatalf("set limit flag: %v", err)
		}

		_, err := resolvePageRequest(newSymbolsCmd(), defaultSymbolsLimit)
		if err == nil || !strings.Contains(err.Error(), "--limit must be greater than 0") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("offset", func(t *testing.T) {
		previousCmd := rootCmd
		previousFlags := rootFlags
		rootCmd = newRootCmd()
		rootFlags = app.Flags{}
		t.Cleanup(func() {
			rootCmd = previousCmd
			rootFlags = previousFlags
		})

		if err := rootCmd.PersistentFlags().Set("offset", "-1"); err != nil {
			t.Fatalf("set offset flag: %v", err)
		}

		_, err := resolvePageRequest(newSymbolsCmd(), defaultSymbolsLimit)
		if err == nil || !strings.Contains(err.Error(), "--offset must be greater than or equal to 0") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestSymbolsCommandAppliesDefaultLimitFromRootFlags(t *testing.T) {
	ensureCommandLanguages(t, "rust")
	root := commandProject(t, map[string]string{
		"src/main.rs": buildRustFunctions(101),
	})

	stdout, stderr, exitCode := executeRootFixtureCommand(t, root, "symbols", "src")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", exitCode, stderr)
	}

	var rows []map[string]any
	if err := json.Unmarshal([]byte(stdout), &rows); err != nil {
		t.Fatalf("unmarshal stdout: %v\nstdout=%s", err, stdout)
	}
	if len(rows) != defaultSymbolsLimit {
		t.Fatalf("expected %d rows, got %d", defaultSymbolsLimit, len(rows))
	}
	if !strings.Contains(stderr, "gx: showing 1-100 of 101; narrow query, use --offset 100, or --all") {
		t.Fatalf("expected pagination hint, got %q", stderr)
	}
}

func TestOverviewDirectoryPaginationUsesRootFlagsWhileFileModeIgnoresThem(t *testing.T) {
	ensureCommandLanguages(t, "rust")
	root := commandProject(t, map[string]string{
		"src/a.rs":    "fn alpha() {}\n",
		"src/b.rs":    "fn beta() {}\n",
		"src/main.rs": "fn one() {}\nfn two() {}\n",
	})

	dirStdout, dirStderr, dirExitCode := executeRootFixtureCommand(t, root, "--limit", "1", "overview", "src")
	if dirExitCode != 0 {
		t.Fatalf("expected directory overview exit code 0, got %d with stderr %q", dirExitCode, dirStderr)
	}
	var dirRows []map[string]any
	if err := json.Unmarshal([]byte(dirStdout), &dirRows); err != nil {
		t.Fatalf("unmarshal directory stdout: %v\nstdout=%s", err, dirStdout)
	}
	if len(dirRows) != 1 {
		t.Fatalf("expected one directory overview row, got %d", len(dirRows))
	}
	if !strings.Contains(dirStderr, "gx: showing 1-1 of 3; narrow query, use --offset 1, or --all") {
		t.Fatalf("expected directory pagination hint, got %q", dirStderr)
	}

	fileStdout, fileStderr, fileExitCode := executeRootFixtureCommand(t, root, "--limit", "1", "overview", "src/main.rs")
	if fileExitCode != 0 {
		t.Fatalf("expected file overview exit code 0, got %d with stderr %q", fileExitCode, fileStderr)
	}
	var fileRows []map[string]any
	if err := json.Unmarshal([]byte(fileStdout), &fileRows); err != nil {
		t.Fatalf("unmarshal file stdout: %v\nstdout=%s", err, fileStdout)
	}
	if len(fileRows) != 2 {
		t.Fatalf("expected file overview to ignore pagination limit, got %d rows", len(fileRows))
	}
	if strings.TrimSpace(fileStderr) != "" {
		t.Fatalf("expected empty stderr for file overview, got %q", fileStderr)
	}
}

func TestOverviewMultipleDirectoriesApplyPaginationPerTarget(t *testing.T) {
	ensureCommandLanguages(t, "rust")
	root := commandProject(t, map[string]string{
		"src/a.rs":      "fn alpha() {}\n",
		"src/b.rs":      "fn beta() {}\n",
		"pkg/helper.rs": "fn helper() {}\n",
		"pkg/extra.rs":  "fn extra() {}\n",
	})

	stdout, stderr, exitCode := executeRootFixtureCommand(t, root, "--limit", "1", "overview", "src", "pkg")
	if exitCode != 0 {
		t.Fatalf("expected overview exit code 0, got %d with stderr %q", exitCode, stderr)
	}

	var sections []map[string]any
	if err := json.Unmarshal([]byte(stdout), &sections); err != nil {
		t.Fatalf("unmarshal stdout: %v\nstdout=%s", err, stdout)
	}
	if len(sections) != 2 {
		t.Fatalf("expected two overview sections, got %d", len(sections))
	}
	for _, section := range sections {
		rows, ok := section["rows"].([]any)
		if !ok {
			t.Fatalf("rows should decode as array, got %#v", section["rows"])
		}
		if len(rows) != 1 {
			t.Fatalf("expected one paginated row per section, got %d in %#v", len(rows), section)
		}
	}
	if !strings.Contains(stderr, "gx: src showing 1-1 of 2; narrow query, use --offset 1, or --all") {
		t.Fatalf("expected src pagination hint, got %q", stderr)
	}
	if !strings.Contains(stderr, "gx: pkg showing 1-1 of 2; narrow query, use --offset 1, or --all") {
		t.Fatalf("expected pkg pagination hint, got %q", stderr)
	}
}

func buildRustFunctions(count int) string {
	var builder strings.Builder
	for index := 1; index <= count; index++ {
		builder.WriteString("fn item")
		builder.WriteString(strconv.Itoa(index))
		builder.WriteString("() {}\n")
	}
	return builder.String()
}
