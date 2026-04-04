# 进度记录

## 2026-04-04

- 已确认根因在 [cmd/root.go](/Users/xdwanj/Project/Rust/gx/cmd/root.go)：`resolveTargetPaths()` 一直按当前 `cwd` 解析相对位置参数，而查询运行时和索引过滤又按有效 `root` 工作。
- 已确认受影响命令包括：
  - `overview`
  - `symbols`
  - `definition`
  - `references`
- 用户复现场景 `./gx --root ../cx_link o .` 属于典型错位：索引根是 `cx`，位置参数 `.` 却解析成了当前 `gx` 目录。
- 已在 [cmd/root.go](/Users/xdwanj/Project/Rust/gx/cmd/root.go) 调整路径解析：
  - 未显式传 `--root` 时，仍按当前 `cwd` 解析相对位置参数，保持现有习惯。
  - 显式传 `--root` 时，相对位置参数与无参数默认目标都按有效 root 解析。
- 已在 [internal/app/runtime.go](/Users/xdwanj/Project/Rust/gx/internal/app/runtime.go) 将显式 `--root` 先做绝对化，再做 `EvalSymlinks`，修复 `../cx_link` 这类目录符号链接根路径的索引与查询错位问题。
- 已补命令级回归测试到 [cmd/path_args_test.go](/Users/xdwanj/Project/Rust/gx/cmd/path_args_test.go)，覆盖：
  - 显式 root 的 `overview`
  - 相对 symlink root 的 `overview`
  - 显式 root 的 `symbols`
  - 显式 root 的 `definition`
  - 显式 root 的 `references`
- 已同步修正 [cmd/cache_test.go](/Users/xdwanj/Project/Rust/gx/cmd/cache_test.go)，使缓存路径断言与新的 root 规范化语义一致，避免 macOS `/var` 与 `/private/var` 路径差异导致的回归。
- 真实复现场景已验证通过：
  - `go run . --root ../cx_link o .`
- 已完成验证：
  - `go test ./cmd -run 'PathArgs|Overview|Symbols|Definition|References' -timeout 60s`
  - `go test ./... -timeout 60s`
  - `golangci-lint run ./...`
