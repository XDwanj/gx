package cmd

import (
	"gx/internal/app"
	"gx/internal/lang"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestOverviewCommandDefaultsToCurrentDirectory(t *testing.T) {
	ensureCommandLanguages(t, "rust")
	root := commandProject(t, map[string]string{
		"src/main.rs": "fn main() {}\n",
	})
	t.Chdir(filepath.Join(root, "src"))

	previousCmd := rootCmd
	previousFlags := rootFlags
	rootCmd = newRootCmd()
	rootFlags = app.Flags{Root: root}
	t.Cleanup(func() {
		rootCmd = previousCmd
		rootFlags = previousFlags
	})

	var runErr error
	stdout, stderr := captureProcessOutput(t, func() {
		command := newOverviewCmd()
		runErr = command.RunE(command, nil)
	})
	if runErr != nil {
		t.Fatalf("run overview command: %v", runErr)
	}
	if !strings.Contains(stdout, "src/main.rs") {
		t.Fatalf("expected current directory overview, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func TestSymbolsCommandDefaultsToCurrentDirectory(t *testing.T) {
	ensureCommandLanguages(t, "rust")
	root := commandProject(t, map[string]string{
		"src/main.rs":    "fn main() {}\n",
		"other/extra.rs": "fn extra() {}\n",
	})
	t.Chdir(filepath.Join(root, "src"))

	previousCmd := rootCmd
	previousFlags := rootFlags
	rootCmd = newRootCmd()
	rootFlags = app.Flags{Root: root}
	t.Cleanup(func() {
		rootCmd = previousCmd
		rootFlags = previousFlags
	})

	var runErr error
	stdout, stderr := captureProcessOutput(t, func() {
		command := newSymbolsCmd()
		runErr = command.RunE(command, nil)
	})
	if runErr != nil {
		t.Fatalf("run symbols command: %v", runErr)
	}
	if !strings.Contains(stdout, "src/main.rs,main,fn") {
		t.Fatalf("expected current directory symbols, got %q", stdout)
	}
	if strings.Contains(stdout, "other/extra.rs") {
		t.Fatalf("expected current directory filter, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func TestSymbolsCommandSupportsMultiplePathArgs(t *testing.T) {
	ensureCommandLanguages(t, "rust")
	root := commandProject(t, map[string]string{
		"src/main.rs":    "fn main() {}\n",
		"pkg/helper.rs":  "fn helper() {}\n",
		"other/extra.rs": "fn extra() {}\n",
	})
	t.Chdir(root)

	previousCmd := rootCmd
	previousFlags := rootFlags
	rootCmd = newRootCmd()
	rootFlags = app.Flags{Root: root}
	t.Cleanup(func() {
		rootCmd = previousCmd
		rootFlags = previousFlags
	})

	var runErr error
	stdout, stderr := captureProcessOutput(t, func() {
		command := newSymbolsCmd()
		runErr = command.RunE(command, []string{"src", "pkg/helper.rs"})
	})
	if runErr != nil {
		t.Fatalf("run symbols command: %v", runErr)
	}
	if !strings.Contains(stdout, "src/main.rs,main,fn") {
		t.Fatalf("expected src/main.rs symbol, got %q", stdout)
	}
	if !strings.Contains(stdout, "pkg/helper.rs,helper,fn") {
		t.Fatalf("expected pkg/helper.rs symbol, got %q", stdout)
	}
	if strings.Contains(stdout, "other/extra.rs") {
		t.Fatalf("expected path args to filter results, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func TestCommandsRemoveScopeFlag(t *testing.T) {
	for _, command := range []*cobra.Command{
		newSymbolsCmd(),
		newDefinitionCmd(),
		newReferencesCmd(),
	} {
		if command.Flags().Lookup("scope") != nil {
			t.Fatalf("%s should not expose --scope", command.Name())
		}
	}
}

func ensureCommandLanguages(t *testing.T, languages ...string) {
	t.Helper()
	if err := lang.Add(io.Discard, io.Discard, languages); err != nil {
		t.Fatalf("install grammars: %v", err)
	}
}

func commandProject(t *testing.T, files map[string]string) string {
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
