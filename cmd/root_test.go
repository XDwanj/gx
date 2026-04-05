package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/XDwanj/gx/internal/app"
)

func TestRootCommandUsesGx(t *testing.T) {
	if rootCmd.Use != "gx" {
		t.Fatalf("expected root command name gx, got %q", rootCmd.Use)
	}
}

func TestRootCommandExposesVerboseFlag(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("verbose")
	if flag == nil {
		t.Fatalf("expected verbose flag to be registered")
	}
	if flag.DefValue != "false" {
		t.Fatalf("expected verbose default false, got %q", flag.DefValue)
	}
}

func TestRootCommandExposesPaginationFlags(t *testing.T) {
	for _, name := range []string{"limit", "offset", "all"} {
		if rootCmd.PersistentFlags().Lookup(name) == nil {
			t.Fatalf("expected %s flag to be registered", name)
		}
	}
}

func TestRootCommandPrintsVersionForLongFlag(t *testing.T) {
	stdout, stderr, exitCode := executeRootForTest(t, "--version")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", exitCode, stderr)
	}
	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	if stdout != "gx dev\n" {
		t.Fatalf("expected version output, got %q", stdout)
	}
}

func TestRootCommandPrintsVersionForShortAliases(t *testing.T) {
	for _, args := range [][]string{{"-v"}, {"-V"}} {
		stdout, stderr, exitCode := executeRootForTest(t, args...)
		if exitCode != 0 {
			t.Fatalf("args %v: expected exit code 0, got %d with stderr %q", args, exitCode, stderr)
		}
		if stderr != "" {
			t.Fatalf("args %v: expected empty stderr, got %q", args, stderr)
		}
		if stdout != "gx dev\n" {
			t.Fatalf("args %v: expected version output, got %q", args, stdout)
		}
	}
}

func TestRootCommandPrintsVersionAsJSON(t *testing.T) {
	stdout, stderr, exitCode := executeRootForTest(t, "--json", "--version")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", exitCode, stderr)
	}
	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}

	expected := "{\n  \"version\": \"dev\"\n}\n"
	if stdout != expected {
		t.Fatalf("expected JSON version output, got %q", stdout)
	}
}

func executeRootForTest(t *testing.T, args ...string) (string, string, int) {
	t.Helper()

	previousCmd := rootCmd
	previousFlags := rootFlags
	previousShowVersion := showVersion
	rootCmd = newRootCmd()
	rootFlags = app.Flags{}
	showVersion = false

	t.Cleanup(func() {
		rootCmd = previousCmd
		rootFlags = previousFlags
		showVersion = previousShowVersion
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs(args)

	exitCode := Execute()
	return stdout.String(), strings.TrimSpace(stderr.String()), exitCode
}
