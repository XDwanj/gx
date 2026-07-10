package language

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	gts "github.com/odvcencio/gotreesitter"
	"github.com/odvcencio/gotreesitter/grammars"

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
	loadLanguage  func(ext string) *gts.Language
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
		grammarName:  func(_ string) string { return languageNameRust },
		loadLanguage: func(_ string) *gts.Language { return grammars.RustLanguage() },
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
		loadLanguage: func(ext string) *gts.Language {
			if isTSXFamilyExtension(ext) {
				return grammars.TsxLanguage()
			}
			return grammars.TypescriptLanguage()
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
		grammarName:  func(_ string) string { return languageNamePython },
		loadLanguage: func(_ string) *gts.Language { return grammars.PythonLanguage() },
	},
	{
		Name:         languageNameProtobuf,
		Extensions:   []string{"proto"},
		Query:        protobufQuery,
		RefNodeTypes: []string{nodeKindIdentifier},
		grammarName:  func(_ string) string { return languageNameProtobuf },
		loadLanguage: func(_ string) *gts.Language { return grammars.ProtoLanguage() },
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
		grammarName:  func(_ string) string { return languageNameGo },
		loadLanguage: func(_ string) *gts.Language { return grammars.GoLanguage() },
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
		loadLanguage: func(_ string) *gts.Language { return grammars.CLanguage() },
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
		grammarName:  func(_ string) string { return languageNameCPP },
		loadLanguage: func(_ string) *gts.Language { return grammars.CppLanguage() },
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
		grammarName:  func(_ string) string { return languageNameJava },
		loadLanguage: func(_ string) *gts.Language { return grammars.JavaLanguage() },
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
		grammarName:  func(_ string) string { return languageNameKotlin },
		loadLanguage: func(_ string) *gts.Language { return grammars.KotlinLanguage() },
	},
	{
		Name:         languageNameRuby,
		Extensions:   []string{"rb"},
		Query:        rubyQuery,
		RefNodeTypes: []string{nodeKindIdentifier, "constant"},
		CallNodeTypes: []string{
			nodeKindCall,
		},
		grammarName:  func(_ string) string { return languageNameRuby },
		loadLanguage: func(_ string) *gts.Language { return grammars.RubyLanguage() },
	},
	{
		Name:         languageNameLua,
		Extensions:   []string{languageNameLua},
		Query:        luaQuery,
		RefNodeTypes: []string{nodeKindIdentifier},
		CallNodeTypes: []string{
			"function_call",
		},
		grammarName:  func(_ string) string { return languageNameLua },
		loadLanguage: func(_ string) *gts.Language { return grammars.LuaLanguage() },
	},
	{
		Name:         languageNameZig,
		Extensions:   []string{"zig"},
		Query:        zigQuery,
		SigDelimiter: '{',
		RefNodeTypes: []string{nodeKindIdentifier},
		grammarName:  func(_ string) string { return languageNameZig },
		loadLanguage: func(_ string) *gts.Language { return grammars.ZigLanguage() },
	},
	{
		Name:         languageNameBash,
		Extensions:   []string{"sh", languageNameBash},
		Query:        bashQuery,
		SigDelimiter: '{',
		RefNodeTypes: []string{"word"},
		grammarName:  func(_ string) string { return languageNameBash },
		loadLanguage: func(_ string) *gts.Language { return grammars.BashLanguage() },
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
		grammarName:  func(_ string) string { return languageNameSwift },
		loadLanguage: func(_ string) *gts.Language { return grammars.SwiftLanguage() },
	},
	{
		Name:         languageNameVue,
		Extensions:   []string{languageNameVue},
		Query:        vueQuery,
		RefNodeTypes: []string{"tag_name", "attribute_name"},
		grammarName:  func(_ string) string { return languageNameVue },
		loadLanguage: func(_ string) *gts.Language { return grammars.VueLanguage() },
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
	defer tree.Release()

	query, err := compiledQuery(config, languageValue, grammarKey)
	if err != nil {
		return nil, err
	}

	return extractSymbols(config, query, tree, languageValue, source), nil
}

func FindReferences(languageName string, source []byte, path string, name string) ([]Reference, error) {
	return FindReferencesForNames(languageName, source, path, []string{name})
}

func FindReferencesForNames(languageName string, source []byte, path string, names []string) ([]Reference, error) {
	config, tree, languageValue, _, err := parseSource(languageName, source, path)
	if err != nil {
		return nil, err
	}
	defer tree.Release()

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
	walkReferenceLeaves(config, languageValue, tree.RootNode(), func(node *gts.Node) {
		if _, ok := nameSet[node.Text(source)]; !ok {
			return
		}
		references = append(references, Reference{
			Line:       int(node.StartPoint().Row) + 1,
			ByteOffset: uint(node.StartByte()),
		})
	})

	return references, nil
}

func FindReferenceNames(languageName string, source []byte, path string) ([]string, error) {
	config, tree, languageValue, _, err := parseSource(languageName, source, path)
	if err != nil {
		return nil, err
	}
	defer tree.Release()

	names := make([]string, 0)
	seen := make(map[string]struct{})
	walkReferenceLeaves(config, languageValue, tree.RootNode(), func(node *gts.Node) {
		name := node.Text(source)
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

func walkReferenceLeaves(config *Config, languageValue *gts.Language, root *gts.Node, visit func(node *gts.Node)) {
	stack := []*gts.Node{root}
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if node.ChildCount() == 0 && containsString(config.RefNodeTypes, node.Type(languageValue)) {
			visit(node)
			continue
		}

		for childIndex := node.ChildCount() - 1; childIndex >= 0; childIndex-- {
			child := node.Child(childIndex)
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

func parseSource(languageName string, source []byte, path string) (*Config, *gts.Tree, *gts.Language, string, error) {
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

	parser := gts.NewParser(languageValue)
	if parseTimeout > 0 {
		parser.SetTimeoutMicros(parseTimeoutMicros(parseTimeout))
	}
	tree, err := parser.ParseStrict(source)
	if err != nil || tree == nil {
		if tree != nil {
			tree.Release()
		}
		var stoppedEarly *gts.ParseStoppedEarlyError
		if errors.As(err, &stoppedEarly) && stoppedEarly.Reason == gts.ParseStopTimeout {
			return nil, nil, nil, "", newParseTimedOut(config.Name)
		}
		return nil, nil, nil, "", newParseFailed()
	}

	return config, tree, languageValue, config.grammarName(ext), nil
}

func parseTimeoutMicros(timeout time.Duration) uint64 {
	return uint64((timeout + time.Microsecond - 1) / time.Microsecond)
}

func compiledQuery(config *Config, languageValue *gts.Language, grammarKey string) (*gts.Query, error) {
	cacheKey := config.Name + ":" + grammarKey
	if cached, ok := queryCache.Load(cacheKey); ok {
		return cached.(*gts.Query), nil
	}

	query, queryErr := gts.NewQuery(config.Query, languageValue)
	if queryErr != nil {
		return nil, queryErr
	}

	actual, _ := queryCache.LoadOrStore(cacheKey, query)
	return actual.(*gts.Query), nil
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
