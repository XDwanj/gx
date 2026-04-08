package language

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	tree_sitter_swift "github.com/XDwanj/tree-sitter-swift/bindings/go"
	tree_sitter_zig "github.com/XDwanj/tree-sitter-zig/bindings/go"
	tree_sitter_proto "github.com/coder3101/tree-sitter-proto/bindings/go"
	tree_sitter_lua "github.com/tree-sitter-grammars/tree-sitter-lua/bindings/go"
	sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_bash "github.com/tree-sitter/tree-sitter-bash/bindings/go"
	tree_sitter_c "github.com/tree-sitter/tree-sitter-c/bindings/go"
	tree_sitter_cpp "github.com/tree-sitter/tree-sitter-cpp/bindings/go"
	tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
	tree_sitter_java "github.com/tree-sitter/tree-sitter-java/bindings/go"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
	tree_sitter_ruby "github.com/tree-sitter/tree-sitter-ruby/bindings/go"
	tree_sitter_rust "github.com/tree-sitter/tree-sitter-rust/bindings/go"
	tree_sitter_typescript "github.com/tree-sitter/tree-sitter-typescript/bindings/go"

	langpkg "github.com/XDwanj/gx/internal/lang"
)

const (
	jsxExtension          = "jsx"
	tsxExtension          = "tsx"
	grammarNameTSX        = "tsx"
	grammarNameTypeScript = "typescript"
)

type KindOverride struct {
	CaptureName string
	NodeKind    string
	Kind        SymbolKind
}

type SymbolKind string

const (
	SymbolKindFn        SymbolKind = "fn"
	SymbolKindStruct    SymbolKind = "struct"
	SymbolKindEnum      SymbolKind = "enum"
	SymbolKindType      SymbolKind = "type"
	SymbolKindConst     SymbolKind = "const"
	SymbolKindClass     SymbolKind = "class"
	SymbolKindInterface SymbolKind = "interface"
	SymbolKindMethod    SymbolKind = "method"
	SymbolKindModule    SymbolKind = "module"
)

type Symbol struct {
	Name      string
	Kind      SymbolKind
	Signature string
	Line      int
	ByteStart uint
	ByteEnd   uint
	IsTest    bool
}

type Config struct {
	Name          string
	Extensions    []string
	Query         string
	SigBodyChild  string
	SigDelimiter  byte
	KindOverrides []KindOverride
	RefNodeTypes  []string
	grammarName   func(ext string) string
	loadLanguage  func(ext string) *sitter.Language
}

type LangError struct {
	Name string
	Kind string
}

func (errorValue *LangError) Error() string {
	if errorValue.Kind == "not-installed" {
		return fmt.Sprintf("%s grammar not installed — run: gx lang enable %s", errorValue.Name, errorValue.Name)
	}
	return "parse failed"
}

func isNotInstalled(err error) bool {
	typed, ok := err.(*LangError)
	return ok && typed.Kind == "not-installed"
}

func newNotInstalled(name string) error {
	return &LangError{Name: name, Kind: "not-installed"}
}

func newParseFailed() error {
	return &LangError{Kind: "parse-failed"}
}

var queryCache sync.Map

