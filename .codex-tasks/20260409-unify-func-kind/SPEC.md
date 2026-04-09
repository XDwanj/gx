# Task Specification

> Scope anchor for the task. Update only when goals or constraints change, and log the reason in PROGRESS.md.

## Task Shape

- **Shape**: `single-full`

## Goals

- 将公开的可调用符号 kind 从 `fn`/`method` 合并为单一的 `func`
- 保持跨语言查询体验稳定，避免用户因忘记区分 `fn` 和 `method` 而漏查
- 同步更新 CLI 帮助、README、embedded skill 和自动化测试

## Non-Goals

- 不修改 `references` 的接口
- 不重做索引或查询架构
- 不引入兼容别名或静默 fallback

## Constraints

- 公共命令名保持为 `gx`
- 变更属于公开 CLI 契约替换，不保留 `fn` 兼容路径
- 先读入口与现有测试，再改 Cobra wiring 相关文案和实现
- 保持 human-readable 与 `--json` 输出一致
- Go 验证命令使用内建超时：`go test ./... -timeout 60s`
- `golangci-lint run ./...` 必须通过

## Environment

- **Project root**: `/Users/xdwanj/Project/Rust/gx`
- **Language/runtime**: `Go`
- **Package manager**: `go`
- **Test framework**: `go test`
- **Build command**: `go build ./...`
- **Existing test count**: `待确认`

## Risk Assessment

- [x] External dependencies (APIs, services) — availability confirmed?
- [x] Breaking changes to existing code — impact assessed?
- [x] Large file generation — disk space sufficient?
- [x] Long-running tests — timeout configured?

## Deliverables

- 公开 kind 枚举与解析逻辑改为使用 `func`
- 查询/排序/帮助文案/文档与 fixture 同步更新
- 自动化验证通过

## Done-When

- [ ] `gx symbols/definition --kind func` 覆盖原 `fn` 与 `method` 查询场景
- [ ] CLI 帮助、README、embedded skill 与测试同步完成
- [ ] `go test ./... -timeout 60s` 与 `golangci-lint run ./...` 通过

## Final Validation Command

```bash
go test ./... -timeout 60s && golangci-lint run ./...
```

## Demo Flow (optional)

1. 运行 `gx symbols --kind func --name 'build*' .`
2. 确认原本 `fn` 与 `method` 的可调用符号都可命中
