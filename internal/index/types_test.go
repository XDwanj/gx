package index

import "testing"

func TestParseSymbolKindRoundTrip(t *testing.T) {
	for _, kind := range PublicSymbolKinds() {
		parsed, err := ParseSymbolKind(string(kind))
		if err != nil {
			t.Fatalf("public kind %q should parse: %v", kind, err)
		}
		if parsed != kind {
			t.Fatalf("round trip %q became %q", kind, parsed)
		}
	}
}

func TestParseSymbolKindRejectsUnknown(t *testing.T) {
	if _, err := ParseSymbolKind("nosuchkind"); err == nil {
		t.Fatal("unknown kind should error")
	}
}
