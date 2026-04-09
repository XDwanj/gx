package language

import (
	"fmt"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

type Callee struct {
	Name       string
	Line       int
	ByteOffset uint
}

func SupportsCallees(languageName string) bool {
	config := configFor(languageName)
	return config != nil && len(config.CallNodeTypes) > 0
}

func FindCallees(languageName string, source []byte, path string, startByte uint, endByte uint) ([]Callee, error) {
	config, tree, _, err := parseSource(languageName, source, path)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	if len(config.CallNodeTypes) == 0 {
		return nil, fmt.Errorf("gx: callees not supported for language: %s", languageName)
	}

	callees := make([]Callee, 0)
	stack := []*sitter.Node{tree.RootNode()}
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if node == nil {
			continue
		}
		if node.EndByte() <= startByte || node.StartByte() >= endByte {
			continue
		}

		if node.StartByte() >= startByte && node.EndByte() <= endByte && containsString(config.CallNodeTypes, node.Kind()) {
			callee := extractCallTarget(node, source)
			if callee != "" {
				callees = append(callees, Callee{
					Name:       callee,
					Line:       int(node.StartPosition().Row) + 1,
					ByteOffset: node.StartByte(),
				})
			}
		}

		for childIndex := int(node.ChildCount()) - 1; childIndex >= 0; childIndex-- {
			child := node.Child(uint(childIndex))
			if child != nil {
				stack = append(stack, child)
			}
		}
	}

	return callees, nil
}

func extractCallTarget(node *sitter.Node, source []byte) string {
	if node == nil {
		return ""
	}

	switch node.Kind() {
	case "call_expression":
		if functionNode := node.ChildByFieldName("function"); functionNode != nil {
			return normalizeCallText(functionNode.Utf8Text(source))
		}
	case "call":
		if functionNode := node.ChildByFieldName("function"); functionNode != nil {
			return normalizeCallText(functionNode.Utf8Text(source))
		}
		methodNode := node.ChildByFieldName("method")
		receiverNode := node.ChildByFieldName("receiver")
		return joinCallParts(receiverNode, methodNode, source)
	case "method_invocation":
		nameNode := node.ChildByFieldName("name")
		objectNode := node.ChildByFieldName("object")
		return joinCallParts(objectNode, nameNode, source)
	case "function_call":
		if nameNode := node.ChildByFieldName("name"); nameNode != nil {
			return normalizeCallText(nameNode.Utf8Text(source))
		}
	}

	for childIndex := uint(0); childIndex < node.NamedChildCount(); childIndex++ {
		child := node.NamedChild(childIndex)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "arguments", "argument_list", "call_suffix", "template_string", "type_arguments":
			continue
		default:
			return normalizeCallText(child.Utf8Text(source))
		}
	}

	return ""
}

func joinCallParts(left *sitter.Node, right *sitter.Node, source []byte) string {
	if right == nil {
		return ""
	}
	if left == nil {
		return normalizeCallText(right.Utf8Text(source))
	}
	return normalizeCallText(left.Utf8Text(source) + "." + right.Utf8Text(source))
}

func normalizeCallText(text string) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return ""
	}
	return strings.Join(strings.Fields(trimmed), " ")
}
