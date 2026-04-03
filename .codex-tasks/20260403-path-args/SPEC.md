# 任务

将 `overview`、`symbols`、`definition`、`references` 的路径作用域从 `--scope` 调整为位置参数。

# 目标

- 保留 `--name`、`--kind`、`--unique` 等语义 flag。
- 路径支持 0..N 个位置参数。
- 未传路径时默认使用当前路径 `.`。
- 直接移除 `--scope`，不做兼容。
- 保持人类输出与 `--json` 输出行为一致。

# 验证

- `go test ./... -timeout 60s`
- `golangci-lint run ./...`
