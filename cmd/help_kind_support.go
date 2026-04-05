package cmd

import (
	"strings"

	"github.com/XDwanj/gx/internal/index"
)

type languageKindSupport struct {
	Language string
	Kinds    []index.SymbolKind
}

func publicKinds() []index.SymbolKind {
	return index.PublicSymbolKinds()
}

func languageKindSupportMatrix() []languageKindSupport {
	return []languageKindSupport{
		{
			Language: "bash",
			Kinds: []index.SymbolKind{
				index.SymbolKindFn,
			},
		},
		{
			Language: "c",
			Kinds: []index.SymbolKind{
				index.SymbolKindFn,
				index.SymbolKindStruct,
				index.SymbolKindEnum,
				index.SymbolKindType,
			},
		},
		{
			Language: "cpp",
			Kinds: []index.SymbolKind{
				index.SymbolKindFn,
				index.SymbolKindMethod,
				index.SymbolKindStruct,
				index.SymbolKindClass,
				index.SymbolKindEnum,
				index.SymbolKindModule,
				index.SymbolKindType,
			},
		},
		{
			Language: "go",
			Kinds: []index.SymbolKind{
				index.SymbolKindFn,
				index.SymbolKindMethod,
				index.SymbolKindConst,
				index.SymbolKindStruct,
				index.SymbolKindInterface,
				index.SymbolKindType,
			},
		},
		{
			Language: "java",
			Kinds: []index.SymbolKind{
				index.SymbolKindClass,
				index.SymbolKindMethod,
				index.SymbolKindConst,
				index.SymbolKindEnum,
				index.SymbolKindInterface,
				index.SymbolKindModule,
			},
		},
		{
			Language: "lua",
			Kinds: []index.SymbolKind{
				index.SymbolKindFn,
				index.SymbolKindMethod,
			},
		},
		{
			Language: "python",
			Kinds: []index.SymbolKind{
				index.SymbolKindFn,
				index.SymbolKindConst,
				index.SymbolKindClass,
			},
		},
		{
			Language: "protobuf",
			Kinds: []index.SymbolKind{
				index.SymbolKindStruct,
				index.SymbolKindEnum,
				index.SymbolKindInterface,
				index.SymbolKindMethod,
			},
		},
		{
			Language: "ruby",
			Kinds: []index.SymbolKind{
				index.SymbolKindMethod,
				index.SymbolKindClass,
				index.SymbolKindModule,
			},
		},
		{
			Language: "rust",
			Kinds: []index.SymbolKind{
				index.SymbolKindFn,
				index.SymbolKindMethod,
				index.SymbolKindConst,
				index.SymbolKindStruct,
				index.SymbolKindEnum,
				index.SymbolKindInterface,
				index.SymbolKindModule,
				index.SymbolKindType,
			},
		},
		{
			Language: "swift",
			Kinds: []index.SymbolKind{
				index.SymbolKindFn,
				index.SymbolKindMethod,
				index.SymbolKindConst,
				index.SymbolKindStruct,
				index.SymbolKindEnum,
				index.SymbolKindClass,
				index.SymbolKindInterface,
				index.SymbolKindModule,
				index.SymbolKindType,
			},
		},
		{
			Language: "typescript",
			Kinds: []index.SymbolKind{
				index.SymbolKindFn,
				index.SymbolKindMethod,
				index.SymbolKindConst,
				index.SymbolKindClass,
				index.SymbolKindEnum,
				index.SymbolKindInterface,
				index.SymbolKindModule,
				index.SymbolKindType,
			},
		},
		{
			Language: "zig",
			Kinds: []index.SymbolKind{
				index.SymbolKindFn,
				index.SymbolKindStruct,
				index.SymbolKindEnum,
			},
		},
	}
}

func formatKinds(kinds []index.SymbolKind) string {
	names := make([]string, 0, len(kinds))
	for _, kind := range kinds {
		names = append(names, string(kind))
	}
	return strings.Join(names, ", ")
}

func symbolsLongDescription() string {
	return strings.Join([]string{
		"Search symbols across project.",
		"",
		kindSupportHelpText(),
	}, "\n")
}

func definitionLongDescription() string {
	return strings.Join([]string{
		"Get a function or type body without reading the whole file.",
		"",
		kindSupportHelpText(),
	}, "\n")
}

func kindSupportHelpText() string {
	supportMatrix := languageKindSupportMatrix()
	lines := make([]string, 0, len(supportMatrix)+6)
	lines = append(lines,
		"Public kinds:",
		"- "+formatKinds(publicKinds()),
		"",
		"Language support summary:",
	)

	for _, support := range supportMatrix {
		lines = append(lines, "- "+support.Language+": "+formatKinds(support.Kinds))
	}

	lines = append(lines,
		"",
		"These lists describe current gx extraction coverage, not every syntax form that Tree-sitter could theoretically expose.",
	)

	return strings.Join(lines, "\n")
}
