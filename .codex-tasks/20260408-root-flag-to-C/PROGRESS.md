# 进度记录

## 2026-04-08

- 已确认这是一次公开 CLI 契约变更，不做兼容过渡。
- 已定位受影响入口：
  - [cmd/root.go](/Users/xdwanj/Project/Rust/gx/cmd/root.go)
  - [internal/app/runtime.go](/Users/xdwanj/Project/Rust/gx/internal/app/runtime.go)
  - [cmd/path_args_test.go](/Users/xdwanj/Project/Rust/gx/cmd/path_args_test.go)
  - [cmd/fixture_test.go](/Users/xdwanj/Project/Rust/gx/cmd/fixture_test.go)
  - [cmd/cache_test.go](/Users/xdwanj/Project/Rust/gx/cmd/cache_test.go)
  - [README.md](/Users/xdwanj/Project/Rust/gx/README.md)
  - [README.zh-CN.md](/Users/xdwanj/Project/Rust/gx/README.zh-CN.md)
  - [internal/skill/skill.md](/Users/xdwanj/Project/Rust/gx/internal/skill/skill.md)
  - [CLAUDE.md](/Users/xdwanj/Project/Rust/gx/CLAUDE.md)
- 当前执行策略：
  - 仅保留 `-C` 作为公开切换目录参数。
  - 保持现有路径解析行为，避免把这次契约替换扩大成语义重写。
- 已完成代码层替换：
  - [cmd/root.go](/Users/xdwanj/Project/Rust/gx/cmd/root.go) 现已改为 `-C` / `--chdir` 接口，并移除 `--root`。
  - [internal/app/runtime.go](/Users/xdwanj/Project/Rust/gx/internal/app/runtime.go) 的 flag 字段已从 `Root` 改为 `Directory`。
- 已完成测试同步：
  - [cmd/root_test.go](/Users/xdwanj/Project/Rust/gx/cmd/root_test.go) 新增 `-C` 注册断言，并确认 `root` flag 不再暴露。
  - [cmd/path_args_test.go](/Users/xdwanj/Project/Rust/gx/cmd/path_args_test.go)、[cmd/fixture_test.go](/Users/xdwanj/Project/Rust/gx/cmd/fixture_test.go)、[cmd/cache_test.go](/Users/xdwanj/Project/Rust/gx/cmd/cache_test.go)、[cmd/overview_test.go](/Users/xdwanj/Project/Rust/gx/cmd/overview_test.go) 已全部切到新字段与新调用方式。
- 已完成文档与规则同步：
  - [README.md](/Users/xdwanj/Project/Rust/gx/README.md)
  - [README.zh-CN.md](/Users/xdwanj/Project/Rust/gx/README.zh-CN.md)
  - [internal/skill/skill.md](/Users/xdwanj/Project/Rust/gx/internal/skill/skill.md)
  - [CLAUDE.md](/Users/xdwanj/Project/Rust/gx/CLAUDE.md)
- 已完成阶段验证：
  - `go test ./cmd -timeout 60s`
- 全量验证已通过：
  - `go test ./... -timeout 60s`
  - `golangci-lint run ./...`
- 已检查终端帮助输出，`go run . --help` 现显示：
  - `-C, --chdir string   Run as if gx was started in this directory`
- 已确认仓库业务文档与代码中不再残留公开 `--root` 文案：
  - `rg -n --hidden --glob '!internal/grammars/**' --glob '!.codex-tasks/**' -- '--root' .`
- lint 过程中暴露出一个与本次改动无关的现有问题，已顺手修复以恢复仓库硬门：
  - [internal/language/language.go](/Users/xdwanj/Project/Rust/gx/internal/language/language.go) 中 `walkReferenceLeaves` 的未使用参数已改为 `_`，不影响行为。
