# 目标

为 `gx symbols --kind` 和 `gx definition --kind` 增加 shell completion，补全候选与当前公开 `kind` 集合保持一致。

# 范围

- 将公开 `kind` 集合收敛为可复用真源，避免 completion、help 文案和解析逻辑漂移。
- 给 `symbols` 与 `definition` 的 `--kind` 注册 Cobra flag completion。
- 补命令级测试，验证 completion 候选包含全部公开 `kind`。

# 验证

- `go test ./cmd -run 'Help|Completion' -timeout 60s`
- `go test ./... -timeout 60s`
- `golangci-lint run ./...`
