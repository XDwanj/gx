package language

import (
	"io"
	"strings"
	"testing"
	"time"

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

func TestParseAndExtractSupportedLanguageSmoke(t *testing.T) {
	tests := []struct {
		language string
		path     string
		source   string
		expected map[string]SymbolKind
	}{
		{
			language: "bash",
			path:     "build.sh",
			source:   "build() { echo ok; }\n",
			expected: map[string]SymbolKind{"build": SymbolKindFunc},
		},
		{
			language: "c",
			path:     "demo.c",
			source:   "int add(int left, int right) { return left + right; }\n",
			expected: map[string]SymbolKind{"add": SymbolKindFunc},
		},
		{
			language: "cpp",
			path:     "demo.cpp",
			source:   "struct User {};\nint add(int left, int right) { return left + right; }\n",
			expected: map[string]SymbolKind{"User": SymbolKindStruct, "add": SymbolKindFunc},
		},
		{
			language: "go",
			path:     "demo.go",
			source:   "package demo\n\ntype User struct{}\nfunc Build() {}\n",
			expected: map[string]SymbolKind{"User": SymbolKindStruct, "Build": SymbolKindFunc},
		},
		{
			language: "java",
			path:     "Demo.java",
			source:   "class User { void build() {} }\n",
			expected: map[string]SymbolKind{"User": SymbolKindClass, "build": SymbolKindFunc},
		},
		{
			language: "kotlin",
			path:     "Demo.kt",
			source:   "class User\nfun build() {}\n",
			expected: map[string]SymbolKind{"User": SymbolKindClass, "build": SymbolKindFunc},
		},
		{
			language: "lua",
			path:     "demo.lua",
			source:   "function build() end\n",
			expected: map[string]SymbolKind{"build": SymbolKindFunc},
		},
		{
			language: "protobuf",
			path:     "demo.proto",
			source:   "syntax = \"proto3\";\nmessage User {}\n",
			expected: map[string]SymbolKind{"User": SymbolKindStruct},
		},
		{
			language: "python",
			path:     "demo.py",
			source:   "class User:\n    pass\n\ndef build():\n    pass\n",
			expected: map[string]SymbolKind{"User": SymbolKindClass, "build": SymbolKindFunc},
		},
		{
			language: "ruby",
			path:     "demo.rb",
			source:   "class User\nend\ndef build\nend\n",
			expected: map[string]SymbolKind{"User": SymbolKindClass, "build": SymbolKindFunc},
		},
		{
			language: "rust",
			path:     "demo.rs",
			source:   "struct User;\nfn build() {}\n",
			expected: map[string]SymbolKind{"User": SymbolKindStruct, "build": SymbolKindFunc},
		},
		{
			language: "swift",
			path:     "Demo.swift",
			source:   "struct User {}\nfunc build() {}\n",
			expected: map[string]SymbolKind{"User": SymbolKindStruct, "build": SymbolKindFunc},
		},
		{
			language: "typescript",
			path:     "demo.ts",
			source:   "class User {}\nfunction build() {}\n",
			expected: map[string]SymbolKind{"User": SymbolKindClass, "build": SymbolKindFunc},
		},
		{
			language: "vue",
			path:     "Demo.vue",
			source: `<template>
  <main>{{ title }}</main>
</template>
<script setup lang="ts">
const title = "gx"
</script>
<style scoped>
.title { color: red; }
</style>
`,
			expected: map[string]SymbolKind{"template": SymbolKindModule, "script": SymbolKindModule, "style": SymbolKindModule},
		},
		{
			language: "zig",
			path:     "demo.zig",
			source:   "pub fn build() void {}\n",
			expected: map[string]SymbolKind{"build": SymbolKindFunc},
		},
	}

	for _, test := range tests {
		t.Run(test.language, func(t *testing.T) {
			ensureInstalled(t, test.language)

			symbols, err := ParseAndExtract(test.language, []byte(test.source), test.path)
			if err != nil {
				t.Fatalf("parse %s: %v", test.language, err)
			}
			assertKinds(t, symbols, test.expected)
		})
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

func TestParseAndExtractReportsParseTimeout(t *testing.T) {
	ensureInstalled(t, "protobuf")

	previousTimeout := parseTimeout
	parseTimeout = time.Microsecond
	t.Cleanup(func() {
		parseTimeout = previousTimeout
	})

	var builder strings.Builder
	builder.WriteString("syntax = \"proto3\";\n")
	for i := range 5000 {
		builder.WriteString("message Item")
		builder.WriteString(string(rune('A' + i%26)))
		builder.WriteString(" { string name = 1; }\n")
	}

	_, err := ParseAndExtract("protobuf", []byte(builder.String()), "large.proto")
	if !IsParseTimedOut(err) {
		t.Fatalf("ParseAndExtract error = %v, want parse timeout", err)
	}
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

func TestParseAndExtractKotlinKinds(t *testing.T) {
	ensureInstalled(t, "kotlin")

	source := []byte(`interface Repository
class User
object Defaults
enum class Status { Open, Closed }
typealias UserId = String
const val Limit = 10
fun buildUser(): User = User()
`)
	symbols, err := ParseAndExtract("kotlin", source, "Demo.kt")
	if err != nil {
		t.Fatalf("parse kotlin kinds: %v", err)
	}

	assertKinds(t, symbols, map[string]SymbolKind{
		"Repository": SymbolKindInterface,
		"User":       SymbolKindClass,
		"Defaults":   SymbolKindClass,
		"Status":     SymbolKindEnum,
		"Open":       SymbolKindConst,
		"Closed":     SymbolKindConst,
		"UserId":     SymbolKindType,
		"Limit":      SymbolKindConst,
		"buildUser":  SymbolKindFunc,
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

func TestDetectLanguageKotlin(t *testing.T) {
	if got := DetectLanguage("src/main/kotlin/App.kt"); got != "kotlin" {
		t.Fatalf("unexpected kotlin detection for .kt: %q", got)
	}
	if got := DetectLanguage("scripts/build.kts"); got != "kotlin" {
		t.Fatalf("unexpected kotlin detection for .kts: %q", got)
	}
}

func TestDetectLanguageVue(t *testing.T) {
	if got := DetectLanguage("src/components/App.vue"); got != "vue" {
		t.Fatalf("unexpected vue detection: %q", got)
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

func TestFindCalleesKotlin(t *testing.T) {
	ensureInstalled(t, "kotlin")

	source := []byte("fun render() {\n\tprintln(formatName(\"gx\"))\n}\n")
	symbols, err := ParseAndExtract("kotlin", source, "main.kt")
	if err != nil {
		t.Fatalf("parse kotlin symbols: %v", err)
	}
	if len(symbols) == 0 {
		t.Fatal("expected at least one symbol")
	}

	callees, err := FindCallees("kotlin", source, "main.kt", symbols[0].ByteStart, symbols[0].ByteEnd)
	if err != nil {
		t.Fatalf("find kotlin callees: %v", err)
	}
	if len(callees) != 2 {
		t.Fatalf("expected 2 callees, got %d: %+v", len(callees), callees)
	}
	if callees[0].Name != "println" || callees[1].Name != "formatName" {
		t.Fatalf("unexpected callees: %+v", callees)
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
