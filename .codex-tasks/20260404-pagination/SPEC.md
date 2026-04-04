# Task Specification

> Scope anchor for the task. Update only when goals or constraints change, and log the reason in PROGRESS.md.

## Task Shape

- **Shape**: `single-full`

## Goals

- 为 `gx definition`、`gx symbols`、`gx references`、`gx overview` 增加统一分页能力。
- 提供新的全局 flags：`--limit N`、`--offset N`、`--all`。
- 默认分页值固定为 `definition=5`、`symbols=100`、`references=50`、`dir_overview=none`。
- 当结果被截断时，在 `stderr` 输出紧凑提示，引导用户收窄查询或继续翻页。

## Non-Goals

- 不修改 `definition` 的 `--max-lines` 语义。
- 不为 `overview` 的文件模式额外引入分页。
- 不改变现有 `--json` 与人类可读输出的字段结构，只增加分页行为与提示。

## Constraints

- 遵守项目现有分层：Cobra wiring 在 `cmd/`，业务逻辑在 `internal/`。
- 保持 Debug-First，不添加 silent fallback。
- 所有命令文档、帮助和示例继续使用 `gx` 作为公共命令名。
- 变更后必须通过 `go test ./... -timeout 60s` 与 `golangci-lint run ./...`。

## Environment

- **Project root**: `/Users/xdwanj/Project/Rust/gx`
- **Language/runtime**: `Go`
- **Package manager**: `Go modules`
- **Test framework**: `go test`
- **Build command**: `go build ./...`
- **Existing test count**: `70`

## Risk Assessment

- [x] External dependencies (APIs, services) — availability confirmed?
- [x] Breaking changes to existing code — impact assessed?
- [x] Large file generation — disk space sufficient?
- [x] Long-running tests — timeout configured?

## Deliverables

- 根命令上的分页 flags 与参数解析。
- 查询层统一分页模型与截断 hint。
- `symbols / definition / references / overview` 的分页接入。
- 对应命令/查询层测试与帮助文本更新。

## Done-When

- [ ] 四个命令按约定支持分页，默认值正确，`overview` 仅目录模式分页。
- [ ] 截断时 stderr 会输出紧凑提示，未截断时不产生额外噪音。
- [ ] `--limit`、`--offset`、`--all`、默认值和 JSON/TOON 输出都被测试覆盖。
- [ ] 全量测试和 lint 通过。

## Final Validation Command

```bash
go test ./... -timeout 60s && golangci-lint run ./...
```

## Demo Flow (optional)

1. `gx symbols --name 'build*' .`
2. `gx symbols --name 'build*' --offset 100 .`
3. `gx definition --name 'build*' --all .`
4. `gx overview internal/query`
