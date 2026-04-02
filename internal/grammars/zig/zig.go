package tree_sitter_zig

// #cgo CPPFLAGS: -I.
// #cgo CFLAGS: -std=c11 -fPIC
// #include "./tree_sitter/parser.h"
// const TSLanguage *tree_sitter_zig(void);
import "C"

import "unsafe"

func Language() unsafe.Pointer {
	return unsafe.Pointer(C.tree_sitter_zig())
}
