package language

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	tree_sitter_swift "github.com/XDwanj/tree-sitter-swift/bindings/go"
	tree_sitter_zig "github.com/XDwanj/tree-sitter-zig/bindings/go"
	tree_sitter_proto "github.com/coder3101/tree-sitter-proto/bindings/go"
	tree_sitter_kotlin "github.com/tree-sitter-grammars/tree-sitter-kotlin/bindings/go"
	tree_sitter_lua "github.com/tree-sitter-grammars/tree-sitter-lua/bindings/go"
	sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_bash "github.com/tree-sitter/tree-sitter-bash/bindings/go"
	tree_sitter_c "github.com/tree-sitter/tree-sitter-c/bindings/go"
	tree_sitter_cpp "github.com/tree-sitter/tree-sitter-cpp/bindings/go"
	tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
	tree_sitter_html "github.com/tree-sitter/tree-sitter-html/bindings/go"
	tree_sitter_java "github.com/tree-sitter/tree-sitter-java/bindings/go"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
	tree_sitter_ruby "github.com/tree-sitter/tree-sitter-ruby/bindings/go"
	tree_sitter_rust "github.com/tree-sitter/tree-sitter-rust/bindings/go"
	tree_sitter_typescript "github.com/tree-sitter/tree-sitter-typescript/bindings/go"

	langpkg "github.com/XDwanj/gx/internal/lang"
)

const (
	jsxExtension            = "jsx"
	tsxExtension            = "tsx"
	languageNameBash        = "bash"
	languageNameCPP         = "cpp"
	languageNameGo          = "go"
	languageNameJava        = "java"
	languageNameKotlin      = "kotlin"
	languageNameLua         = "lua"
	languageNamePython      = "python"
	languageNameProtobuf    = "protobuf"
	languageNameRuby        = "ruby"
	languageNameRust        = "rust"
	languageNameSwift       = "swift"
	languageNameVue         = "vue"
	languageNameZig         = "zig"
	grammarNameTSX          = "tsx"
	grammarNameTypeScript   = "typescript"
	langErrorNotInstalled   = "not-installed"
	langErrorParseTimeout   = "parse-timeout"
	nodeKindCall            = "call"
	nodeKindCallExpression  = "call_expression"
	nodeKindFieldIdentifier = "field_identifier"
	nodeKindIdentifier      = "identifier"
	nodeKindTypeIdentifier  = "type_identifier"
	defaultParseTimeout     = 5 * time.Second
)

type KindOverride struct {
	CaptureName string
	NodeKind    string
	Kind        SymbolKind
}

type SymbolKind string

const (
	SymbolKindFunc      SymbolKind = "func"
	SymbolKindStruct    SymbolKind = "struct"
	SymbolKindEnum      SymbolKind = "enum"
	SymbolKindType      SymbolKind = "type"
	SymbolKindConst     SymbolKind = "const"
	SymbolKindClass     SymbolKind = "class"
	SymbolKindInterface SymbolKind = "interface"
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
	CallNodeTypes []string
	grammarName   func(ext string) string
	loadLanguage  func(ext string) *sitter.Language
}

type LangError struct {
	Name string
	Kind string
}

func (errorValue *LangError) Error() string {
	if errorValue.Kind == langErrorNotInstalled {
		return fmt.Sprintf("%s grammar not installed — run: gx lang enable %s", errorValue.Name, errorValue.Name)
	}
	if errorValue.Kind == langErrorParseTimeout {
		return fmt.Sprintf("%s parse timed out", errorValue.Name)
	}
	return "parse failed"
}

func isNotInstalled(err error) bool {
	typed, ok := err.(*LangError)
	return ok && typed.Kind == langErrorNotInstalled
}

func newNotInstalled(name string) error {
	return &LangError{Name: name, Kind: langErrorNotInstalled}
}

func newParseFailed() error {
	return &LangError{Kind: "parse-failed"}
}

func newParseTimedOut(name string) error {
	return &LangError{Name: name, Kind: langErrorParseTimeout}
}

