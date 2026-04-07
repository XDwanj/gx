package cmd

import (
	"strings"
	"testing"
)

func TestLangHelpUsesEnableDisableCommands(t *testing.T) {
	stdout, stderr, exitCode := executeRootForTest(t, "lang", "--help")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", exitCode, stderr)
	}
	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	if !strings.Contains(stdout, "enable      Enable language grammars") {
		t.Fatalf("expected enable command in help output, got %q", stdout)
	}
	if !strings.Contains(stdout, "disable     Disable language grammars") {
		t.Fatalf("expected disable command in help output, got %q", stdout)
	}
	if strings.Contains(stdout, " add ") || strings.Contains(stdout, "\n  add") {
		t.Fatalf("unexpected legacy add command in help output: %q", stdout)
	}
	if strings.Contains(stdout, " remove ") || strings.Contains(stdout, "\n  remove") {
		t.Fatalf("unexpected legacy remove command in help output: %q", stdout)
	}
}

func TestLangEnableRequiresAtLeastOneLanguage(t *testing.T) {
	_, stderr, exitCode := executeRootForTest(t, "lang", "enable")
	if exitCode == 0 {
		t.Fatalf("expected non-zero exit code, got %d", exitCode)
	}
	expected := "gx: specify at least one language, e.g.: gx lang enable rust typescript"
	if !strings.Contains(stderr, expected) {
		t.Fatalf("expected error %q, got %q", expected, stderr)
	}
}

func TestLangDisableRequiresAtLeastOneLanguage(t *testing.T) {
	_, stderr, exitCode := executeRootForTest(t, "lang", "disable")
	if exitCode == 0 {
		t.Fatalf("expected non-zero exit code, got %d", exitCode)
	}
	expected := "gx: specify at least one language, e.g.: gx lang disable rust"
	if !strings.Contains(stderr, expected) {
		t.Fatalf("expected error %q, got %q", expected, stderr)
	}
}

func TestLangLegacyAddRemoveCommandsAreUnavailable(t *testing.T) {
	command := newLangCmd()
	subcommands := map[string]bool{}
	for _, subcommand := range command.Commands() {
		subcommands[subcommand.Name()] = true
	}
	for _, legacy := range []string{"add", "remove"} {
		if subcommands[legacy] {
			t.Fatalf("unexpected legacy command %q", legacy)
		}
	}
	for _, required := range []string{"enable", "disable", "list"} {
		if !subcommands[required] {
			t.Fatalf("missing required command %q", required)
		}
	}
}
