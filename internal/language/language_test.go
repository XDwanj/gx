package language

import (
	"gx/internal/lang"
	"io"
	"strings"
	"testing"
)

func ensureInstalled(t *testing.T, languages ...string) {
	t.Helper()
	if err := lang.Add(io.Discard, io.Discard, languages); err != nil {
		t.Fatalf("install grammars: %v", err)
	}
}

func TestParseAndExtractRust(t *testing.T) {
	ensureInstalled(t, "rust")

	symbols, err := ParseAndExtract("rust", []byte("pub fn calculate_fee(amount: u64) -> u64 { amount }\n"), "test.rs")
	if err != nil {
		t.Fatalf("parse rust: %v", err)
	}
	if len(symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(symbols))
	}
	if symbols[0].Name != "calculate_fee" {
		t.Fatalf("unexpected symbol name: %s", symbols[0].Name)
	}
	if symbols[0].Kind != SymbolKindFn {
		t.Fatalf("unexpected kind: %s", symbols[0].Kind)
	}
}

func TestParseAndExtractTSX(t *testing.T) {
	ensureInstalled(t, "typescript")

	symbols, err := ParseAndExtract("typescript", []byte("export function App() { return <div />; }"), "test.tsx")
	if err != nil {
		t.Fatalf("parse tsx: %v", err)
	}
	if len(symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(symbols))
	}
	if symbols[0].Name != "App" {
		t.Fatalf("unexpected symbol name: %s", symbols[0].Name)
	}
}

func TestFindReferencesRust(t *testing.T) {
	ensureInstalled(t, "rust")

	source := []byte("struct Symbol;\nfn use_it(symbol: Symbol) -> Symbol { let copy = Symbol; copy }\n")
	references, err := FindReferences("rust", source, "test.rs", "Symbol")
	if err != nil {
		t.Fatalf("find references: %v", err)
	}
	if len(references) < 3 {
		t.Fatalf("expected at least 3 references, got %d", len(references))
	}
}

func TestNotInstalledMessageUsesGxCommand(t *testing.T) {
	err := newNotInstalled("rust")
	if !strings.Contains(err.Error(), "gx lang add rust") {
		t.Fatalf("expected gx install hint, got %q", err.Error())
	}
	if strings.Contains(err.Error(), "cx lang add rust") {
		t.Fatalf("unexpected cx install hint, got %q", err.Error())
	}
}