var languages = []*Config{
	{
		Name:         "rust",
		Extensions:   []string{"rs"},
		Query:        rustQuery,
		SigDelimiter: '{',
		RefNodeTypes: []string{"identifier", "type_identifier", "field_identifier"},
		grammarName:  func(_ string) string { return "rust" },
		loadLanguage: func(_ string) *sitter.Language {
			return sitter.NewLanguage(tree_sitter_rust.Language())
		},
	},
	{
		Name:         "typescript",
		Extensions:   []string{"ts", "tsx", "js", "jsx"},
		Query:        typeScriptQuery,
		SigDelimiter: '{',
		RefNodeTypes: []string{"identifier", "type_identifier", "property_identifier", "shorthand_property_identifier", "shorthand_property_identifier_pattern"},
		grammarName: func(ext string) string {
			if isTSXFamilyExtension(ext) {
				return grammarNameTSX
			}
			return grammarNameTypeScript
		},
		loadLanguage: func(ext string) *sitter.Language {
			if isTSXFamilyExtension(ext) {
				return sitter.NewLanguage(tree_sitter_typescript.LanguageTSX())
			}
			return sitter.NewLanguage(tree_sitter_typescript.LanguageTypescript())
		},
	},
	{
		Name:         "python",
		Extensions:   []string{"py"},
		Query:        pythonQuery,
		SigBodyChild: "block",
		RefNodeTypes: []string{"identifier"},
		grammarName:  func(_ string) string { return "python" },
		loadLanguage: func(_ string) *sitter.Language {
			return sitter.NewLanguage(tree_sitter_python.Language())
		},
	},
	{
		Name:         "protobuf",
		Extensions:   []string{"proto"},
		Query:        protobufQuery,
		RefNodeTypes: []string{"identifier"},
		grammarName:  func(_ string) string { return "protobuf" },
		loadLanguage: func(_ string) *sitter.Language {
			return sitter.NewLanguage(tree_sitter_proto.Language())
		},
	},
	{
		Name:         "go",
		Extensions:   []string{"go"},
		Query:        goQuery,
		SigDelimiter: '{',
		RefNodeTypes: []string{"identifier", "type_identifier", "field_identifier"},
		grammarName:  func(_ string) string { return "go" },
		loadLanguage: func(_ string) *sitter.Language {
			return sitter.NewLanguage(tree_sitter_go.Language())
		},
	},
	{
		Name:         "c",
		Extensions:   []string{"c"},
		Query:        cQuery,
		SigDelimiter: '{',
		RefNodeTypes: []string{"identifier", "type_identifier", "field_identifier"},
		grammarName:  func(_ string) string { return "c" },
		loadLanguage: func(_ string) *sitter.Language { return sitter.NewLanguage(tree_sitter_c.Language()) },
	},
	{
		Name:         "cpp",
		Extensions:   []string{"cpp", "cc", "cxx", "h", "hpp", "hxx", "hh"},
		Query:        cppQuery,
		SigDelimiter: '{',
		RefNodeTypes: []string{"identifier", "type_identifier", "field_identifier"},
		grammarName:  func(_ string) string { return "cpp" },
		loadLanguage: func(_ string) *sitter.Language {
			return sitter.NewLanguage(tree_sitter_cpp.Language())
		},
	},
	{
		Name:         "java",
		Extensions:   []string{"java"},
		Query:        javaQuery,
		SigDelimiter: '{',
		RefNodeTypes: []string{"identifier"},
		grammarName:  func(_ string) string { return "java" },
		loadLanguage: func(_ string) *sitter.Language {
			return sitter.NewLanguage(tree_sitter_java.Language())
		},
	},
	{
		Name:         "ruby",
		Extensions:   []string{"rb"},
		Query:        rubyQuery,
		RefNodeTypes: []string{"identifier", "constant"},
		grammarName:  func(_ string) string { return "ruby" },
		loadLanguage: func(_ string) *sitter.Language {
			return sitter.NewLanguage(tree_sitter_ruby.Language())
		},
	},
	{
		Name:         "lua",
		Extensions:   []string{"lua"},
		Query:        luaQuery,
		RefNodeTypes: []string{"identifier"},
		grammarName:  func(_ string) string { return "lua" },
		loadLanguage: func(_ string) *sitter.Language {
			return sitter.NewLanguage(tree_sitter_lua.Language())
		},
	},
	{
		Name:         "zig",
		Extensions:   []string{"zig"},
		Query:        zigQuery,
		SigDelimiter: '{',
		RefNodeTypes: []string{"IDENTIFIER"},
		grammarName:  func(_ string) string { return "zig" },
		loadLanguage: func(_ string) *sitter.Language {
			return sitter.NewLanguage(tree_sitter_zig.Language())
		},
	},
	{
		Name:         "bash",
		Extensions:   []string{"sh", "bash"},
		Query:        bashQuery,
		SigDelimiter: '{',
		RefNodeTypes: []string{"word"},
		grammarName:  func(_ string) string { return "bash" },
		loadLanguage: func(_ string) *sitter.Language {
			return sitter.NewLanguage(tree_sitter_bash.Language())
		},
	},
	{
		Name:         "swift",
		Extensions:   []string{"swift"},
		Query:        swiftQuery,
		SigDelimiter: '{',
		RefNodeTypes: []string{"simple_identifier", "type_identifier"},
		grammarName:  func(_ string) string { return "swift" },
		loadLanguage: func(_ string) *sitter.Language {
			return sitter.NewLanguage(tree_sitter_swift.Language())
		},
	},
}

