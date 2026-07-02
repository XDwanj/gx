package language

import (
	"fmt"
	"strings"

	gts "github.com/odvcencio/gotreesitter"
)

const (
	symbolKindPriorityUnknown   = 0
	symbolKindPriorityType      = 1
	symbolKindPriorityValue     = 2
	symbolKindPriorityCallable  = 3
	symbolKindPriorityContainer = 4
	definitionKindFunction      = "definition.function"
)

type Reference struct {
	Line       int
	ByteOffset uint
}

func extractSymbols(config *Config, query *gts.Query, tree *gts.Tree, languageValue *gts.Language, source []byte) []Symbol {
	cursor := query.Exec(tree.RootNode(), languageValue, source)
	symbols := make([]Symbol, 0)

	for match, ok := cursor.NextMatch(); ok; match, ok = cursor.NextMatch() {
		nameNodes := make([]*gts.Node, 0, 1)
		var definitionNode *gts.Node
		definitionKind := ""

		for _, capture := range match.Captures {
			if capture.Name == "name" {
				nameNodes = append(nameNodes, capture.Node)
				continue
			}
			if strings.HasPrefix(capture.Name, "definition.") {
				definitionNode = capture.Node
				definitionKind = capture.Name
			}
		}

		if len(nameNodes) == 0 || definitionNode == nil || definitionKind == "" {
			continue
		}

		kind, ok := resolveKind(config, definitionKind, definitionNode.Type(languageValue))
		if !ok {
			continue
		}

		byteStart := symbolStartByte(config, languageValue, definitionKind, definitionNode)
		signature := buildSignature(config, languageValue, definitionNode, source, byteStart)
		isTest := detectTestSymbol(config.Name, languageValue, definitionNode, source)
		position := definitionNode.StartPoint()
		line := int(position.Row) + 1
		for _, nameNode := range nameNodes {
			symbols = append(symbols, Symbol{
				Name:      nameNode.Text(source),
				Kind:      kind,
				Signature: signature,
				Line:      line,
				ByteStart: uint(byteStart),
				ByteEnd:   uint(definitionNode.EndByte()),
				IsTest:    isTest,
			})
		}
	}

	return deduplicate(symbols)
}

func symbolStartByte(config *Config, languageValue *gts.Language, definitionKind string, node *gts.Node) uint32 {
	if config.Name != languageNameZig || definitionKind != definitionKindFunction {
		return node.StartByte()
	}
	for _, child := range node.Children() {
		if child.Type(languageValue) == "fn" {
			return child.StartByte()
		}
	}
	return node.StartByte()
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
	case definitionKindFunction, "definition.macro", "definition.method":
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

func buildSignature(config *Config, languageValue *gts.Language, node *gts.Node, source []byte, start uint32) string {
	end := node.EndByte()
	if int(end) > len(source) {
		end = uint32(len(source))
	}
	text := source[start:end]

	if config.SigBodyChild != "" {
		for _, child := range node.Children() {
			if child.Type(languageValue) == config.SigBodyChild {
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

func detectTestSymbol(languageName string, languageValue *gts.Language, node *gts.Node, source []byte) bool {
	if languageName != languageNameRust {
		return false
	}
	return hasRustTestAttribute(languageValue, node, source)
}

func hasRustTestAttribute(languageValue *gts.Language, node *gts.Node, source []byte) bool {
	for sibling := previousNamedSibling(node); sibling != nil; sibling = previousNamedSibling(sibling) {
		if sibling.Type(languageValue) != "attribute_item" {
			break
		}
		if isTestAttribute(sibling.Text(source)) {
			return true
		}
	}

	for parent := node.Parent(); parent != nil; parent = parent.Parent() {
		if parent.Type(languageValue) != "mod_item" {
			continue
		}
		for sibling := previousNamedSibling(parent); sibling != nil; sibling = previousNamedSibling(sibling) {
			if sibling.Type(languageValue) != "attribute_item" {
				break
			}
			if strings.Contains(sibling.Text(source), "cfg(test)") {
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

func previousNamedSibling(node *gts.Node) *gts.Node {
	for sibling := node.PrevSibling(); sibling != nil; sibling = sibling.PrevSibling() {
		if sibling.IsNamed() {
			return sibling
		}
	}
	return nil
}
