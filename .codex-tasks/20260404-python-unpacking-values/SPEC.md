# 目标

修复 Python 模块级 unpacking 赋值在 `gx` 中的符号抽取问题，使 tuple/list unpacking 左值中的名字也能按 `const` 被索引和查询。

# 范围

- 扩展 Python 的 Tree-sitter query，覆盖：
  - `pattern_list`
  - `tuple_pattern`
  - `list_pattern`
- 保持当前设计不变：
  - 模块级赋值 -> `const`
- 新增 `tests/python` fixture，覆盖 unpacking `symbols` 与 `definition`。

# 验证

- `go test ./cmd -run 'TestLanguageFixtures/python' -timeout 60s`
- `go test ./... -timeout 60s`
- `golangci-lint run ./...`