func isTSXFamilyExtension(ext string) bool {
	return ext == tsxExtension || ext == jsxExtension
}

func DetectLanguage(path string) string {
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	for _, config := range languages {
		for _, candidate := range config.Extensions {
			if candidate == ext {
				return config.Name
			}
		}
	}
	return ""
}

func SupportedLanguages() []string {
	return langpkg.SupportedLanguages()
}

func PrimaryExtension(languageName string) string {
	for _, config := range languages {
		if config.Name == languageName && len(config.Extensions) > 0 {
			return config.Extensions[0]
		}
	}
	return languageName
}

func ParseAndExtract(languageName string, source []byte, path string) ([]Symbol, error) {
	config, tree, grammarKey, err := parseSource(languageName, source, path)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	query, err := compiledQuery(config, grammarKey, path)
	if err != nil {
		return nil, err
	}

	return extractSymbols(config, query, tree, source), nil
}

func FindReferences(languageName string, source []byte, path string, name string) ([]Reference, error) {
	config, tree, _, err := parseSource(languageName, source, path)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	references := make([]Reference, 0)
	walkReferenceLeaves(config, tree.RootNode(), source, func(node *sitter.Node) {
		if node.Utf8Text(source) != name {
			return
		}
		references = append(references, Reference{
			Line:       int(node.StartPosition().Row) + 1,
			ByteOffset: node.StartByte(),
		})
	})

	return references, nil
}

func FindReferenceNames(languageName string, source []byte, path string) ([]string, error) {
	config, tree, _, err := parseSource(languageName, source, path)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	names := make([]string, 0)
	seen := make(map[string]struct{})
	walkReferenceLeaves(config, tree.RootNode(), source, func(node *sitter.Node) {
		name := node.Utf8Text(source)
		if name == "" {
			return
		}
		if _, exists := seen[name]; exists {
			return
		}
		seen[name] = struct{}{}
		names = append(names, name)
	})
	sort.Strings(names)
	return names, nil
}

func walkReferenceLeaves(config *Config, root *sitter.Node, source []byte, visit func(node *sitter.Node)) {
	stack := []*sitter.Node{root}
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if node.ChildCount() == 0 && containsString(config.RefNodeTypes, node.Kind()) {
			visit(node)
			continue
		}

		for childIndex := int(node.ChildCount()) - 1; childIndex >= 0; childIndex-- {
			child := node.Child(uint(childIndex))
			if child != nil {
				stack = append(stack, child)
			}
		}
	}
}

func IsNotInstalled(err error) bool {
	return isNotInstalled(err)
}

func parseSource(languageName string, source []byte, path string) (*Config, *sitter.Tree, string, error) {
	config := configFor(languageName)
	if config == nil {
		return nil, nil, "", newNotInstalled(languageName)
	}
	if !langpkg.IsInstalled(config.Name) {
		return nil, nil, "", newNotInstalled(config.Name)
	}

	parser := sitter.NewParser()
	defer parser.Close()

	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	languageValue := config.loadLanguage(ext)
	if languageValue == nil {
		return nil, nil, "", newParseFailed()
	}
	if err := parser.SetLanguage(languageValue); err != nil {
		return nil, nil, "", newParseFailed()
	}

	tree := parser.Parse(source, nil)
	if tree == nil {
		return nil, nil, "", newParseFailed()
	}

	return config, tree, config.grammarName(ext), nil
}

func compiledQuery(config *Config, grammarKey string, path string) (*sitter.Query, error) {
	cacheKey := config.Name + ":" + grammarKey
	if cached, ok := queryCache.Load(cacheKey); ok {
		return cached.(*sitter.Query), nil
	}

	query, queryErr := sitter.NewQuery(config.loadLanguage(strings.TrimPrefix(filepath.Ext(path), ".")), config.Query)
	if queryErr != nil {
		return nil, queryErr
	}

	actual, _ := queryCache.LoadOrStore(cacheKey, query)
	return actual.(*sitter.Query), nil
}

func configFor(languageName string) *Config {
	for _, config := range languages {
		if config.Name == languageName {
			return config
		}
	}
	return nil
}

func containsString(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}
