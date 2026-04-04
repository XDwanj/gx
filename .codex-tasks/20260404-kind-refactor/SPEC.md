# 目标

重构 `gx` 的公开 kind 体系与语言支持范围，使其与 `Architecture.md` 中的设计结论一致。

# 范围

- 移除 `solidity` 和 `elixir` 的语言注册与公开支持。
- 将公开 `kind` 收敛为：
  - `fn`
  - `method`
  - `const`
  - `struct`
  - `enum`
  - `class`
  - `interface`
  - `module`
  - `type`
- 去掉 `trait` 和 `event`。
- 优先补齐 Go、Rust、TypeScript、Java、C++ 的 query 与 kind 映射。
- 同步更新测试与文档。

# 验证

- `go test ./... -timeout 60s`
- `golangci-lint run ./...`
