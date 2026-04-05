# Task Specification

> Scope anchor for the task. Update only when goals or constraints change, and log the reason in PROGRESS.md.

## Task Shape

- **Shape**: `single-full`

## Goals

- 补上 `definition --max-lines` 截断提示测试，明确要求用户能看出“内容未输出完”。
- 补上 `references` 分页提示测试，覆盖结果被分页截断时的 stderr hint。
- 用测试结果显式暴露当前实现是否满足上述契约。

## Non-Goals

- 本任务不顺带修改分页或截断提示实现。
- 不改变现有 `--json` 输出结构。

## Constraints

- 遵守项目现有分层，测试优先补在已有 `cmd/` 与 `internal/query/` 测试文件中。
- 保持 Debug-First；如果测试暴露缺口，应保留失败结果，不添加静默回退。
- 验证命令使用 Go 内建超时：`go test ./... -timeout 60s`。

## Environment

- **Project root**: `/Users/xdwanj/Project/Rust/gx`
- **Language/runtime**: `Go`
- **Test framework**: `go test`

## Deliverables

- `internal/query/runtime_test.go` 中新增 `--max-lines` 截断提示测试。
- `internal/query/runtime_test.go` 中新增 `references` 分页提示测试。
- `.codex-tasks/20260405-pagination-test-gaps/` 任务记录与验证结果。

## Done-When

- [ ] 新增测试准确表达目标行为。
- [ ] 运行相关 `go test`，并记录通过/失败结果。
- [ ] 如果存在失败，失败信息能直接说明是实现缺口而非测试噪音。

## Final Validation Command

```bash
go test ./internal/query ./cmd -timeout 60s
```
