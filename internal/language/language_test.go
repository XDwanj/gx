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

func TestLeadingTriviaStart(t *testing.T) {
	tests := []struct {
		name     string
		lang     string
		source   string
		symbol   string
		wantBody string
	}{
		{
			name:     "go doc comment",
			lang:     "go",
			source:   "package demo\n\n// NewServer builds a server.\nfunc NewServer() {}\n",
			symbol:   "NewServer",
			wantBody: "// NewServer builds a server.\nfunc NewServer() {}",
		},
		{
			name:     "go detached comment",
			lang:     "go",
			source:   "package demo\n\n// not a doc comment.\n\nfunc NewServer() {}\n",
			symbol:   "NewServer",
			wantBody: "func NewServer() {}",
		},
		{
			name:     "go trailing comment",
			lang:     "go",
			source:   "package demo\n\nfunc helper() {} // note\nfunc NewServer() {}\n",
			symbol:   "NewServer",
			wantBody: "func NewServer() {}",
		},
		{
			name:     "rust doc plus derive",
			lang:     "rust",
			source:   "/// A server.\n#[derive(Debug)]\nstruct Server {\n}\n",
			symbol:   "Server",
			wantBody: "/// A server.\n#[derive(Debug)]\nstruct Server {\n}",
		},
		{
			name:     "python decorated",
			lang:     "python",
			source:   "# cached getter\n@cache\ndef get():\n    pass\n",
			symbol:   "get",
			wantBody: "# cached getter\n@cache\ndef get():\n    pass",
		},
		{
			name:     "bash shebang",
			lang:     "bash",
			source:   "#!/usr/bin/env bash\nhello() {\n  echo hi\n}\n",
			symbol:   "hello",
			wantBody: "hello() {\n  echo hi\n}",
		},
		{
			name:     "typescript jsdoc",
			lang:     "typescript",
			source:   "/** Adds a and b. */\nfunction add(a: number, b: number) {\n  return a + b;\n}\n",
			symbol:   "add",
			wantBody: "/** Adds a and b. */\nfunction add(a: number, b: number) {\n  return a + b;\n}",
		},
		{
			name:     "c block comment",
			lang:     "c",
			source:   "/* Adds. */\nint add(int a, int b) {\n  return a + b;\n}\n",
			symbol:   "add",
			wantBody: "/* Adds. */\nint add(int a, int b) {\n  return a + b;\n}",
		},
		{
			name:     "java method doc",
			lang:     "java",
			source:   "class A {\n  /** Adds. */\n  int add(int a, int b) { return a + b; }\n}\n",
			symbol:   "add",
			wantBody: "/** Adds. */\n  int add(int a, int b) { return a + b; }",
		},
		{
			name:     "kotlin kdoc",
			lang:     "kotlin",
			source:   "/** Adds. */\nfun add(a: Int, b: Int) = a + b\n",
			symbol:   "add",
			wantBody: "/** Adds. */\nfun add(a: Int, b: Int) = a + b",
		},
		{
			name:     "ruby comment",
			lang:     "ruby",
			source:   "# Adds.\ndef add(a, b)\n  a + b\nend\n",
			symbol:   "add",
			wantBody: "# Adds.\ndef add(a, b)\n  a + b\nend",
		},
		{
			name:     "lua doc comment",
			lang:     "lua",
			source:   "--- Adds.\nlocal function add(a, b)\n  return a + b\nend\n",
			symbol:   "add",
			wantBody: "--- Adds.\nlocal function add(a, b)\n  return a + b\nend",
		},
		{
			name:     "zig doc comment",
			lang:     "zig",
			source:   "/// Adds.\nfn add(a: i32, b: i32) i32 {\n  return a + b;\n}\n",
			symbol:   "add",
			wantBody: "/// Adds.\nfn add(a: i32, b: i32) i32 {\n  return a + b;\n}",
		},
		{
			name:     "swift doc comment",
			lang:     "swift",
			source:   "/// Adds.\nfunc add(a: Int, b: Int) -> Int {\n  return a + b\n}\n",
			symbol:   "add",
			wantBody: "/// Adds.\nfunc add(a: Int, b: Int) -> Int {\n  return a + b\n}",
		},
		{
			name:     "protobuf comment",
			lang:     "protobuf",
			source:   "syntax = \"proto3\";\n\n// A message.\nmessage Foo {\n  int32 a = 1;\n}\n",
			symbol:   "Foo",
			wantBody: "// A message.\nmessage Foo {\n  int32 a = 1;\n}",
		},
		{
			name:     "vue block comment",
			lang:     "vue",
			source:   "<!-- app root -->\n<template>\n  <div/>\n</template>\n",
			symbol:   "template",
			wantBody: "<!-- app root -->\n<template>\n  <div/>\n</template>",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			symbols, err := ParseAndExtract(test.lang, []byte(test.source), "x")
			if err != nil {
				t.Fatalf("ParseAndExtract: %v", err)
			}
			for _, symbol := range symbols {
				if symbol.Name != test.symbol {
					continue
				}
				body := test.source[symbol.ByteStart:symbol.ByteEnd]
				if body != test.wantBody {
					t.Fatalf("body = %q, want %q", body, test.wantBody)
				}
				return
			}
			t.Fatalf("symbol %q not found in %+v", test.symbol, symbols)
		})
	}
}

