# 目标

修复 `gx` 在显式传入 `--root` 时，相对位置参数仍按当前 shell 工作目录解析的问题，使 `overview`、`symbols`、`definition`、`references` 的相对路径与显式 `--root` 保持一致。

# 范围

- 调整 `cmd/` 层路径解析逻辑。
- 当显式传入 `--root` 时：
  - 相对位置参数相对有效 root 解析。
  - 无位置参数时默认作用域也落到有效 root。
- 不改变绝对路径参数行为。
- 补命令级回归测试，覆盖从 root 外部执行 `gx --root <other-project> ... .` 的场景。

# 验证

- `go test ./cmd -run 'PathArgs|Overview|Symbols|Definition|References' -timeout 60s`
- `go test ./... -timeout 60s`
- `golangci-lint run ./...`
