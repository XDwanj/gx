package lang

import (
	"bytes"
	"io"
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
	foundKotlin := false
	foundProtobuf := false
	foundVue := false
	for _, language := range SupportedLanguages() {
		if language == "elixir" || language == "solidity" {
			t.Fatalf("unexpected removed language in supported list: %s", language)
		}
		if language == "kotlin" {
			foundKotlin = true
		}
		if language == "protobuf" {
			foundProtobuf = true
		}
		if language == "vue" {
			foundVue = true
		}
	}
	if !foundKotlin {
		t.Fatal("expected kotlin in supported language list")
	}
	if !foundProtobuf {
		t.Fatal("expected protobuf in supported language list")
	}
	if !foundVue {
		t.Fatal("expected vue in supported language list")
	}
}

func TestListUsesEnableDisableMarkers(t *testing.T) {
	isolateCacheEnv(t)
	if err := Add(io.Discard, io.Discard, []string{"go"}); err != nil {
		t.Fatalf("enable go grammar: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := List(&stdout, &stderr); err != nil {
		t.Fatalf("list grammars: %v", err)
	}

	output := stdout.String()
	if strings.Contains(output, "[installed]") || strings.Contains(output, "[missing]") {
		t.Fatalf("unexpected legacy markers in output: %q", output)
	}
	if marker := markerForLanguage(output, "go"); marker != "[enabled]" {
		t.Fatalf("unexpected marker for go: %q", marker)
	}
	if marker := markerForLanguage(output, "rust"); marker != "[disabled]" {
		t.Fatalf("unexpected marker for rust: %q", marker)
	}
}

func isolateCacheEnv(t *testing.T) {
	t.Helper()
	cacheRoot := t.TempDir()
	t.Setenv("HOME", cacheRoot)
	t.Setenv("XDG_CACHE_HOME", cacheRoot)
	t.Setenv("LOCALAPPDATA", cacheRoot)
}

func markerForLanguage(output string, language string) string {
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		if fields[0] == language {
			return fields[1]
		}
	}
	return ""
}
