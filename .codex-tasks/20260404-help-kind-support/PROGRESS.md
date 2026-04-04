# 进度记录

## 2026-04-04

- 已确认 `cmd/symbols.go` 与 `cmd/definition.go` 当前只有 `Short` 文案，没有 `Long` 或单独 help 模板。
- 已确认用户要求帮助文案中的追加内容使用列表形式，而不是表格。
- 已新增 [cmd/help_kind_support.go](/Users/xdwanj/Project/Rust/gx/cmd/help_kind_support.go)，集中维护：
  - 公开 `kind` 列表
  - 活跃语言的当前支持列表
  - 说明这是 `gx` 当前抽取覆盖，而不是 Tree-sitter 全能力
- 已将该帮助文本接入：
  - [cmd/symbols.go](/Users/xdwanj/Project/Rust/gx/cmd/symbols.go)
  - [cmd/definition.go](/Users/xdwanj/Project/Rust/gx/cmd/definition.go)
- 已新增 [cmd/help_test.go](/Users/xdwanj/Project/Rust/gx/cmd/help_test.go)，校验：
  - `gx symbols --help` 包含 kind 支持列表
  - `gx definition --help` 包含 kind 支持列表
  - 输出采用列表形式，而不是表格
- 已完成验证：
  - `go test ./... -timeout 60s`
  - `golangci-lint run ./...`