var (
	queryCache   sync.Map
	parseTimeout = defaultParseTimeout
)

var languages = []*Config{
	{
		Name:         languageNameRust,
		Extensions:   []string{"rs"},
		Query:        rustQuery,
		SigDelimiter: '{',
		RefNodeTypes: []string{nodeKindIdentifier, nodeKindTypeIdentifier, nodeKindFieldIdentifier},
		CallNodeTypes: []string{
			nodeKindCallExpression,
		},
		grammarName: func(_ string) string { return languageNameRust },
		loadLanguage: func(_ string) *sitter.Language {
			return sitter.NewLanguage(tree_sitter_rust.Language())
		},
	},
	{
		Name:         "typescript",
		Extensions:   []string{"ts", "tsx", "js", "jsx"},
		Query:        typeScriptQuery,
		SigDelimiter: '{',
		RefNodeTypes: []string{
			nodeKindIdentifier,
			nodeKindTypeIdentifier,
			"property_identifier",
			"shorthand_property_identifier",
			"shorthand_property_identifier_pattern",
		},
		CallNodeTypes: []string{
			nodeKindCallExpression,
		},
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
		Name:         languageNamePython,
		Extensions:   []string{"py"},
		Query:        pythonQuery,
		SigBodyChild: "block",
		RefNodeTypes: []string{nodeKindIdentifier},
		CallNodeTypes: []string{
			nodeKindCall,
		},
		grammarName: func(_ string) string { return languageNamePython },
		loadLanguage: func(_ string) *sitter.Language {
			return sitter.NewLanguage(tree_sitter_python.Language())
		},
	},
	{
		Name:         languageNameProtobuf,
		Extensions:   []string{"proto"},
		Query:        protobufQuery,
		RefNodeTypes: []string{nodeKindIdentifier},
		grammarName:  func(_ string) string { return languageNameProtobuf },
		loadLanguage: func(_ string) *sitter.Language {
			return sitter.NewLanguage(tree_sitter_proto.Language())
		},
	},
	{
		Name:         languageNameGo,
		Extensions:   []string{"go"},
		Query:        goQuery,
		SigDelimiter: '{',
		RefNodeTypes: []string{nodeKindIdentifier, nodeKindTypeIdentifier, nodeKindFieldIdentifier},
		CallNodeTypes: []string{
			nodeKindCallExpression,
		},
		grammarName: func(_ string) string { return languageNameGo },
		loadLanguage: func(_ string) *sitter.Language {
			return sitter.NewLanguage(tree_sitter_go.Language())
		},
	},
	{
		Name:         "c",
		Extensions:   []string{"c"},
		Query:        cQuery,
		SigDelimiter: '{',
		RefNodeTypes: []string{nodeKindIdentifier, nodeKindTypeIdentifier, nodeKindFieldIdentifier},
		CallNodeTypes: []string{
			nodeKindCallExpression,
		},
		grammarName:  func(_ string) string { return "c" },
		loadLanguage: func(_ string) *sitter.Language { return sitter.NewLanguage(tree_sitter_c.Language()) },
	},
	{
		Name:         languageNameCPP,
		Extensions:   []string{"cpp", "cc", "cxx", "h", "hpp", "hxx", "hh"},
		Query:        cppQuery,
		SigDelimiter: '{',
		RefNodeTypes: []string{nodeKindIdentifier, nodeKindTypeIdentifier, nodeKindFieldIdentifier},
		CallNodeTypes: []string{
			nodeKindCallExpression,
		},
		grammarName: func(_ string) string { return languageNameCPP },
		loadLanguage: func(_ string) *sitter.Language {
			return sitter.NewLanguage(tree_sitter_cpp.Language())
		},
	},
	{
		Name:         languageNameJava,
		Extensions:   []string{"java"},
		Query:        javaQuery,
		SigDelimiter: '{',
		RefNodeTypes: []string{nodeKindIdentifier},
		CallNodeTypes: []string{
			"method_invocation",
		},
		grammarName: func(_ string) string { return languageNameJava },
		loadLanguage: func(_ string) *sitter.Language {
			return sitter.NewLanguage(tree_sitter_java.Language())
		},
	},
	{
		Name:         languageNameKotlin,
		Extensions:   []string{"kt", "kts"},
		Query:        kotlinQuery,
		SigDelimiter: '{',
		RefNodeTypes: []string{nodeKindIdentifier},
		CallNodeTypes: []string{
			nodeKindCallExpression,
		},
		grammarName: func(_ string) string { return languageNameKotlin },
		loadLanguage: func(_ string) *sitter.Language {
			return sitter.NewLanguage(tree_sitter_kotlin.Language())
		},
	},
	{
		Name:         languageNameRuby,
		Extensions:   []string{"rb"},
		Query:        rubyQuery,
		RefNodeTypes: []string{nodeKindIdentifier, "constant"},
		CallNodeTypes: []string{
			nodeKindCall,
		},
		grammarName: func(_ string) string { return languageNameRuby },
		loadLanguage: func(_ string) *sitter.Language {
			return sitter.NewLanguage(tree_sitter_ruby.Language())
		},
	},
	{
		Name:         languageNameLua,
		Extensions:   []string{languageNameLua},
		Query:        luaQuery,
		RefNodeTypes: []string{nodeKindIdentifier},
		CallNodeTypes: []string{
			"function_call",
		},
		grammarName: func(_ string) string { return languageNameLua },
		loadLanguage: func(_ string) *sitter.Language {
			return sitter.NewLanguage(tree_sitter_lua.Language())
		},
	},
	{
		Name:         languageNameZig,
		Extensions:   []string{"zig"},
		Query:        zigQuery,
		SigDelimiter: '{',
		RefNodeTypes: []string{nodeKindIdentifier},
		grammarName:  func(_ string) string { return languageNameZig },
		loadLanguage: func(_ string) *sitter.Language {
			return sitter.NewLanguage(tree_sitter_zig.Language())
		},
	},
	{
		Name:         languageNameBash,
		Extensions:   []string{"sh", languageNameBash},
		Query:        bashQuery,
		SigDelimiter: '{',
		RefNodeTypes: []string{"word"},
		grammarName:  func(_ string) string { return languageNameBash },
		loadLanguage: func(_ string) *sitter.Language {
			return sitter.NewLanguage(tree_sitter_bash.Language())
		},
	},
	{
		Name:         languageNameSwift,
		Extensions:   []string{"swift"},
		Query:        swiftQuery,
		SigDelimiter: '{',
		RefNodeTypes: []string{"simple_identifier", "type_identifier"},
		CallNodeTypes: []string{
			"call_expression",
		},
		grammarName: func(_ string) string { return languageNameSwift },
		loadLanguage: func(_ string) *sitter.Language {
			return sitter.NewLanguage(tree_sitter_swift.Language())
		},
	},
	{
		Name:         languageNameVue,
		Extensions:   []string{languageNameVue},
		Query:        vueQuery,
		RefNodeTypes: []string{"tag_name", "attribute_name"},
		grammarName:  func(_ string) string { return languageNameVue },
		loadLanguage: func(_ string) *sitter.Language {
			return sitter.NewLanguage(tree_sitter_html.Language())
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

func DetectLanguageFromSource(path string, source []byte) string {
	if languageName := DetectLanguage(path); languageName != "" {
		return languageName
	}
	if strings.TrimPrefix(filepath.Ext(path), ".") != "" {
		return ""
	}
	return detectShebangLanguage(source)
}

func detectShebangLanguage(source []byte) string {
	line, _, _ := bytes.Cut(source, []byte("\n"))
	if !bytes.HasPrefix(line, []byte("#!")) {
		return ""
	}

	fields := strings.Fields(strings.TrimSpace(strings.TrimPrefix(string(line), "#!")))
	if len(fields) == 0 {
		return ""
	}

	interpreter := filepath.Base(fields[0])
	if interpreter == "env" {
		interpreter = envInterpreter(fields[1:])
	}

	switch {
	case strings.HasPrefix(interpreter, "python"):
		return languageNamePython
	case interpreter == "sh" || interpreter == languageNameBash || interpreter == "dash" || interpreter == "ksh" || interpreter == "zsh":
		return languageNameBash
	case interpreter == "ruby":
		return languageNameRuby
	case interpreter == "lua":
		return languageNameLua
	default:
		return ""
	}
}

func envInterpreter(args []string) string {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		return filepath.Base(arg)
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
	config, tree, languageValue, grammarKey, err := parseSource(languageName, source, path)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	query, err := compiledQuery(config, languageValue, grammarKey)
	if err != nil {
		return nil, err
	}

	return extractSymbols(config, query, tree, source), nil
}

func FindReferences(languageName string, source []byte, path string, name string) ([]Reference, error) {
	return FindReferencesForNames(languageName, source, path, []string{name})
}

func FindReferencesForNames(languageName string, source []byte, path string, names []string) ([]Reference, error) {
	config, tree, _, _, err := parseSource(languageName, source, path)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	nameSet := make(map[string]struct{}, len(names))
	for _, name := range names {
		if name == "" {
			continue
		}
		nameSet[name] = struct{}{}
	}
	if len(nameSet) == 0 {
		return []Reference{}, nil
	}

	references := make([]Reference, 0)
	walkReferenceLeaves(config, tree.RootNode(), func(node *sitter.Node) {
		if _, ok := nameSet[node.Utf8Text(source)]; !ok {
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
	config, tree, _, _, err := parseSource(languageName, source, path)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	names := make([]string, 0)
	seen := make(map[string]struct{})
	walkReferenceLeaves(config, tree.RootNode(), func(node *sitter.Node) {
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

func walkReferenceLeaves(config *Config, root *sitter.Node, visit func(node *sitter.Node)) {
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

func IsParseTimedOut(err error) bool {
	typed, ok := err.(*LangError)
	return ok && typed.Kind == langErrorParseTimeout
}

func parseSource(languageName string, source []byte, path string) (*Config, *sitter.Tree, *sitter.Language, string, error) {
	config := configFor(languageName)
	if config == nil {
		return nil, nil, nil, "", newNotInstalled(languageName)
	}
	if !langpkg.IsInstalled(config.Name) {
		return nil, nil, nil, "", newNotInstalled(config.Name)
	}

	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	languageValue := config.loadLanguage(ext)
	if languageValue == nil {
		return nil, nil, nil, "", newParseFailed()
	}

	parser := sitter.NewParser()
	defer parser.Close()
	if err := parser.SetLanguage(languageValue); err != nil {
		return nil, nil, nil, "", newParseFailed()
	}

	tree, timedOut := parseWithTimeout(parser, source)
	if timedOut {
		if tree != nil {
			tree.Close()
		}
		return nil, nil, nil, "", newParseTimedOut(config.Name)
	}
	if tree == nil {
		return nil, nil, nil, "", newParseFailed()
	}

	return config, tree, languageValue, config.grammarName(ext), nil
}

func parseWithTimeout(parser *sitter.Parser, source []byte) (*sitter.Tree, bool) {
	if parseTimeout <= 0 {
		return parser.Parse(source, nil), false
	}

	deadline := time.Now().Add(parseTimeout)
	timedOut := false
	tree := parser.ParseWithOptions(func(offset int, _ sitter.Point) []byte {
		if offset < 0 || offset >= len(source) {
			return nil
		}
		return source[offset:]
	}, nil, &sitter.ParseOptions{
		ProgressCallback: func(_ sitter.ParseState) bool {
			if time.Now().Before(deadline) {
				return false
			}
			timedOut = true
			return true
		},
	})
	return tree, timedOut
}

func compiledQuery(config *Config, languageValue *sitter.Language, grammarKey string) (*sitter.Query, error) {
	cacheKey := config.Name + ":" + grammarKey
	if cached, ok := queryCache.Load(cacheKey); ok {
		return cached.(*sitter.Query), nil
	}

	query, queryErr := sitter.NewQuery(languageValue, config.Query)
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
