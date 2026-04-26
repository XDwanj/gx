package language

import (
	"io"
	"strings"
	"testing"

	"github.com/XDwanj/gx/internal/lang"
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
	if symbols[0].Kind != SymbolKindFunc {
		t.Fatalf("unexpected kind: %s", symbols[0].Kind)
	}
	if symbols[0].Line != 1 {
		t.Fatalf("unexpected coordinates: %+v", symbols[0])
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

func TestParseAndExtractProtobufKinds(t *testing.T) {
	ensureInstalled(t, "protobuf")

	source := []byte(`syntax = "proto3";

message HelloRequest {
  string name = 1;
}

enum Status {
  STATUS_UNSPECIFIED = 0;
}

service Greeter {
  rpc SayHello (HelloRequest) returns (HelloReply) {
    option deprecated = true;
  }
}

message HelloReply {
  string message = 1;
}
`)

	symbols, err := ParseAndExtract("protobuf", source, "greeter.proto")
	if err != nil {
		t.Fatalf("parse protobuf kinds: %v", err)
	}

	assertKinds(t, symbols, map[string]SymbolKind{
		"HelloRequest": SymbolKindStruct,
		"Status":       SymbolKindEnum,
		"Greeter":      SymbolKindInterface,
		"SayHello":     SymbolKindFunc,
		"HelloReply":   SymbolKindStruct,
	})
}

func TestParseAndExtractGoKinds(t *testing.T) {
	ensureInstalled(t, "go")

	source := []byte(`package demo

type User struct{}
type Store interface{ Save() }
type Status int
type Alias = int

const (
	StatusOpen Status = iota
	StatusClosed
)

var DefaultStatus = StatusOpen

func Build() {}

func (status Status) Label() string { return "" }
`)

	symbols, err := ParseAndExtract("go", source, "demo.go")
	if err != nil {
		t.Fatalf("parse go: %v", err)
	}

	assertKinds(t, symbols, map[string]SymbolKind{
		"User":          SymbolKindStruct,
		"Store":         SymbolKindInterface,
		"Status":        SymbolKindType,
		"Alias":         SymbolKindType,
		"StatusOpen":    SymbolKindConst,
		"StatusClosed":  SymbolKindConst,
		"DefaultStatus": SymbolKindConst,
		"Build":         SymbolKindFunc,
		"Label":         SymbolKindFunc,
	})
}

func TestParseAndExtractRustKinds(t *testing.T) {
	ensureInstalled(t, "rust")

	source := []byte(`pub trait Reader { fn read(&self); }
pub struct User;
pub enum Status { Open, Closed }
const LIMIT: usize = 10;
static CACHE: usize = 1;
`)

	symbols, err := ParseAndExtract("rust", source, "demo.rs")
	if err != nil {
		t.Fatalf("parse rust kinds: %v", err)
	}

	assertKinds(t, symbols, map[string]SymbolKind{
		"Reader": SymbolKindInterface,
		"User":   SymbolKindStruct,
		"Status": SymbolKindEnum,
		"Open":   SymbolKindConst,
		"Closed": SymbolKindConst,
		"LIMIT":  SymbolKindConst,
		"CACHE":  SymbolKindConst,
	})
}

func TestParseAndExtractTypeScriptKinds(t *testing.T) {
	ensureInstalled(t, "typescript")

	source := []byte(`namespace App {}
enum Color { Red, Blue = 2 }
const answer = 42
const build = () => 1
`)

	symbols, err := ParseAndExtract("typescript", source, "demo.ts")
	if err != nil {
		t.Fatalf("parse typescript kinds: %v", err)
	}

	assertKinds(t, symbols, map[string]SymbolKind{
		"App":    SymbolKindModule,
		"Color":  SymbolKindEnum,
		"Red":    SymbolKindConst,
		"Blue":   SymbolKindConst,
		"answer": SymbolKindConst,
		"build":  SymbolKindFunc,
	})
}

func TestParseAndExtractJavaKinds(t *testing.T) {
	ensureInstalled(t, "java")

	moduleSymbols, err := ParseAndExtract("java", []byte("module demo.core {}"), "module-info.java")
	if err != nil {
		t.Fatalf("parse java module: %v", err)
	}
	assertKinds(t, moduleSymbols, map[string]SymbolKind{
		"demo.core": SymbolKindModule,
	})

	source := []byte(`class App { static final int LIMIT = 10, MAX = 20; }
enum Status { OPEN, CLOSED }
`)
	symbols, err := ParseAndExtract("java", source, "App.java")
	if err != nil {
		t.Fatalf("parse java kinds: %v", err)
	}

	assertKinds(t, symbols, map[string]SymbolKind{
		"App":    SymbolKindClass,
		"LIMIT":  SymbolKindConst,
		"MAX":    SymbolKindConst,
		"Status": SymbolKindEnum,
		"OPEN":   SymbolKindConst,
		"CLOSED": SymbolKindConst,
	})
}

func TestParseAndExtractCppKinds(t *testing.T) {
	ensureInstalled(t, "cpp")

	source := []byte(`namespace api {}
struct User {};
class Store {};
enum Color { Red };
typedef int Count;
`)

	symbols, err := ParseAndExtract("cpp", source, "demo.cpp")
	if err != nil {
		t.Fatalf("parse cpp kinds: %v", err)
	}

	assertKinds(t, symbols, map[string]SymbolKind{
		"api":   SymbolKindModule,
		"User":  SymbolKindStruct,
		"Store": SymbolKindClass,
		"Color": SymbolKindEnum,
		"Count": SymbolKindType,
	})
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

func TestFindCalleesGo(t *testing.T) {
	ensureInstalled(t, "go")

	source := []byte("package main\n\nimport \"fmt\"\n\nfunc A() {\n\thelper()\n\tfmt.Println(\"hello\")\n}\n")
	symbols, err := ParseAndExtract("go", source, "main.go")
	if err != nil {
		t.Fatalf("parse go symbols: %v", err)
	}
	if len(symbols) == 0 {
		t.Fatal("expected at least one symbol")
	}

	callees, err := FindCallees("go", source, "main.go", symbols[0].ByteStart, symbols[0].ByteEnd)
	if err != nil {
		t.Fatalf("find callees: %v", err)
	}
	if len(callees) != 2 {
		t.Fatalf("expected 2 callees, got %d", len(callees))
	}
	if callees[0].Name != "helper" || callees[0].Line != 6 {
		t.Fatalf("unexpected first callee: %+v", callees[0])
	}
	if callees[1].Name != "fmt.Println" || callees[1].Line != 7 {
		t.Fatalf("unexpected second callee: %+v", callees[1])
	}
}

func TestFindCalleesTypeScript(t *testing.T) {
	ensureInstalled(t, "typescript")

	source := []byte("function A() {\n  helper()\n  console.log('hello')\n}\n")
	symbols, err := ParseAndExtract("typescript", source, "main.ts")
	if err != nil {
		t.Fatalf("parse typescript symbols: %v", err)
	}
	if len(symbols) == 0 {
		t.Fatal("expected at least one symbol")
	}

	callees, err := FindCallees("typescript", source, "main.ts", symbols[0].ByteStart, symbols[0].ByteEnd)
	if err != nil {
		t.Fatalf("find callees: %v", err)
	}
	if len(callees) != 2 {
		t.Fatalf("expected 2 callees, got %d", len(callees))
	}
	if callees[0].Name != "helper" {
		t.Fatalf("unexpected first callee: %+v", callees[0])
	}
	if callees[1].Name != "console.log" {
		t.Fatalf("unexpected second callee: %+v", callees[1])
	}
}

func TestDetectLanguageProtobuf(t *testing.T) {
	if got := DetectLanguage("api/greeter.proto"); got != "protobuf" {
		t.Fatalf("unexpected protobuf detection: %q", got)
	}
}

func TestDetectLanguageFromSourceUsesShebangForExtensionlessScripts(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		source string
		want   string
	}{
		{
			name:   "python env",
			path:   "script",
			source: "#!/usr/bin/env python3\n\ndef main():\n    pass\n",
			want:   "python",
		},
		{
			name:   "bash direct",
			path:   "run",
			source: "#!/bin/bash\nmain() { echo ok; }\n",
			want:   "bash",
		},
		{
			name:   "extension wins",
			path:   "script.py",
			source: "#!/bin/bash\n",
			want:   "python",
		},
		{
			name:   "unsupported extension ignores shebang",
			path:   "script.txt",
			source: "#!/usr/bin/env python3\n",
			want:   "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := DetectLanguageFromSource(test.path, []byte(test.source)); got != test.want {
				t.Fatalf("unexpected language: got %q want %q", got, test.want)
			}
		})
	}
}

func assertKinds(t *testing.T, symbols []Symbol, expected map[string]SymbolKind) {
	t.Helper()

	actual := make(map[string]SymbolKind, len(symbols))
	for _, symbol := range symbols {
		actual[symbol.Name] = symbol.Kind
	}

	for name, kind := range expected {
		if actual[name] != kind {
			t.Fatalf("unexpected kind for %s: got %q want %q", name, actual[name], kind)
		}
	}
}

func TestNotInstalledMessageUsesGxCommand(t *testing.T) {
	err := newNotInstalled("rust")
	if !strings.Contains(err.Error(), "gx lang enable rust") {
		t.Fatalf("expected gx install hint, got %q", err.Error())
	}
	if strings.Contains(err.Error(), "cx lang enable rust") {
		t.Fatalf("unexpected cx install hint, got %q", err.Error())
	}
}
