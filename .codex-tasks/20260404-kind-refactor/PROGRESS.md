# 进度记录

## 2026-04-04

- 已确认重构方向以 `Architecture.md` 为准。
- 已读取 `cmd/definition.go`、`cmd/symbols.go`、`cmd/references.go`、`internal/index/types.go`、`internal/language/language.go`、`internal/language/queries.go`、`internal/language/extract.go`、`internal/query/runtime.go` 以及现有相关测试。
- 当前发现：
  - `trait` 和 `event` 仍是公开 kind。
  - `solidity` 与 `elixir` 仍在语言注册和 `lang list` 中。
  - Go 仅支持 `fn`、`method`、`type`，粒度明显不足。
  - Rust、TypeScript、Java、C++ 均存在可补齐的值级或结构级符号提取。
- 已完成公开 `kind` 收缩，当前只保留 `fn`、`method`、`const`、`struct`、`enum`、`class`、`interface`、`module`、`type`。
- 已移除 `solidity` 与 `elixir` 的活跃语言注册、公开列表和无效模块依赖。
- 已补齐重点语言：
  - Go：`struct`、`interface`、包级 `const`、包级 `var -> const`
  - Rust：`trait -> interface`、`const`、`static`、`enum_variant -> const`
  - TypeScript：`const`、枚举成员、`internal_module`
  - Java：`module`、`final` 字段、枚举成员
  - C++：`struct` 与 `class` 拆分、命名空间映射为 `module`
- 为多名字声明新增了提取能力，同一条定义里的多个名字现在都能进索引，不再只保留最后一个 capture。
- 已同步更新 `README.md`、`README.zh-CN.md`、`internal/skill/skill.md`、`Architecture.md`。
- 已完成验证：
  - `go test ./... -timeout 60s`
  - `golangci-lint run ./...`
