package index

import (
	"reflect"
	"testing"
)

func TestPublicSymbolKindsMatchesSupportedKinds(t *testing.T) {
	expected := []SymbolKind{
		SymbolKindFunc,
		SymbolKindConst,
		SymbolKindStruct,
		SymbolKindEnum,
		SymbolKindClass,
		SymbolKindInterface,
		SymbolKindModule,
		SymbolKindType,
	}

	if !reflect.DeepEqual(PublicSymbolKinds(), expected) {
		t.Fatalf("unexpected public symbol kinds: %v", PublicSymbolKinds())
	}
}

func TestParseSymbolKindSupportsRefinedKinds(t *testing.T) {
	for _, kind := range PublicSymbolKinds() {
		raw := string(kind)
		if _, err := ParseSymbolKind(raw); err != nil {
			t.Fatalf("expected kind %q to parse: %v", raw, err)
		}
	}
}

func TestParseSymbolKindRejectsRemovedKinds(t *testing.T) {
	for _, raw := range []string{"fn", "method", "trait", "event"} {
		if _, err := ParseSymbolKind(raw); err == nil {
			t.Fatalf("expected removed kind %q to be rejected", raw)
		}
	}
}
