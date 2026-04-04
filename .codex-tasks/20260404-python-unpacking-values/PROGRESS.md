# 进度记录

## 2026-04-04

- 已用最小样例确认 Python Tree-sitter 节点结构：
  - `X, Y = ...` -> `assignment left: pattern_list`
  - `(U, V) = ...` -> `assignment left: tuple_pattern`
  - `[P, Q] = ...` -> `assignment left: list_pattern`
- 已确认当前 [internal/language/queries.go](/Users/xdwanj/Project/Rust/gx/internal/language/queries.go) 只覆盖 `left: (identifier)`，因此 unpacking 左值当前不会进索引。
- 已扩展 [internal/language/queries.go](/Users/xdwanj/Project/Rust/gx/internal/language/queries.go) 的 `pythonQuery`，现在支持：
  - `left: (identifier)`
  - `left: (pattern_list ...)`
  - `left: (tuple_pattern ...)`
  - `left: (list_pattern ...)`
- 已新增 Python fixture：
  - [tests/python/symbol/unpacking_kind](/Users/xdwanj/Project/Rust/gx/tests/python/symbol/unpacking_kind)
  - [tests/python/definition/unpacking_kind](/Users/xdwanj/Project/Rust/gx/tests/python/definition/unpacking_kind)
- 真实命令验证：
  - `go run . --root /tmp/gx_python_unpack_fix s --kind const .`
  - `go run . --root /tmp/gx_python_unpack_fix d --name 'Q' --kind const .`
  都已正常返回 unpacking 左值中的名字和定义体。
- 已完成验证：
  - `go test ./cmd -run 'TestLanguageFixtures/python' -timeout 60s`
  - `go test ./... -timeout 60s`
  - `golangci-lint run ./...`
