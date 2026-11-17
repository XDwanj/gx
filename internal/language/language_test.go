package language

import (
	"testing"
)

const goSource = `package demo

type Server struct {
	addr string
}

func (s *Server) Start() error {
	return nil
}

func NewServer() *Server {
	return &Server{}
}

const defaultPort = 8080
`

func TestParseAndExtractGoKinds(t *testing.T) {
	symbols, err := ParseAndExtract("go", []byte(goSource), "server.go")
	if err != nil {
		t.Fatalf("ParseAndExtract: %v", err)
	}

	kinds := map[string]SymbolKind{}
	for _, symbol := range symbols {
		kinds[symbol.Name] = symbol.Kind
	}

	expected := map[string]SymbolKind{
		"Server":      SymbolKindStruct,
		"Start":       SymbolKindFunc,
		"NewServer":   SymbolKindFunc,
		"defaultPort": SymbolKindConst,
	}
	for name, kind := range expected {
		if kinds[name] != kind {
			t.Fatalf("%s kind = %q, want %q (all symbols: %+v)", name, kinds[name], kind, symbols)
		}
	}
}

func TestParseAndExtractPythonKinds(t *testing.T) {
	source := "class Greeter:\n    def hello(self):\n        pass\n\ndef top_level():\n    pass\n"
	symbols, err := ParseAndExtract("python", []byte(source), "greet.py")
	if err != nil {
		t.Fatalf("ParseAndExtract: %v", err)
	}

	kinds := map[string]SymbolKind{}
	for _, symbol := range symbols {
		kinds[symbol.Name] = symbol.Kind
	}
	if kinds["Greeter"] != SymbolKindClass || kinds["top_level"] != SymbolKindFunc {
		t.Fatalf("python kinds %+v", symbols)
	}
}

func TestFindReferencesGo(t *testing.T) {
	source := "package demo\n\nfunc helper() {}\n\nfunc caller() {\n\thelper()\n\thelper()\n}\n"
	references, err := FindReferencesForNames("go", []byte(source), "demo.go", []string{"helper"})
	if err != nil {
		t.Fatalf("FindReferencesForNames: %v", err)
	}
	if len(references) < 3 {
		t.Fatalf("expected at least 3 references (decl + 2 calls), got %+v", references)
	}

	lines := map[int]bool{}
	for _, reference := range references {
		lines[reference.Line] = true
	}
	if !lines[6] || !lines[7] {
		t.Fatalf("call sites on lines 6 and 7 missing: %+v", references)
	}
}
