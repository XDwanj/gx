package language

import (
	"fmt"
	"strings"

	gts "github.com/odvcencio/gotreesitter"
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
	config, tree, languageValue, _, err := parseSource(languageName, source, path)
	if err != nil {
		return nil, err
	}
	defer tree.Release()

	if len(config.CallNodeTypes) == 0 {
		return nil, fmt.Errorf("gx: callees not supported for language: %s", languageName)
	}

	startByte32 := uint32(startByte)
	endByte32 := uint32(endByte)
	callees := make([]Callee, 0)
	stack := []*gts.Node{tree.RootNode()}
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if node == nil {
			continue
		}
		if node.EndByte() <= startByte32 || node.StartByte() >= endByte32 {
			continue
		}

		if node.StartByte() >= startByte32 && node.EndByte() <= endByte32 && containsString(config.CallNodeTypes, node.Type(languageValue)) {
			callee := extractCallTarget(languageValue, node, source)
			if callee != "" {
				callees = append(callees, Callee{
					Name:       callee,
					Line:       int(node.StartPoint().Row) + 1,
					ByteOffset: uint(node.StartByte()),
				})
			}
		}

		for childIndex := node.ChildCount() - 1; childIndex >= 0; childIndex-- {
			child := node.Child(childIndex)
			if child != nil {
				stack = append(stack, child)
			}
		}
	}

	return callees, nil
}

func extractCallTarget(languageValue *gts.Language, node *gts.Node, source []byte) string {
	if node == nil {
		return ""
	}

	switch node.Type(languageValue) {
	case nodeKindCallExpression:
		if functionNode := node.ChildByFieldName("function", languageValue); functionNode != nil {
			return normalizeCallText(functionNode.Text(source))
		}
	case nodeKindCall:
		if functionNode := node.ChildByFieldName("function", languageValue); functionNode != nil {
			return normalizeCallText(functionNode.Text(source))
		}
		methodNode := node.ChildByFieldName("method", languageValue)
		receiverNode := node.ChildByFieldName("receiver", languageValue)
		return joinCallParts(receiverNode, methodNode, source)
	case "method_invocation":
		nameNode := node.ChildByFieldName("name", languageValue)
		objectNode := node.ChildByFieldName("object", languageValue)
		return joinCallParts(objectNode, nameNode, source)
	case "function_call":
		if nameNode := node.ChildByFieldName("name", languageValue); nameNode != nil {
			return normalizeCallText(nameNode.Text(source))
		}
	}

	for childIndex := 0; childIndex < node.NamedChildCount(); childIndex++ {
		child := node.NamedChild(childIndex)
		if child == nil {
			continue
		}
		switch child.Type(languageValue) {
		case "arguments", "argument_list", "call_suffix", "template_string", "type_arguments":
			continue
		default:
			return normalizeCallText(child.Text(source))
		}
	}

	return ""
}

func joinCallParts(left *gts.Node, right *gts.Node, source []byte) string {
	if right == nil {
		return ""
	}
	if left == nil {
		return normalizeCallText(right.Text(source))
	}
	return normalizeCallText(left.Text(source) + "." + right.Text(source))
}

func normalizeCallText(text string) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return ""
	}
	return strings.Join(strings.Fields(trimmed), " ")
}
