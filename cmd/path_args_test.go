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
	if !strings.Contains(stdout, "src/main.rs,1,main,func") {
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
	if !strings.Contains(stdout, "src/main.rs,1,main,func") {
		t.Fatalf("expected src/main.rs symbol, got %q", stdout)
	}
	if !strings.Contains(stdout, "pkg/helper.rs,1,helper,func") {
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

func TestOverviewCommandIndexesDirectorySymlink(t *testing.T) {
	ensureCommandLanguages(t, "rust")
	targetRoot := commandProject(t, map[string]string{
		"shared/src/main.rs": "fn main() {}\n",
	})
	if err := os.Symlink(filepath.Join(targetRoot, "shared", "src"), filepath.Join(targetRoot, "linked-src")); err != nil {
		t.Fatalf("create directory symlink: %v", err)
	}
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
		runErr = command.RunE(command, []string{"linked-src"})
	})
	if runErr != nil {
		t.Fatalf("run overview command: %v", runErr)
	}
	if !strings.Contains(stdout, "linked-src/main.rs") {
		t.Fatalf("expected symlinked directory overview, got %q", stdout)
	}
	if strings.Contains(stdout, "shared/src/main.rs") {
		t.Fatalf("expected output to preserve symlink path, got %q", stdout)
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

func TestSymbolsCommandIndexesDirectorySymlink(t *testing.T) {
	ensureCommandLanguages(t, "rust")
	targetRoot := commandProject(t, map[string]string{
		"shared/src/main.rs": "fn main() {}\n",
	})
	if err := os.Symlink(filepath.Join(targetRoot, "shared", "src"), filepath.Join(targetRoot, "linked-src")); err != nil {
		t.Fatalf("create directory symlink: %v", err)
	}
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
		runErr = command.RunE(command, []string{"linked-src"})
	})
	if runErr != nil {
		t.Fatalf("run symbols command: %v", runErr)
	}
	if !strings.Contains(stdout, "linked-src/main.rs,1,main,func") {
		t.Fatalf("expected symlinked directory symbols, got %q", stdout)
	}
	if strings.Contains(stdout, "shared/src/main.rs") {
		t.Fatalf("expected output to preserve symlink path, got %q", stdout)
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
	if !strings.Contains(stdout, "src/main.rs,1,main,func") {
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

func TestCalleesCommandResolvesRelativePathAgainstExplicitDirectory(t *testing.T) {
	ensureCommandLanguages(t, "go")
	targetRoot := commandProject(t, map[string]string{
		"src/main.go":    "package main\n\nfunc helper() {}\nfunc A() { helper() }\n",
		"other/extra.go": "package main\n\nfunc helper() {}\nfunc A() { helper() }\n",
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
		command := newCalleesCmd()
		if err := command.Flags().Set("name", "A"); err != nil {
			t.Fatalf("set name flag: %v", err)
		}
		runErr = command.RunE(command, []string{"src"})
	})
	if runErr != nil {
		t.Fatalf("run callees command: %v", runErr)
	}
	if !strings.Contains(stdout, "src/main.go") {
		t.Fatalf("expected explicit -C scoped callees, got %q", stdout)
	}
	if strings.Contains(stdout, "other/extra.go") {
		t.Fatalf("expected explicit -C relative path to filter callees, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func TestCommandsRemoveScopeFlag(t *testing.T) {
	for _, command := range []*cobra.Command{
		newSymbolsCmd(),
		newDefinitionCmd(),
		newCalleesCmd(),
		newReferencesCmd(),
	} {
		if command.Flags().Lookup("scope") != nil {
			t.Fatalf("%s should not expose --scope", command.Name())
		}
	}
}

func TestSymbolsCommandSupportsGlobPathArgs(t *testing.T) {
	ensureCommandLanguages(t, "rust")
	targetRoot := commandProject(t, map[string]string{
		"src/main.rs":    "fn main() {}\n",
		"pkg/helper.rs":  "fn helper() {}\n",
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
		runErr = command.RunE(command, []string{"{src,pkg}/*.rs"})
	})
	if runErr != nil {
		t.Fatalf("run symbols command with glob path: %v", runErr)
	}
	if !strings.Contains(stdout, "src/main.rs,1,main,func") {
		t.Fatalf("expected src/main.rs symbol, got %q", stdout)
	}
	if !strings.Contains(stdout, "pkg/helper.rs,1,helper,func") {
		t.Fatalf("expected pkg/helper.rs symbol, got %q", stdout)
	}
	if strings.Contains(stdout, "other/extra.rs") {
		t.Fatalf("expected glob path args to filter results, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func TestSymbolsCommandReportsMissingGlobPathMatches(t *testing.T) {
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
	_, _ = captureProcessOutput(t, func() {
		command := newSymbolsCmd()
		runErr = command.RunE(command, []string{"missing/**/*.rs"})
	})
	if runErr == nil {
		t.Fatal("expected missing glob path error")
	}
	if !strings.Contains(runErr.Error(), "gx: no indexed files match missing/**/*.rs") {
		t.Fatalf("unexpected error: %v", runErr)
	}
}

func TestDefinitionCommandSupportsIncludeExcludePathGlobs(t *testing.T) {
	ensureCommandLanguages(t, "go")
	targetRoot := commandProject(t, map[string]string{
		"src/main.go":        "package main\n\nfunc helper() {}\n",
		"src/helper_test.go": "package main\n\nfunc helper() {}\n",
		"src/mocks/mock.go":  "package main\n\nfunc helper() {}\n",
		"other/extra.go":     "package main\n\nfunc helper() {}\n",
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
		if err := command.Flags().Set("name", "helper"); err != nil {
			t.Fatalf("set name flag: %v", err)
		}
		if err := command.Flags().Set("include", "src/**"); err != nil {
			t.Fatalf("set include flag: %v", err)
		}
		if err := command.Flags().Set("exclude", "{**/*_test.go,**/mocks/**}"); err != nil {
			t.Fatalf("set exclude flag: %v", err)
		}
		runErr = command.RunE(command, []string{"."})
	})
	if runErr != nil {
		t.Fatalf("run definition command with include/exclude globs: %v", runErr)
	}
	if !strings.Contains(stdout, "file: src/main.go") {
		t.Fatalf("expected src/main.go definition, got %q", stdout)
	}
	if strings.Contains(stdout, "helper_test.go") || strings.Contains(stdout, "mocks/mock.go") || strings.Contains(stdout, "other/extra.go") {
		t.Fatalf("expected include/exclude globs to filter definitions, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func TestReferencesCommandSupportsIncludeExcludePathGlobs(t *testing.T) {
	ensureCommandLanguages(t, "rust")
	targetRoot := commandProject(t, map[string]string{
		"src/main.rs":        "fn helper() {}\nfn main() { helper(); }\n",
		"src/helper_test.rs": "fn helper() {}\nfn test_call() { helper(); }\n",
		"src/mocks/mock.rs":  "fn helper() {}\nfn mock_call() { helper(); }\n",
		"other/extra.rs":     "fn helper() {}\nfn extra() { helper(); }\n",
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
		if err := command.Flags().Set("include", "src/**"); err != nil {
			t.Fatalf("set include flag: %v", err)
		}
		if err := command.Flags().Set("exclude", "{**/*_test.rs,**/mocks/**}"); err != nil {
			t.Fatalf("set exclude flag: %v", err)
		}
		runErr = command.RunE(command, []string{"."})
	})
	if runErr != nil {
		t.Fatalf("run references command with include/exclude globs: %v", runErr)
	}
	if !strings.Contains(stdout, "src/main.rs") {
		t.Fatalf("expected src/main.rs references, got %q", stdout)
	}
	if strings.Contains(stdout, "helper_test.rs") || strings.Contains(stdout, "mocks/mock.rs") || strings.Contains(stdout, "other/extra.rs") {
		t.Fatalf("expected include/exclude globs to filter references, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func TestCalleesCommandSupportsExcludePathGlobs(t *testing.T) {
	ensureCommandLanguages(t, "go")
	targetRoot := commandProject(t, map[string]string{
		"src/main.go":       "package main\n\nfunc helper() {}\nfunc A() { helper() }\n",
		"src/mocks/mock.go": "package main\n\nfunc helper() {}\nfunc A() { helper() }\n",
		"other/extra.go":    "package main\n\nfunc helper() {}\nfunc A() { helper() }\n",
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
		command := newCalleesCmd()
		if err := command.Flags().Set("name", "A"); err != nil {
			t.Fatalf("set name flag: %v", err)
		}
		if err := command.Flags().Set("exclude", "{**/mocks/**,other/**}"); err != nil {
			t.Fatalf("set exclude flag: %v", err)
		}
		runErr = command.RunE(command, []string{"."})
	})
	if runErr != nil {
		t.Fatalf("run callees command with exclude globs: %v", runErr)
	}
	if !strings.Contains(stdout, "src/main.go") {
		t.Fatalf("expected src/main.go callee rows, got %q", stdout)
	}
	if strings.Contains(stdout, "mocks/mock.go") || strings.Contains(stdout, "other/extra.go") {
		t.Fatalf("expected exclude globs to filter callees, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func TestSymbolsCommandIncludeOverridesIgnoreAndIgnoreRemovalRestoresVisibility(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ensureCommandLanguages(t, "go")
	t.Chdir(t.TempDir())

	targetRoot := commandProject(t, map[string]string{
		".gitignore":     "pkg/ignored.go\n",
		"pkg/ignored.go": "package pkg\n\nfunc Hidden() {}\n",
	})

	previousCmd := rootCmd
	previousFlags := rootFlags
	rootCmd = newRootCmd()
	rootFlags = app.Flags{Directory: targetRoot, JSON: true}
	t.Cleanup(func() {
		rootCmd = previousCmd
		rootFlags = previousFlags
	})

	var runErr error
	baseStdout, baseStderr := captureProcessOutput(t, func() {
		command := newSymbolsCmd()
		if err := command.Flags().Set("name", "Hidden"); err != nil {
			t.Fatalf("set name flag: %v", err)
		}
		runErr = command.RunE(command, []string{"."})
	})
	if runErr != nil {
		t.Fatalf("base query: %v", runErr)
	}
	if baseStdout != "[]\n" {
		t.Fatalf("expected ignored base query to emit empty json array, got %q", baseStdout)
	}
	if strings.TrimSpace(baseStderr) != "" {
		t.Fatalf("expected ignored base query stderr to be empty, got %q", baseStderr)
	}

	runErr = nil
	includedStdout, includedStderr := captureProcessOutput(t, func() {
		command := newSymbolsCmd()
		if err := command.Flags().Set("name", "Hidden"); err != nil {
			t.Fatalf("set name flag: %v", err)
		}
		if err := command.Flags().Set("include", "pkg/ignored.go"); err != nil {
			t.Fatalf("set include flag: %v", err)
		}
		runErr = command.RunE(command, []string{"."})
	})
	if runErr != nil {
		t.Fatalf("include query: %v", runErr)
	}
	if !strings.Contains(includedStdout, "\"name\": \"Hidden\"") {
		t.Fatalf("expected include query to restore ignored symbol, got %q", includedStdout)
	}
	if strings.TrimSpace(includedStderr) != "" {
		t.Fatalf("expected include query stderr to be empty, got %q", includedStderr)
	}

	if err := os.WriteFile(filepath.Join(targetRoot, ".gitignore"), []byte{}, 0o644); err != nil {
		t.Fatalf("clear .gitignore: %v", err)
	}

	runErr = nil
	visibleStdout, visibleStderr := captureProcessOutput(t, func() {
		command := newSymbolsCmd()
		if err := command.Flags().Set("name", "Hidden"); err != nil {
			t.Fatalf("set name flag: %v", err)
		}
		runErr = command.RunE(command, []string{"."})
	})
	if runErr != nil {
		t.Fatalf("visible query: %v", runErr)
	}
	if !strings.Contains(visibleStdout, "\"name\": \"Hidden\"") {
		t.Fatalf("expected symbol to become visible after clearing ignore, got %q", visibleStdout)
	}
	if strings.TrimSpace(visibleStderr) != "" {
		t.Fatalf("expected visible query stderr to be empty, got %q", visibleStderr)
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
