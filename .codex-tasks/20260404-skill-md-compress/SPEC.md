# Task Specification

## Task Shape

- **Shape**: `single-full`

## Goals

- 将 `internal/skill/skill.md` 从 305 行压缩到 250 行以内。
- 保留示例优先结构和 `gx` 的核心使用规则。
- 删除重复、低价值、可由上下文推出的说明。

## Non-Goals

- 不拆分成多个文件。
- 不修改 `gx` 的实际 CLI 行为。
- 不引入新的命令或语义。

## Constraints

- 保持文档可直接给 agent 使用。
- 优先删除重复解释，而不是删掉关键例子。
- 修改仅限 `internal/skill/skill.md` 和任务跟踪文件。

## Environment

- **Project root**: `/Users/xdwanj/Project/Rust/gx`
- **Language/runtime**: `Markdown`
- **Build command**: `n/a`
- **Validation**: `wc -l` + 关键段落人工复查

## Deliverables

- 压缩后的 `internal/skill/skill.md`
- `.codex-tasks/20260404-skill-md-compress/` 任务记录

## Done-When

- [ ] `internal/skill/skill.md` 行数 <= 250
- [ ] 保留 `overview / symbols / definition / references` 使用说明
- [ ] 保留 `--name`、路径规则、输出模式、实用启发

## Final Validation Command

```bash
wc -l internal/skill/skill.md
```
