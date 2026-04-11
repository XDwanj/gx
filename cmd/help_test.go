package cmd

import (
	"strings"
	"testing"
)

func TestSymbolsHelpIncludesKindSupportList(t *testing.T) {
	stdout, stderr, exitCode := executeRootForTest(t, "symbols", "--help")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", exitCode, stderr)
	}
	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	if !strings.Contains(stdout, "Public kinds:") {
		t.Fatalf("expected public kinds section, got %q", stdout)
	}
	if !strings.Contains(stdout, "- go: func, const, struct, interface, type") {
		t.Fatalf("expected go support list, got %q", stdout)
	}
	if strings.Contains(stdout, "| go |") {
		t.Fatalf("expected list format, got table-like output %q", stdout)
	}
}

func TestDefinitionHelpIncludesKindSupportList(t *testing.T) {
	stdout, stderr, exitCode := executeRootForTest(t, "definition", "--help")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", exitCode, stderr)
	}
	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	if !strings.Contains(stdout, "Language support summary:") {
		t.Fatalf("expected language support section, got %q", stdout)
	}
	if !strings.Contains(stdout, "- rust: func, const, struct, enum, interface, module, type") {
		t.Fatalf("expected rust support list, got %q", stdout)
	}
}

func TestCalleesHelpUsesCallerNameFlagDescription(t *testing.T) {
	stdout, stderr, exitCode := executeRootForTest(t, "callees", "--help")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", exitCode, stderr)
	}
	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	if !strings.Contains(stdout, "Glob pattern to match caller symbol names") {
		t.Fatalf("expected caller name flag description, got %q", stdout)
	}
	if strings.Contains(stdout, "--kind") {
		t.Fatalf("did not expect kind flag in callees help, got %q", stdout)
	}
}

func TestQueryHelpIncludesPathFilterFlags(t *testing.T) {
	for _, commandName := range []string{"symbols", "definition", "references", "callees"} {
		stdout, stderr, exitCode := executeRootForTest(t, commandName, "--help")
		if exitCode != 0 {
			t.Fatalf("%s help exit code = %d, stderr=%q", commandName, exitCode, stderr)
		}
		if stderr != "" {
			t.Fatalf("%s help stderr = %q", commandName, stderr)
		}
		if !strings.Contains(stdout, "--include stringArray") {
			t.Fatalf("expected --include flag in %s help, got %q", commandName, stdout)
		}
		if !strings.Contains(stdout, "--exclude stringArray") {
			t.Fatalf("expected --exclude flag in %s help, got %q", commandName, stdout)
		}
	}
}
