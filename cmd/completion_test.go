package cmd

import (
	"reflect"
	"testing"

	"github.com/XDwanj/gx/internal/index"

	"github.com/spf13/cobra"
)

func TestSymbolsKindFlagCompletionIncludesPublicKinds(t *testing.T) {
	assertKindFlagCompletion(t, newSymbolsCmd())
}

func TestDefinitionKindFlagCompletionIncludesPublicKinds(t *testing.T) {
	assertKindFlagCompletion(t, newDefinitionCmd())
}

func assertKindFlagCompletion(t *testing.T, command *cobra.Command) {
	t.Helper()

	completionFunc, ok := command.GetFlagCompletionFunc("kind")
	if !ok {
		t.Fatalf("expected kind flag completion on %s", command.Name())
	}

	completions, directive := completionFunc(command, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("expected no-file completion directive, got %v", directive)
	}

	expected := make([]string, 0, len(index.PublicSymbolKinds()))
	for _, kind := range index.PublicSymbolKinds() {
		expected = append(expected, string(kind))
	}
	if !reflect.DeepEqual(completions, expected) {
		t.Fatalf("unexpected completions for %s: %v", command.Name(), completions)
	}
}
