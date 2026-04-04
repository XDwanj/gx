# Tree-sitter Capability Survey And Kind Architecture

## Goal

调查 `gx` 当前与拟保留语言的 Tree-sitter grammar 能力边界，梳理 `gx` 的 `kind`
设计原则、语言映射策略与重构建议，并将调查与结论记录到 `Architecture.md`。

## Scope

- 盘点 `internal/language/language.go` 中当前注册语言
- 明确 `internal/grammars/` 中本地 vendored grammar 与外部 bindings 的边界
- 重点调查拟保留语言的 Tree-sitter 节点能力与当前 `gx` query 覆盖差异
- 对 `gx kind` 集合做跨语言适配分析，评估保留、删除、合并与补齐策略
- 产出新的 `Architecture.md`

## Non-Goals

- 本任务不直接修改语言支持实现或删除语言
- 本任务不直接修改 `internal/language/queries.go`
- 本任务不直接修改 CLI 行为或 README
