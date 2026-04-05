# Task Specification

> Scope anchor for the task. Update only when goals or constraints change, and log the reason in PROGRESS.md.

## Task Shape

- **Shape**: `single-full`

## Goals

- 修复 `gx definition --max-lines` 截断时的人类可读提示。
- 让输出明确表达“内容未输出完”，并指向 `--max-lines` 的继续查看方式。
- 保持现有 JSON 输出结构不变。

## Non-Goals

- 不修改分页提示模型。
- 不新增 `column` 或其他定义输出字段。

## Constraints

- 只在 `internal/query/` 中实现业务逻辑，Cobra wiring 保持不变。
- 保持 Debug-First，不添加 silent fallback。
- 变更后必须通过 `go test ./... -timeout 60s` 与 `golangci-lint run ./...`。

## Deliverables

- `internal/query/runtime.go` 中新的截断提示文案。
- 已有 `internal/query/runtime_test.go` 新测试转绿。
- `.codex-tasks/20260405-max-lines-hint-fix/` 任务记录。

## Final Validation Command

```bash
go test ./... -timeout 60s && golangci-lint run ./...
```
