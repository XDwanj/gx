# 进度记录

## 2026-04-04

- 已确认当前 `tests/` 只保证了全局公开 `kind` 总覆盖，还没有保证“每个语言声明支持的 kind 都被该语言 fixture 覆盖”。
- 已根据 `cmd/help_kind_support.go` 的语言支持列表和当前 `tests/` 做差集分析，确认多个语言仍缺大量 `symbol`/`definition` kind 用例。
- 已将语言支持矩阵收敛到 [cmd/help_kind_support.go](/Users/xdwanj/Project/Rust/gx/cmd/help_kind_support.go)，`--help` 文案和 fixture 覆盖测试现在共用一份真源。
- 已在 [cmd/fixture_test.go](/Users/xdwanj/Project/Rust/gx/cmd/fixture_test.go) 增加 `TestFixtureLanguageKindCoverage`，直接校验每个语言声明支持的每个 `kind` 都在 `symbol` 与 `definition` fixture 中出现。
- 已补齐 `tests/` 下所有活跃语言的缺失 `kind` case，目录继续采用 `tests/<language>/<command>/<case>/`，case 名统一为 `<kind>_kind`。
- 已顺手修正两个已有 definition fixture 的 `query.kind` 缺口：
  - [tests/c/definition/basic/query.json](/Users/xdwanj/Project/Rust/gx/tests/c/definition/basic/query.json)
  - [tests/cpp/definition/basic/query.json](/Users/xdwanj/Project/Rust/gx/tests/cpp/definition/basic/query.json)
- 已用脚本做一轮矩阵校验，当前“语言声明支持的 kind 与 tests 实际覆盖”差集为 `OK`。
- 已完成验证：
  - `go test ./cmd -run Fixture -timeout 60s`
  - `go test ./... -timeout 60s`
  - `golangci-lint run ./...`
