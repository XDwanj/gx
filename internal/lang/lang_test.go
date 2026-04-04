package lang

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestGrammarCacheDirUsesGxNamespace(t *testing.T) {
	cacheDir := GrammarCacheDir()
	if !strings.Contains(cacheDir, filepath.Join("gx", "grammars")) {
		t.Fatalf("expected gx grammar cache namespace, got %q", cacheDir)
	}
	if strings.Contains(cacheDir, filepath.Join("cx", "grammars")) {
		t.Fatalf("unexpected cx grammar cache namespace in %q", cacheDir)
	}
}

func TestSupportedLanguagesExcludeRemovedGrammars(t *testing.T) {
	for _, language := range SupportedLanguages() {
		if language == "elixir" || language == "solidity" {
			t.Fatalf("unexpected removed language in supported list: %s", language)
		}
	}
}
