# 进度记录

## 2026-04-04

- 已确认仓库当前还没有接任何 Cobra completion。
- 已确认可以通过 `RegisterFlagCompletionFunc()` 注册 `--kind` 补全，并通过 `GetFlagCompletionFunc()` 做命令级测试。
- 已将公开 `kind` 集合收敛到 [internal/index/types.go](/Users/xdwanj/Project/Rust/gx/internal/index/types.go) 的 `PublicSymbolKinds()`，避免 completion、help 文案和 kind 解析漂移。
- 已新增 [cmd/completion_kind.go](/Users/xdwanj/Project/Rust/gx/cmd/completion_kind.go)，为 `symbols` 与 `definition` 的 `--kind` 注册 Cobra flag completion，候选即全部公开 `kind`，并使用 `ShellCompDirectiveNoFileComp` 关闭文件补全。
- 已补测试：
  - [cmd/completion_test.go](/Users/xdwanj/Project/Rust/gx/cmd/completion_test.go)
  - [internal/index/types_test.go](/Users/xdwanj/Project/Rust/gx/internal/index/types_test.go)
- 已完成验证：
  - `go test ./cmd -run 'Help|Completion' -timeout 60s`
  - `go test ./... -timeout 60s`
  - `golangci-lint run ./...`
