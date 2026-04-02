package tree_sitter_typescript

// #cgo CPPFLAGS: -I./tsx/src -I./typescript/src
// #cgo CFLAGS: -std=c11 -fPIC
// #include "./tsx/src/parser.c"
// #include "./tsx/src/scanner.c"
import "C"

import "unsafe"

func LanguageTSX() unsafe.Pointer {
	return unsafe.Pointer(C.tree_sitter_tsx())
}
