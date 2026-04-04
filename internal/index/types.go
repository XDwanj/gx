package index

import (
	"fmt"
	"strings"
)

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

var validSymbolKinds = map[string]SymbolKind{
	"fn":        SymbolKindFn,
	"struct":    SymbolKindStruct,
	"enum":      SymbolKindEnum,
	"type":      SymbolKindType,
	"const":     SymbolKindConst,
	"class":     SymbolKindClass,
	"interface": SymbolKindInterface,
	"method":    SymbolKindMethod,
	"module":    SymbolKindModule,
}

func PublicSymbolKinds() []SymbolKind {
	return []SymbolKind{
		SymbolKindFn,
		SymbolKindMethod,
		SymbolKindConst,
		SymbolKindStruct,
		SymbolKindEnum,
		SymbolKindClass,
		SymbolKindInterface,
		SymbolKindModule,
		SymbolKindType,
	}
}

func ParseSymbolKind(raw string) (SymbolKind, error) {
	kind, ok := validSymbolKinds[strings.ToLower(strings.TrimSpace(raw))]
	if !ok {
		return "", fmt.Errorf("gx: invalid symbol kind: %s", raw)
	}
	return kind, nil
}

type Symbol struct {
	Name      string     `json:"name"`
	Kind      SymbolKind `json:"kind"`
	Signature string     `json:"signature"`
	ByteStart uint       `json:"byte_start"`
	ByteEnd   uint       `json:"byte_end"`
	IsTest    bool       `json:"is_test,omitempty"`
}

type FileEntry struct {
	MTimeSecs  int64  `json:"mtime_secs"`
	MTimeNanos int64  `json:"mtime_nanos"`
	Language   string `json:"language"`
}

type FileData struct {
	Meta    FileEntry `json:"meta"`
	Symbols []Symbol  `json:"symbols"`
}
