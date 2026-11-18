package language

import (
	"regexp"
	"sort"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"

	langpkg "github.com/XDwanj/gx/internal/lang"
)

// embeddedSpec 描述标记语言文件里可二次解析的嵌入代码：
// collect 在宿言语法树上圈出字节区间，交给 language 指定的 grammar
// 用 Parser.SetIncludedRanges 再解析一遍，产出的节点偏移即原文件偏移。
type embeddedSpec struct {
	language string
	collect  func(root *sitter.Node, source []byte) []sitter.Range
}

var mustachePattern = regexp.MustCompile(`\{\{([\s\S]*?)\}\}`)

// eachEmbeddedTree 对宿主树圈出的每个区间单独做一次嵌入解析，
// 并回调产出的语法树。
//
// 区间必须逐个解析而不能合并传给 SetIncludedRanges：tree-sitter 会把
// 多个 included range 拼接成一个文档，跨区间的文本会被当成相邻 token
// （`{{a}}{{b}}` 变成 `a}}...{{b` 的粘连标识符）。
//
// 宿主语言没有嵌入声明、嵌入 grammar 未启用或区间为空时静默跳过，
// 调用方退化为只用宿主树的结果。
func eachEmbeddedTree(config *Config, source []byte, host *sitter.Tree, visit func(tree *sitter.Tree, config *Config, language *sitter.Language)) {
	spec := config.Embedded
	if spec == nil {
		return
	}
	embeddedConfig := configFor(spec.language)
	if embeddedConfig == nil || !langpkg.IsInstalled(embeddedConfig.Name) {
		return
	}
	ranges := spec.collect(host.RootNode(), source)
	if len(ranges) == 0 {
		return
	}

	languageValue := embeddedConfig.loadLanguage("")
	if languageValue == nil {
		return
	}
	parser := sitter.NewParser()
	defer parser.Close()
	if err := parser.SetLanguage(languageValue); err != nil {
		return
	}
	for _, item := range ranges {
		if err := parser.SetIncludedRanges([]sitter.Range{item}); err != nil {
			continue
		}
		tree, timedOut := parseWithTimeout(parser, source)
		if timedOut || tree == nil {
			if tree != nil {
				tree.Close()
			}
			continue
		}
		visit(tree, embeddedConfig, languageValue)
		tree.Close()
	}
}

// collectExpressionRanges 遍历 HTML 语法树，圈出三类嵌入区间：
// <script> 的 raw_text、谓词命中的属性值（Vue 的 @/: /v- 绑定、
// WXML 的 bind/catch 事件绑定）、以及文本和剩余属性值里的 {{ }} 表达式。
func collectExpressionRanges(root *sitter.Node, source []byte, isExprAttr func(string) bool) []sitter.Range {
	ranges := make([]sitter.Range, 0, 4)
	stack := []*sitter.Node{root}
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		switch node.Kind() {
		case "raw_text":
			if parent := node.Parent(); parent != nil && parent.Kind() == "script_element" {
				ranges = append(ranges, node.Range())
			}
		case "attribute":
			name, value := attributeNameAndValue(node, source)
			switch {
			case value == nil:
			case isExprAttr(name):
				ranges = append(ranges, value.Range())
			default:
				ranges = appendMustacheRanges(ranges, value, source)
			}
		case "text":
			ranges = appendMustacheRanges(ranges, node, source)
		}

		for index := int(node.ChildCount()) - 1; index >= 0; index-- {
			if child := node.Child(uint(index)); child != nil {
				stack = append(stack, child)
			}
		}
	}

	sort.Slice(ranges, func(i, j int) bool { return ranges[i].StartByte < ranges[j].StartByte })
	return ranges
}

func vueExpressionRanges(root *sitter.Node, source []byte) []sitter.Range {
	return collectExpressionRanges(root, source, func(name string) bool {
		return strings.HasPrefix(name, "@") ||
			strings.HasPrefix(name, ":") ||
			strings.HasPrefix(name, "v-")
	})
}

func wxmlExpressionRanges(root *sitter.Node, source []byte) []sitter.Range {
	return collectExpressionRanges(root, source, func(name string) bool {
		return strings.HasPrefix(name, "bind") ||
			strings.HasPrefix(name, "catch") ||
			strings.HasPrefix(name, "mut-bind") ||
			strings.HasPrefix(name, "capture-bind") ||
			strings.HasPrefix(name, "capture-catch")
	})
}

func attributeNameAndValue(attr *sitter.Node, source []byte) (string, *sitter.Node) {
	var name string
	var value *sitter.Node
	for index := uint(0); index < attr.ChildCount(); index++ {
		child := attr.Child(index)
		switch child.Kind() {
		case "attribute_name":
			name = child.Utf8Text(source)
		case "attribute_value":
			value = child
		case "quoted_attribute_value":
			for inner := uint(0); inner < child.ChildCount(); inner++ {
				if grandchild := child.Child(inner); grandchild.Kind() == "attribute_value" {
					value = grandchild
				}
			}
		}
	}
	return name, value
}

func appendMustacheRanges(ranges []sitter.Range, node *sitter.Node, source []byte) []sitter.Range {
	base := node.Range()
	text := source[base.StartByte:base.EndByte]
	for _, match := range mustachePattern.FindAllSubmatchIndex(text, -1) {
		innerStart, innerEnd := match[2], match[3]
		if innerStart < 0 || innerStart == innerEnd {
			continue
		}
		ranges = append(ranges, sitter.Range{
			StartByte:  base.StartByte + uint(innerStart),
			EndByte:    base.StartByte + uint(innerEnd),
			StartPoint: pointWithin(base.StartPoint, text, innerStart),
			EndPoint:   pointWithin(base.StartPoint, text, innerEnd),
		})
	}
	return ranges
}

// pointWithin 以 base（text 起点的 Point）为基准，推出 text 内 offset 处的 Point。
// tree-sitter 的 Point.Column 按字节计。
func pointWithin(base sitter.Point, text []byte, offset int) sitter.Point {
	point := base
	for _, b := range text[:offset] {
		if b == '\n' {
			point.Row++
			point.Column = 0
		} else {
			point.Column++
		}
	}
	return point
}
