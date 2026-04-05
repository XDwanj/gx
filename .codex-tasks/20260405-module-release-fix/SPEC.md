# Task Specification

> Scope anchor for the task. Update only when goals or constraints change, and log the reason in PROGRESS.md.

## Task Shape

- **Shape**: `single-full`

## Goals

- 将模块路径从 `gx` 改为可公开安装的 `github.com/XDwanj/gx`。
- 删除阻塞 `go install module@version` 的 `replace` 指令。
- 保持 `gx` CLI 行为不变，并让仓库重新满足发布安装条件。

## Non-Goals

- 不改动 `gx` 命令的用户可见语义。
- 不重写 tree-sitter grammar 生成代码或更新 vendored grammar 内容。
- 不在本任务内发布 tag。

## Constraints

- 先读入口与测试，再改动模块路径和依赖。
- Cobra wiring 继续留在 `cmd/`，业务逻辑不额外搬迁。
- 保持 Debug-First，不增加 silent fallback 或假成功路径。
- 变更后必须运行 `go mod tidy`、`go test ./... -timeout 60s`、`golangci-lint run ./...`。

## Deliverables

- `go.mod` 使用完整模块路径且不再包含 `replace`。
- 仓库内 Go import 与版本注入路径同步更新。
- 任务记录文件 `.codex-tasks/20260405-module-release-fix/`。

## Done-When

- [ ] `go install github.com/XDwanj/gx@<new-version>` 不再被当前模块配置阻塞。
- [ ] 仓库通过 `go mod tidy`、全量测试和 lint。

## Final Validation Command

```bash
go mod tidy && go test ./... -timeout 60s && golangci-lint run ./...
```
