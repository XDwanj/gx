# 目标

将 `gx` 的公开路径切换参数从 `--root` 彻底替换为 `-C`，对外契约、帮助、测试和文档统一改为 `-C`，不保留 `--root` 兼容入口。

# 范围

- 调整 `cmd/` 根命令的 persistent flag 定义与相关解析逻辑。
- 保持现有有效目录解析与相对位置参数行为不变，仅替换公开调用方式。
- 删除测试与 fixture 中对 `--root` 的依赖，改为 `-C`。
- 同步更新 `README.md`、`README.zh-CN.md`、嵌入 `internal/skill/skill.md` 与必要的命令帮助文案。
- 按用户纠正规则更新 `CLAUDE.md`，避免再次默认采用兼容过渡方案。

# 非目标

- 不实现 Git 多次 `-C` 叠加切换目录语义。
- 不改动查询输出格式、索引格式和业务行为。

# 验证

- `go test ./cmd -timeout 60s`
- `go test ./... -timeout 60s`
- `golangci-lint run ./...`
