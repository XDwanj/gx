package cmd

import (
	"bytes"
	"gx/internal/app"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
	rootFlags = app.Flags{Root: root}
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
