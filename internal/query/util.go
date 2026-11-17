package query

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/XDwanj/gx/internal/index"
)

func formatLocation(path string, line int) string {
	if line <= 0 {
		return path
	}
	return fmt.Sprintf("%s:%d", path, line)
}

func humanizeRows(rows any) any {
	switch typed := rows.(type) {
	case []SymbolRow:
		human := make([]symbolTextRow, 0, len(typed))
		for _, row := range typed {
			human = append(human, symbolTextRow{
				File:      formatLocation(row.File, row.Line),
				Name:      row.Name,
				Kind:      row.Kind,
				Signature: row.Signature,
			})
		}
		return human
	case []ReferenceRow:
		human := make([]referenceTextRow, 0, len(typed))
		for _, row := range typed {
			human = append(human, referenceTextRow{
				File:    formatLocation(row.File, row.Line),
				Caller:  row.Caller,
				Context: row.Context,
			})
		}
		return human
	case []UniqueCallerRow:
		human := make([]uniqueCallerTextRow, 0, len(typed))
		for _, row := range typed {
			human = append(human, uniqueCallerTextRow{
				File:   formatLocation(row.File, row.Line),
				Caller: row.Caller,
			})
		}
		return human
	default:
		return rows
	}
}

func readBody(root string, file string, symbol index.Symbol) (string, int, bool) {
	source, err := os.ReadFile(filepath.Join(root, file))
	if err != nil {
		return "", 0, false
	}
	if int(symbol.ByteEnd) > len(source) {
		return "", 0, false
	}
	start := int(symbol.ByteStart)
	line := bytes.Count(source[:start], []byte{'\n'}) + 1
	return string(source[symbol.ByteStart:symbol.ByteEnd]), line, true
}

func findEnclosingSymbol(symbols []index.Symbol, byteOffset uint) string {
	result := ""
	bestWidth := ^uint(0)
	for _, symbol := range symbols {
		if symbol.ByteStart <= byteOffset && byteOffset < symbol.ByteEnd {
			width := symbol.ByteEnd - symbol.ByteStart
			if width < bestWidth {
				bestWidth = width
				result = symbol.Name
			}
		}
	}
	return result
}

func splitLines(source []byte) []string {
	text := strings.ReplaceAll(string(source), "\r\n", "\n")
	return strings.Split(text, "\n")
}

func prepareSymbols(symbols []index.Symbol) []index.Symbol {
	result := make([]index.Symbol, 0)
	for _, symbol := range symbols {
		if !symbol.IsTest {
			result = append(result, symbol)
		}
	}
	sort.Slice(result, func(left int, right int) bool {
		leftPriority := symbolPriority(result[left].Kind)
		rightPriority := symbolPriority(result[right].Kind)
		if leftPriority == rightPriority {
			return result[left].Name < result[right].Name
		}
		return leftPriority < rightPriority
	})
	return result
}

func symbolPriority(kind index.SymbolKind) int {
	switch kind {
	case index.SymbolKindStruct, index.SymbolKindEnum, index.SymbolKindInterface, index.SymbolKindClass:
		return symbolPriorityPrimary
	case index.SymbolKindFunc, index.SymbolKindConst, index.SymbolKindType, index.SymbolKindModule:
		return symbolPrioritySecondary
	default:
		return symbolPriorityFallback
	}
}

func isTestFile(path string) bool {
	for _, part := range strings.Split(displayPath(path), "/") {
		if part == "tests" || part == "test" || part == "__tests__" {
			return true
		}
	}
	name := filepath.Base(path)
	switch {
	case strings.HasSuffix(name, "_test.go"):
		return true
	case strings.HasSuffix(name, ".test.ts"), strings.HasSuffix(name, ".test.tsx"), strings.HasSuffix(name, ".test.js"), strings.HasSuffix(name, ".test.jsx"):
		return true
	case strings.HasSuffix(name, ".spec.ts"), strings.HasSuffix(name, ".spec.tsx"), strings.HasSuffix(name, ".spec.js"), strings.HasSuffix(name, ".spec.jsx"):
		return true
	case strings.HasPrefix(name, "test_") && strings.HasSuffix(name, ".py"):
		return true
	case strings.HasSuffix(name, "_spec.rb"):
		return true
	default:
		return false
	}
}

func childComponent(path string, dir string) (string, bool) {
	relative := path
	if dir != "" {
		var err error
		relative, err = filepath.Rel(dir, path)
		if err != nil {
			return "", false
		}
	}
	parts := strings.Split(displayPath(relative), "/")
	if len(parts) == 0 {
		return "", false
	}
	if len(parts) > 1 {
		return parts[0], true
	}
	return relative, true
}

func formatSubdir(relDir string, dirName string) string {
	if relDir == "" {
		return displayPath(dirName) + "/"
	}
	return displayPath(filepath.Join(relDir, dirName)) + "/"
}

func sortedKeys(values map[string][2]int) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func minInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}
