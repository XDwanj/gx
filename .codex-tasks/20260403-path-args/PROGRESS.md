# 当前状态

任务: 将路径作用域从 `--scope` 迁到位置参数
形态: single-full
进度: 5/5
当前: 已完成
验证: `go test ./... -timeout 60s` 通过；`golangci-lint run ./...` 通过
文件: `.codex-tasks/20260403-path-args/TODO.csv`

# 备注

- 已先读取 `cmd/` 入口、`internal/query/runtime.go` 和现有相关测试。
- 已先改查询层多路径过滤，再改 Cobra 参数，避免中间态不一致。
- 新语义为：路径使用位置参数传入；省略路径时默认当前工作目录；相对路径按当前工作目录解析。
