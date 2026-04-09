# Task Specification

## Task Shape

- **Shape**: `single-full`

## Goals

- 调整 `gx symbols` 的路径校验行为，在用户传入多个不存在的文件或目录时，一次性汇总并显式报出全部无效路径。
- 保持现有 `-C` 路径解析语义不变，只改错误呈现与测试覆盖。

## Non-Goals

- 不修改 `-C` 的解析基准目录规则。
- 不引入兼容兜底、自动猜测文件扩展名或静默跳过无效路径。
- 不改动 `definition`、`references`、`overview` 的对外契约，除非测试证明共用逻辑一并受益且不破坏现有行为。

## Constraints

- 遵循仓库 Debug-First 原则，错误必须显式暴露。
- Cobra wiring 保持在 `cmd/`，业务校验放在 `internal/`。
- 变更后需补测试并通过 `go test ./... -timeout 60s` 与 `golangci-lint run ./...`。

## Environment

- **Project root**: `/home/xdwanj/Project/Golang/gx`
- **Language/runtime**: Go
- **Package manager**: Go modules
- **Test framework**: `go test`
- **Build command**: `go run .`
- **Existing test count**: 待验证

## Risk Assessment

- [x] Breaking changes to existing code — 仅调整错误汇总文案，需靠测试锁定范围
- [x] Long-running tests — 使用 `go test ./... -timeout 60s`

## Deliverables

- `internal/query/runtime.go` 中的路径不存在错误汇总实现
- 对应命令/查询层测试更新
- `.codex-tasks/20260409-symbols-missing-path-errors/PROGRESS.md` 记录执行过程

## Done-When

- [ ] `gx symbols` 在多个 path 不存在时返回一条包含全部缺失路径的显式错误
- [ ] 现有 `-C` 相对路径解析测试仍通过
- [ ] 全量测试与 lint 通过

## Final Validation Command

```bash
go test ./... -timeout 60s && golangci-lint run ./...
```
