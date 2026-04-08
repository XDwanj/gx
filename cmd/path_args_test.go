package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/XDwanj/gx/internal/app"
	"github.com/XDwanj/gx/internal/lang"

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
	rootFlags = app.Flags{}
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
	rootFlags = app.Flags{}
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
	if !strings.Contains(stdout, "src/main.rs,1,main,fn") {
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
	rootFlags = app.Flags{Directory: root}
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
	if !strings.Contains(stdout, "src/main.rs,1,main,fn") {
		t.Fatalf("expected src/main.rs symbol, got %q", stdout)
	}
	if !strings.Contains(stdout, "pkg/helper.rs,1,helper,fn") {
		t.Fatalf("expected pkg/helper.rs symbol, got %q", stdout)
	}
	if strings.Contains(stdout, "other/extra.rs") {
		t.Fatalf("expected path args to filter results, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func TestOverviewCommandResolvesRelativePathAgainstRelativeSymlinkRoot(t *testing.T) {
	ensureCommandLanguages(t, "rust")
	parentDir := t.TempDir()
	targetRoot := filepath.Join(parentDir, "target")
	if err := os.MkdirAll(targetRoot, 0o755); err != nil {
		t.Fatalf("mkdir target root: %v", err)
	}
	if err := os.Mkdir(filepath.Join(targetRoot, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	mainFile := filepath.Join(targetRoot, "src", "main.rs")
	if err := os.MkdirAll(filepath.Dir(mainFile), 0o755); err != nil {
		t.Fatalf("mkdir parents: %v", err)
	}
	if err := os.WriteFile(mainFile, []byte("fn main() {}\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	targetLink := filepath.Join(parentDir, "target_link")
	if err := os.Symlink(targetRoot, targetLink); err != nil {
		t.Fatalf("create symlink root: %v", err)
	}
	callerDir := filepath.Join(parentDir, "caller")
	if err := os.MkdirAll(callerDir, 0o755); err != nil {
		t.Fatalf("mkdir caller: %v", err)
	}
	t.Chdir(callerDir)

	previousCmd := rootCmd
	previousFlags := rootFlags
	rootCmd = newRootCmd()
	rootFlags = app.Flags{Directory: "../target_link"}
	t.Cleanup(func() {
		rootCmd = previousCmd
		rootFlags = previousFlags
	})

	var runErr error
	stdout, stderr := captureProcessOutput(t, func() {
		command := newOverviewCmd()
		runErr = command.RunE(command, []string{"."})
	})
	if runErr != nil {
		t.Fatalf("run overview command: %v", runErr)
	}
	if !strings.Contains(stdout, "src/,") {
		t.Fatalf("expected explicit root overview, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func TestOverviewCommandDefaultsToExplicitDirectory(t *testing.T) {
	ensureCommandLanguages(t, "rust")
	targetRoot := commandProject(t, map[string]string{
		"src/main.rs": "fn main() {}\n",
	})
	t.Chdir(t.TempDir())

	previousCmd := rootCmd
	previousFlags := rootFlags
	rootCmd = newRootCmd()
	rootFlags = app.Flags{Directory: targetRoot}
	t.Cleanup(func() {
		rootCmd = previousCmd
		rootFlags = previousFlags
	})

	var runErr error
	stdout, stderr := captureProcessOutput(t, func() {
		command := newOverviewCmd()
		runErr = command.RunE(command, []string{"."})
	})
	if runErr != nil {
		t.Fatalf("run overview command: %v", runErr)
	}
	if !strings.Contains(stdout, "src/,") {
		t.Fatalf("expected explicit -C overview, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func TestSymbolsCommandResolvesRelativePathAgainstExplicitDirectory(t *testing.T) {
	ensureCommandLanguages(t, "rust")
	targetRoot := commandProject(t, map[string]string{
		"src/main.rs":    "fn main() {}\n",
		"other/extra.rs": "fn extra() {}\n",
	})
	t.Chdir(t.TempDir())

	previousCmd := rootCmd
	previousFlags := rootFlags
	rootCmd = newRootCmd()
	rootFlags = app.Flags{Directory: targetRoot}
	t.Cleanup(func() {
		rootCmd = previousCmd
		rootFlags = previousFlags
	})

	var runErr error
	stdout, stderr := captureProcessOutput(t, func() {
		command := newSymbolsCmd()
		runErr = command.RunE(command, []string{"src"})
	})
	if runErr != nil {
		t.Fatalf("run symbols command: %v", runErr)
	}
	if !strings.Contains(stdout, "src/main.rs,1,main,fn") {
		t.Fatalf("expected explicit -C scoped symbols, got %q", stdout)
	}
	if strings.Contains(stdout, "other/extra.rs") {
		t.Fatalf("expected explicit -C relative path to filter results, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func TestDefinitionCommandResolvesRelativePathAgainstExplicitDirectory(t *testing.T) {
	ensureCommandLanguages(t, "rust")
	targetRoot := commandProject(t, map[string]string{
		"src/main.rs":    "fn main() {}\n",
		"other/extra.rs": "fn extra() {}\n",
	})
	t.Chdir(t.TempDir())

	previousCmd := rootCmd
	previousFlags := rootFlags
	rootCmd = newRootCmd()
	rootFlags = app.Flags{Directory: targetRoot}
	t.Cleanup(func() {
		rootCmd = previousCmd
		rootFlags = previousFlags
	})

	var runErr error
	stdout, stderr := captureProcessOutput(t, func() {
		command := newDefinitionCmd()
		if err := command.Flags().Set("name", "main"); err != nil {
			t.Fatalf("set name flag: %v", err)
		}
		runErr = command.RunE(command, []string{"src"})
	})
	if runErr != nil {
		t.Fatalf("run definition command: %v", runErr)
	}
	if !strings.Contains(stdout, "file: src/main.rs") {
		t.Fatalf("expected explicit -C scoped definition, got %q", stdout)
	}
	if strings.Contains(stdout, "other/extra.rs") {
		t.Fatalf("expected explicit -C relative path to filter definitions, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func TestReferencesCommandResolvesRelativePathAgainstExplicitDirectory(t *testing.T) {
	ensureCommandLanguages(t, "rust")
	targetRoot := commandProject(t, map[string]string{
		"src/main.rs":    "fn helper() {}\nfn main() { helper(); }\n",
		"other/extra.rs": "fn helper() {}\nfn extra() { helper(); }\n",
	})
	t.Chdir(t.TempDir())

	previousCmd := rootCmd
	previousFlags := rootFlags
	rootCmd = newRootCmd()
	rootFlags = app.Flags{Directory: targetRoot}
	t.Cleanup(func() {
		rootCmd = previousCmd
		rootFlags = previousFlags
	})

	var runErr error
	stdout, stderr := captureProcessOutput(t, func() {
		command := newReferencesCmd()
		if err := command.Flags().Set("name", "helper"); err != nil {
			t.Fatalf("set name flag: %v", err)
		}
		runErr = command.RunE(command, []string{"src"})
	})
	if runErr != nil {
		t.Fatalf("run references command: %v", runErr)
	}
	if !strings.Contains(stdout, "src/main.rs") {
		t.Fatalf("expected explicit -C scoped references, got %q", stdout)
	}
	if strings.Contains(stdout, "other/extra.rs") {
		t.Fatalf("expected explicit -C relative path to filter references, got %q", stdout)
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
