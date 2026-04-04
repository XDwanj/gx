# 目标

修复 Go 语言在 `gx` 中对 grouped `const (...)` 和 grouped `var (...)` 声明的符号抽取问题，使这些包级值声明也能被索引为 `const`。

# 范围

- 调整 Go 的 Tree-sitter query，覆盖 grouped 与 non-grouped 的 `const` / `var`。
- 维持当前设计不变：
  - Go 包级 `const` -> `const`
  - Go 包级 `var` -> `const`
- 新增 `tests/go` fixture，覆盖 grouped `const`、enum-like const block 和 grouped `var`。

# 验证

- `go test ./cmd -run 'TestLanguageFixtures/go' -timeout 60s`
- `go test ./... -timeout 60s`
- `golangci-lint run ./...`
