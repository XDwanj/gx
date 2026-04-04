# 目标

为 `gx` 建立基于 `tests/<language>/<command>/<case>/` 的 fixture 集成测试结构，覆盖每个活跃语言的 `symbols` 与 `definition` 基本功能。

# 范围

- 新增通用测试 harness，自动发现并执行 `tests/` 下的语言 fixture。
- 目录结构采用：
  - `tests/<language>/symbol/<case>/...`
  - `tests/<language>/definition/<case>/...`
- 每个 case 至少包含：
  - `_project/`：待索引的源码文件
  - `query.json`：命令参数
  - `expected.json`：期望输出
- 覆盖当前活跃语言：
  - `bash`
  - `c`
  - `cpp`
  - `go`
  - `java`
  - `lua`
  - `python`
  - `ruby`
  - `rust`
  - `swift`
  - `typescript`
  - `zig`

# 验证

- `go test ./... -timeout 60s`
- `golangci-lint run ./...`
