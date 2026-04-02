package tree_sitter_typescript

// #cgo CPPFLAGS: -I./typescript/src -I./tsx/src
// #cgo CFLAGS: -std=c11 -fPIC
// #include "./typescript/src/parser.c"
// #include "./typescript/src/scanner.c"
import "C"

import "unsafe"

func LanguageTypescript() unsafe.Pointer {
	return unsafe.Pointer(C.tree_sitter_typescript())
}
