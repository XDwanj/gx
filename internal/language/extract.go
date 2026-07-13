package language

import (
	"fmt"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

const (
	symbolKindPriorityUnknown   = 0
	symbolKindPriorityType      = 1
	symbolKindPriorityValue     = 2
	symbolKindPriorityCallable  = 3
	symbolKindPriorityContainer = 4
	vueBlockTemplate            = "template"
	vueBlockScript              = "script"
	vueBlockStyle               = "style"
)

type Reference struct {
	Line       int
	ByteOffset uint
}

func extractSymbols(config *Config, query *sitter.Query, tree *sitter.Tree, source []byte) []Symbol {
	cursor := sitter.NewQueryCursor()
	defer cursor.Close()

	matches := cursor.Matches(query, tree.RootNode(), source)
	captureNames := query.CaptureNames()
	symbols := make([]Symbol, 0)

	for match := matches.Next(); match != nil; match = matches.Next() {
		nameNodes := make([]sitter.Node, 0, 1)
		var definitionNode *sitter.Node
		definitionKind := ""

		for _, capture := range match.Captures {
			name := captureNames[capture.Index]
			node := capture.Node
			if name == "name" {
				nameNodes = append(nameNodes, node)
				continue
			}
			if strings.HasPrefix(name, "definition.") {
				definitionNode = &node
				definitionKind = name
			}
		}

		if len(nameNodes) == 0 || definitionNode == nil || definitionKind == "" {
			continue
		}

		kind, ok := resolveKind(config, definitionKind, definitionNode.Kind())
		if !ok {
			continue
		}

		signature := buildSignature(config, definitionNode, source)
		isTest := detectTestSymbol(config.Name, definitionNode, source)
		position := definitionNode.StartPosition()
		line := int(position.Row) + 1
		for _, nameNode := range nameNodes {
			name := nameNode.Utf8Text(source)
			if config.Name == languageNameVue && !isVueTopLevelBlock(definitionNode, name) {
				continue
			}
			symbols = append(symbols, Symbol{
				Name:      name,
				Kind:      kind,
				Signature: signature,
				Line:      line,
				ByteStart: definitionNode.StartByte(),
				ByteEnd:   definitionNode.EndByte(),
				IsTest:    isTest,
			})
		}
	}

	return deduplicate(symbols)
}

func isVueTopLevelBlock(node *sitter.Node, name string) bool {
	parent := node.Parent()
	if parent == nil || parent.Kind() != "document" {
		return false
	}
	return name == vueBlockTemplate || name == vueBlockScript || name == vueBlockStyle
}

func resolveKind(config *Config, captureName string, nodeKind string) (SymbolKind, bool) {
	for _, override := range config.KindOverrides {
		if override.CaptureName == captureName && override.NodeKind == nodeKind {
			return override.Kind, true
		}
	}
	for _, override := range config.KindOverrides {
		if override.CaptureName == captureName && override.NodeKind == "" {
			return override.Kind, true
		}
	}

	switch captureName {
	case "definition.function", "definition.macro", "definition.method":
		return SymbolKindFunc, true
	case "definition.struct":
		return SymbolKindStruct, true
	case "definition.class":
		return SymbolKindClass, true
	case "definition.interface":
		return SymbolKindInterface, true
	case "definition.type":
		return SymbolKindType, true
	case "definition.enum":
		return SymbolKindEnum, true
	case "definition.module":
		return SymbolKindModule, true
	case "definition.constant":
		return SymbolKindConst, true
	default:
		return "", false
	}
}

func buildSignature(config *Config, node *sitter.Node, source []byte) string {
	start := node.StartByte()
	end := node.EndByte()
	if int(end) > len(source) {
		end = uint(len(source))
	}
	text := source[start:end]

	if config.SigBodyChild != "" {
		cursor := node.Walk()
		defer cursor.Close()
		for _, child := range node.Children(cursor) {
			if child.Kind() == config.SigBodyChild {
				if int(child.StartByte()) <= len(source) {
					signature := strings.TrimSpace(string(source[start:child.StartByte()]))
					signature = strings.TrimSpace(strings.TrimSuffix(signature, ":"))
					if signature != "" {
						return signature
					}
				}
			}
		}
	}

	if config.SigDelimiter != 0 {
		if indexValue := strings.IndexByte(string(text), config.SigDelimiter); indexValue >= 0 {
			signature := strings.TrimSpace(string(text[:indexValue]))
			if signature != "" {
				return signature
			}
		}
	}

	if indexValue := strings.Index(string(text), "=>"); indexValue >= 0 {
		signature := strings.TrimSpace(string(text[:indexValue+2]))
		if signature != "" {
			return signature
		}
	}

	firstLine := string(text)
	if newLineIndex := strings.IndexByte(firstLine, '\n'); newLineIndex >= 0 {
		firstLine = firstLine[:newLineIndex]
	}

	firstLine = strings.TrimSpace(firstLine)
	firstLine = strings.TrimSpace(strings.TrimSuffix(firstLine, "{"))
	firstLine = strings.TrimSpace(strings.TrimSuffix(firstLine, ":"))
	return firstLine
}

func deduplicate(symbols []Symbol) []Symbol {
	seen := map[string]int{}
	result := make([]Symbol, 0, len(symbols))
	for _, symbol := range symbols {
		key := fmt.Sprintf("%s:%d:%d", symbol.Name, symbol.ByteStart, symbol.ByteEnd)
		if existingIndex, ok := seen[key]; ok {
			if preferredSymbolKind(symbol.Kind, result[existingIndex].Kind) {
				result[existingIndex] = symbol
			}
			continue
		}
		seen[key] = len(result)
		result = append(result, symbol)
	}
	return result
}

func preferredSymbolKind(candidate SymbolKind, current SymbolKind) bool {
	return symbolKindSpecificity(candidate) > symbolKindSpecificity(current)
}

func symbolKindSpecificity(kind SymbolKind) int {
	switch kind {
	case SymbolKindStruct, SymbolKindEnum, SymbolKindClass, SymbolKindInterface:
		return symbolKindPriorityContainer
	case SymbolKindFunc:
		return symbolKindPriorityCallable
	case SymbolKindConst, SymbolKindModule:
		return symbolKindPriorityValue
	case SymbolKindType:
		return symbolKindPriorityType
	default:
		return symbolKindPriorityUnknown
	}
}

func detectTestSymbol(languageName string, node *sitter.Node, source []byte) bool {
	if languageName != languageNameRust {
		return false
	}
	return hasRustTestAttribute(node, source)
}

func hasRustTestAttribute(node *sitter.Node, source []byte) bool {
	for sibling := node.PrevNamedSibling(); sibling != nil; sibling = sibling.PrevNamedSibling() {
		if sibling.Kind() != "attribute_item" {
			break
		}
		if isTestAttribute(sibling.Utf8Text(source)) {
			return true
		}
	}

	for parent := node.Parent(); parent != nil; parent = parent.Parent() {
		if parent.Kind() != "mod_item" {
			continue
		}
		for sibling := parent.PrevNamedSibling(); sibling != nil; sibling = sibling.PrevNamedSibling() {
			if sibling.Kind() != "attribute_item" {
				break
			}
			if strings.Contains(sibling.Utf8Text(source), "cfg(test)") {
				return true
			}
		}
	}
	return false
}

func isTestAttribute(text string) bool {
	trimmed := strings.TrimPrefix(text, "#[")
	trimmed = strings.TrimSuffix(trimmed, "]")
	if trimmed == "test" {
		return true
	}
	if strings.HasSuffix(trimmed, "::test") || strings.Contains(trimmed, "::test(") {
		return true
	}
	return strings.Contains(trimmed, "cfg(test)")
}