const vueSource = `<template>
  <a-modal @ok="handleDelete(record.id)" :visible="showModal">
    {{ formatName(user.name) }}
  </a-modal>
</template>
<script lang="ts" setup>
import { deleteBanner } from '@/api/banner'

const handleDelete = (id: string) => {
  deleteBanner({ id })
}
</script>
<style scoped></style>
`

func skipUnlessInstalled(t *testing.T, languageName string) {
	t.Helper()
	if _, _, _, _, err := parseSource(languageName, nil, "probe"); err != nil {
		t.Skipf("%s not installed: %v", languageName, err)
	}
	// 嵌入解析依赖 typescript grammar，缺省时 Vue/WXML 退化为块级结果。
	if _, _, _, _, err := parseSource("typescript", nil, "probe.ts"); err != nil {
		t.Skipf("typescript not installed: %v", err)
	}
}

func TestVueEmbeddedScriptSymbols(t *testing.T) {
	skipUnlessInstalled(t, "vue")
	symbols, err := ParseAndExtract("vue", []byte(vueSource), "index.vue")
	if err != nil {
		t.Fatalf("ParseAndExtract: %v", err)
	}

	kinds := map[string]SymbolKind{}
	for _, symbol := range symbols {
		kinds[symbol.Name] = symbol.Kind
	}
	if kinds["handleDelete"] != SymbolKindFunc {
		t.Fatalf("handleDelete kind = %q, want func (symbols: %+v)", kinds["handleDelete"], symbols)
	}
	for _, block := range []string{"template", "script", "style"} {
		if kinds[block] != SymbolKindModule {
			t.Fatalf("%s block missing (symbols: %+v)", block, symbols)
		}
	}
}

func TestVueEmbeddedReferences(t *testing.T) {
	skipUnlessInstalled(t, "vue")
	source := []byte(vueSource)

	references, err := FindReferencesForNames("vue", source, "index.vue", []string{"deleteBanner"})
	if err != nil {
		t.Fatalf("FindReferencesForNames: %v", err)
	}
	lines := map[int]bool{}
	for _, reference := range references {
		lines[reference.Line] = true
	}
	if !lines[7] || !lines[10] {
		t.Fatalf("expected deleteBanner import (line 7) and call (line 10), got %+v", references)
	}

	references, err = FindReferencesForNames("vue", source, "index.vue", []string{"handleDelete"})
	if err != nil {
		t.Fatalf("FindReferencesForNames: %v", err)
	}
	lines = map[int]bool{}
	for _, reference := range references {
		lines[reference.Line] = true
	}
	if !lines[2] || !lines[9] {
		t.Fatalf("expected @ok binding (line 2) and const (line 9), got %+v", references)
	}
}

func TestWxmlEmbeddedReferences(t *testing.T) {
	skipUnlessInstalled(t, "wxml")
	source := []byte(`<view wx:for="{{addresslist}}" wx:key="index">
  <button class="delete-btn" catchtap="deleteAddress" data-id="{{item.id}}">删除</button>
</view>
`)

	references, err := FindReferencesForNames("wxml", source, "address.wxml", []string{"deleteAddress"})
	if err != nil {
		t.Fatalf("FindReferencesForNames: %v", err)
	}
	if len(references) != 1 || references[0].Line != 2 {
		t.Fatalf("expected catchtap binding on line 2, got %+v", references)
	}

	names, err := FindReferenceNames("wxml", source, "address.wxml")
	if err != nil {
		t.Fatalf("FindReferenceNames: %v", err)
	}
	seen := map[string]bool{}
	for _, name := range names {
		seen[name] = true
	}
	for _, want := range []string{"item", "addresslist", "deleteAddress"} {
		if !seen[want] {
			t.Fatalf("expected reference name %q in %v", want, names)
		}
	}
}
