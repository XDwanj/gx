# 进展记录

## 当前状态

- 任务: 为 `gx` 新增 `callees` 查询命令
- 形态: single-full
- 进度: 4/4
- 当前: 已完成实现、测试与文档同步
- 文件: `.codex-tasks/20260409-callees/TODO.csv`
- 下一步: 无

## 关键决策

- 默认输出采用扁平 TOON，字段为 `file,line,caller,callee,context`。
- `callee` 输出源码中的被调用表达式文本，不做语义级定义解析。
- 预计实现路径为“索引定位 caller + tree-sitter 提取 caller 函数体内的调用表达式”。
- 排序规则采用 `file -> line -> caller -> callee`，与现有查询风格保持一致。

## 验证记录

- `env -u GOROOT go test ./... -timeout 60s`：通过
- `env -u GOROOT golangci-lint run ./...`：通过

## 环境备注

- 本机 `go` 二进制为 `go1.26.2`，但环境变量 `GOROOT` 固定为 `/Users/xdwanj/sdk/go1.26.1`。
- 若直接执行 `go test`，会因标准库版本不一致失败；验证时需显式 `env -u GOROOT`。
