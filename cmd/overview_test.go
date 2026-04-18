package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/XDwanj/gx/internal/app"
)

type overviewSectionJSON struct {
	Target     string           `json:"target"`
	TargetKind string           `json:"target_kind"`
	Rows       []map[string]any `json:"rows"`
}

func TestOverviewCommandSupportsMarkdownFile(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	t.Chdir(root)
	if err := os.WriteFile(
		filepath.Join(root, "README.md"),
		[]byte("# Title\n\n## Section\n"),
		0o644,
	); err != nil {
		t.Fatalf("write README.md: %v", err)
	}

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
		command := newOverviewCmd()
		runErr = command.RunE(command, []string{"README.md"})
	})
	if runErr != nil {
		t.Fatalf("run overview command: %v", runErr)
	}
	if !strings.Contains(stdout, "1,Title") || !strings.Contains(stdout, "2,Section") {
		t.Fatalf("unexpected stdout: %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func TestOverviewCommandSupportsMultiplePathArgs(t *testing.T) {
	ensureCommandLanguages(t, "rust")
	root := commandProject(t, map[string]string{
		"src/main.rs": "fn main() {}\n",
		"README.md":   "# Title\n\n## Section\n",
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
		command := newOverviewCmd()
		runErr = command.RunE(command, []string{"src/main.rs", "README.md"})
	})
	if runErr != nil {
		t.Fatalf("run overview command: %v", runErr)
	}
	for _, expected := range []string{
		"target: src/main.rs",
		"target_kind: file",
		"\"src/main.rs:1\",main,func",
		"target: README.md",
		"target_kind: markdown",
		"1,Title",
		"2,Section",
	} {
		if !strings.Contains(stdout, expected) {
			t.Fatalf("expected %q in stdout, got %q", expected, stdout)
		}
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func TestOverviewCommandMultiplePathsEmitSectionedJSON(t *testing.T) {
	ensureCommandLanguages(t, "rust")
	root := commandProject(t, map[string]string{
		"src/main.rs": "fn main() {}\n",
		"README.md":   "# Title\n\n## Section\n",
	})

	stdout, stderr, exitCode := executeRootFixtureCommand(t, root, "overview", "src/main.rs", "README.md")
	if exitCode != 0 {
		t.Fatalf("expected overview exit code 0, got %d with stderr %q", exitCode, stderr)
	}

	var sections []overviewSectionJSON
	if err := json.Unmarshal([]byte(stdout), &sections); err != nil {
		t.Fatalf("unmarshal stdout: %v\nstdout=%s", err, stdout)
	}
	if len(sections) != 2 {
		t.Fatalf("expected two overview sections, got %d", len(sections))
	}
	if sections[0].Target != "src/main.rs" || sections[0].TargetKind != "file" {
		t.Fatalf("unexpected first section: %+v", sections[0])
	}
	if sections[1].Target != "README.md" || sections[1].TargetKind != "markdown" {
		t.Fatalf("unexpected second section: %+v", sections[1])
	}
	if len(sections[0].Rows) != 1 || sections[0].Rows[0]["name"] != "main" {
		t.Fatalf("unexpected file rows: %+v", sections[0].Rows)
	}
	if len(sections[1].Rows) != 2 || sections[1].Rows[0]["heading"] != "Title" {
		t.Fatalf("unexpected markdown rows: %+v", sections[1].Rows)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func captureProcessOutput(t *testing.T, run func()) (string, string) {
	t.Helper()

	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}

	previousStdout := os.Stdout
	previousStderr := os.Stderr
	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter
	t.Cleanup(func() {
		os.Stdout = previousStdout
		os.Stderr = previousStderr
	})

	run()

	if err := stdoutWriter.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	if err := stderrWriter.Close(); err != nil {
		t.Fatalf("close stderr writer: %v", err)
	}

	var stdout bytes.Buffer
	if _, err := stdout.ReadFrom(stdoutReader); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	var stderr bytes.Buffer
	if _, err := stderr.ReadFrom(stderrReader); err != nil {
		t.Fatalf("read stderr: %v", err)
	}
	return stdout.String(), stderr.String()
}
