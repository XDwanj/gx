# 目标

在 `gx` 的 `--help` 文档中追加公开 `kind` 与各语言支持程度说明，采用紧凑列表形式，适合终端阅读。

# 范围

- 为 `symbols` 与 `definition` 命令补充 `Long` 帮助文本。
- 追加内容包括：
  - 公开 `kind` 列表
  - 每种活跃语言当前支持的 `kind`
- 文案采用列表形式，不使用宽表格。
- 增加帮助输出测试。

# 验证

- `go test ./... -timeout 60s`
- `golangci-lint run ./...`
