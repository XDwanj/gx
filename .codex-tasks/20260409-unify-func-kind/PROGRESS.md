# Progress Log

> Auto-maintained by Taskmaster. Each entry records what happened, why, and what's next.
> This file serves as both decision audit trail and context-recovery anchor.

---

## Session Start

- **Date**: 2026-04-09
- **Task name**: `20260409-unify-func-kind`
- **Task dir**: `.codex-tasks/20260409-unify-func-kind/`
- **Spec**: See SPEC.md
- **Plan**: See TODO.csv (4 milestones)
- **Environment**: Go / Cobra CLI / go test + golangci-lint

---

## Context Recovery Block

- **Current milestone**: #4 — 运行验证命令并收尾
- **Current status**: IN_PROGRESS
- **Last completed**: #3 — 更新测试与 fixture
- **Current artifact**: `.codex-tasks/20260409-unify-func-kind/TODO.csv`
- **Key context**: 公开 kind 已统一为 `func`，并已同步 CLI help、README、embedded skill 和 fixtures。`go test ./... -timeout 60s` 与 `golangci-lint run ./...` 均通过。
- **Known issues**: `fn_kind` / `method_kind` 目录名仍保留作为历史 fixture 命名，但其公开查询值和期望输出已切到 `func`。
- **Next action**: 汇总变更并结束任务。

---

## Milestone 1: 阅读入口与现有测试并确认变更面

- **Status**: DONE
- **Started**: 00:00
- **Completed**: 00:00
- **What was done**:
  - 阅读了 `cmd/symbols.go`、`cmd/definition.go`、`cmd/help_kind_support.go`
  - 阅读了 `internal/index/types.go`、`internal/language/extract.go`、`internal/query/runtime.go`
  - 扫描了 `README*`、`internal/skill/skill.md`、`cmd/help_test.go`、`internal/language/language_test.go` 和 `tests/` fixtures 中的 `fn/method` 使用点
- **Key decisions**:
  - Decision: 直接替换公开 kind 契约，不保留 `fn` 或 `method` 解析兼容
  - Reasoning: 用户明确要求不考虑兼容性，且目标是消除误选 kind 导致的漏查
  - Alternatives considered: 仅改帮助文案或为 `func` 增加别名；两者都无法消除现有精确过滤带来的假阴性
- **Problems encountered**:
  - Problem: 用户在当前会话中刚修改过 `overview`
  - Resolution: 将改动面限制在 kind、帮助、文档和测试相关文件，不触碰 `overview` 逻辑
  - Retry count: 0
- **Validation**: `test -f cmd/symbols.go && test -f internal/index/types.go && test -f internal/query/runtime.go` → exit 0
- **Files changed**:
  - `.codex-tasks/20260409-unify-func-kind/TODO.csv` — 推进 milestone 状态
  - `.codex-tasks/20260409-unify-func-kind/PROGRESS.md` — 记录已完成的变更面梳理
- **Next step**: Milestone 2 — 实现公开 kind 统一为 func 并同步文档

---

## Milestone 2: 实现公开 kind 统一为 func 并同步文档

- **Status**: DONE
- **Started**: 00:00
- **Completed**: 00:00
- **What was done**:
  - 将 `internal/index` 与 `internal/language` 的公开 callable kind 统一改为 `SymbolKindFunc`
  - 将 `resolveKind` 中的 `definition.function`、`definition.method`、`definition.macro` 统一映射到 `func`
  - 调整 `internal/query/runtime.go` 的 callable 排序逻辑
  - 同步更新 `cmd/help_kind_support.go`、`README.md`、`README.zh-CN.md`、`internal/skill/skill.md`
- **Key decisions**:
  - Decision: 不保留 `fn` / `method` 的解析兼容
  - Reasoning: 用户明确要求替换公开契约，目标是消除错误 kind 选择导致的漏查
  - Alternatives considered: 为 `func` 增加别名而保留旧值；这会继续暴露两套公开用法，不符合替换目标
- **Problems encountered**:
  - Problem: README 两份表格列宽不一致，首轮大补丁未能完整命中
  - Resolution: 重新读取实际行内容后分文件精确修改
  - Retry count: 1
