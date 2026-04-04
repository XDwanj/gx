# 目标

为 `gx` 的每个活跃语言补齐对应 `kind` 的 fixture 测试，确保 `tests/` 不仅覆盖全局公开 `kind`，也覆盖每个语言当前声明支持的全部 `kind`。

# 范围

- 以当前公开 `kind` 集合与语言支持矩阵为准，补齐 `tests/<language>/<command>/<case>/`。
- `symbol` 侧要求每个语言的支持 `kind` 都至少在 fixture 输出中出现一次。
- `definition` 侧要求每个语言的支持 `kind` 都至少有一个带 `query.kind` 的 fixture 用例。
- 将语言支持矩阵收敛为代码内可复用真源，避免 `--help` 文案与测试断言漂移。
- 新增按语言的 fixture 覆盖断言。

# 验证

- `go test ./cmd -run Fixture -timeout 60s`
- `go test ./... -timeout 60s`
- `golangci-lint run ./...`
