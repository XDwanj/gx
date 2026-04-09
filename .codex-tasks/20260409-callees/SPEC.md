# 任务说明

## 目标

为 `gx` 新增一个 `callees` 查询命令，用于读取指定 `func` 符号的出调用列表。

## 约束

- 沿用现有 Cobra 命令风格与 `path ...` 语义。
- 默认输出使用 TOON。
- 默认 TOON 输出字段为 `file,line,caller,callee,context`。
- `callee` 表示源码里写出的被调用表达式，不做完整语义分派解析。
- 输出需要显式排序，避免依赖 map 或遍历顺序。
- 需要同步更新测试、README、embedded `skill.md`、`--help` 文本。

## 验收标准

1. `gx callees --name 'A' [path ...]` 可返回指定 caller 的调用点列表。
2. 命中多个同名 `A` 时，默认全部返回，不静默挑选。
3. 默认 TOON 输出为扁平行，字段为 `file,line,caller,callee,context`。
4. `--json` 输出与默认 TOON 数据语义一致。
5. 新增或更新自动化测试覆盖命令、查询层、语言提取与公开帮助/文档。
6. `go test ./... -timeout 60s` 与 `golangci-lint run ./...` 通过。

## 非目标

- 不构建全项目语义级调用图。
- 不尝试解析接口分派、反射、函数变量等最终落点。