- **Validation**: `rg -n '"fn"|` + "`fn`" + `|SymbolKindFn|kind": "fn"|kind": "method"' cmd internal README.md README.zh-CN.md tests | head` → exit 1
- **Files changed**:
  - `internal/index/types.go` — 公开 kind 枚举与解析改为 `func`
  - `internal/language/language.go` — 语言层 kind 常量改为 `func`
  - `internal/language/extract.go` — function/method/macro 统一映射为 `func`
  - `internal/query/runtime.go` — callable 排序优先级合并
  - `cmd/help_kind_support.go` — 支持矩阵切到 `func`
  - `README.md` — kind 文档切到 `func`
  - `README.zh-CN.md` — kind 文档切到 `func`
  - `internal/skill/skill.md` — embedded skill 的 kind 文案和示例切到 `func`
- **Next step**: Milestone 3 — 更新测试与 fixture

---

## Milestone 3: 更新测试与 fixture

- **Status**: DONE
- **Started**: 00:00
- **Completed**: 00:00
- **What was done**:
  - 更新了 command help、overview、path args、runtime、output、language、index 相关测试
  - 批量将 `tests/` 下 query/expected 的公开 kind 期望切到 `func`
- **Key decisions**:
  - Decision: 保留原 fixture 目录名 `fn_kind` / `method_kind`
  - Reasoning: 目录名仅作为测试分类标签，不影响公开 CLI 契约；最小化无关文件移动
  - Alternatives considered: 重命名所有 fixture 目录；收益低且会扩大 diff
- **Problems encountered**:
  - Problem: fixture 分布在多语言目录，变更量大
  - Resolution: 先用 `rg -l` 收敛文件列表，再按统一格式逐个更新
  - Retry count: 0
- **Validation**: `rg -n '"kind": "func"|SymbolKindFunc|` + "`func`" + `' cmd internal tests README.md README.zh-CN.md` → exit 0
- **Files changed**:
  - `cmd/help_test.go` — help 期望改为 `func`
  - `cmd/overview_test.go` — overview 输出 kind 期望改为 `func`
  - `cmd/path_args_test.go` — path args 输出 kind 期望改为 `func`
  - `internal/query/runtime_test.go` — runtime 输出 kind 期望改为 `func`
  - `internal/output/output_test.go` — tabular 输出 kind 期望改为 `func`
  - `internal/language/language_test.go` — kind 断言改为 `SymbolKindFunc`
  - `internal/index/types_test.go` — 公共 kind 列表与拒绝集合同步
  - `internal/index/index_test.go` — 索引样本 kind 改为 `SymbolKindFunc`
  - `tests/...` — query/expected 中的公开 kind 全部切到 `func`
- **Next step**: Milestone 4 — 运行验证命令并收尾

---

## Milestone 4: 运行验证命令并收尾

- **Status**: DONE
- **Started**: 00:00
- **Completed**: 00:00
- **What was done**:
  - 运行 `gofmt`
  - 运行完整测试与 lint 验证
- **Key decisions**:
  - Decision: 使用仓库要求的完整校验链路作为关闭条件
  - Reasoning: 这是公开 CLI 契约变更，必须验证编译、测试与 lint 一致通过
  - Alternatives considered: 仅跑局部测试；覆盖不足
- **Problems encountered**:
  - Problem: none
  - Resolution: none
  - Retry count: 0
- **Validation**: `go test ./... -timeout 60s && golangci-lint run ./...` → exit 0
- **Files changed**:
  - `cmd/help_kind_support.go` — gofmt
  - `cmd/help_test.go` — gofmt
  - `cmd/overview_test.go` — gofmt
  - `cmd/path_args_test.go` — gofmt
  - `internal/index/index_test.go` — gofmt
  - `internal/index/types.go` — gofmt
  - `internal/index/types_test.go` — gofmt
  - `internal/language/extract.go` — gofmt
  - `internal/language/language.go` — gofmt
  - `internal/language/language_test.go` — gofmt
  - `internal/output/output_test.go` — gofmt
  - `internal/query/runtime.go` — gofmt
  - `internal/query/runtime_test.go` — gofmt
- **Next step**: none

---

## Final Summary

- **Total milestones**: 4
- **Completed**: 4
- **Failed + recovered**: 0
- **External unblock events**: 0
- **Total retries**: 1
- **Files created**: 3
- **Files modified**: 69
- **Key learnings**:
  - 公开查询 kind 不该要求用户记住语言内的 callable 细分，否则会产生显著假阴性
  - 公开契约替换必须同步 help、README、embedded skill 与 fixture，否则测试很容易只绿一半
- **Recommendations for future tasks**:
  - 当 CLI 公开契约改变时，优先先收敛 `rg` 文件清单，再批量更新 fixture，可减少漏改
