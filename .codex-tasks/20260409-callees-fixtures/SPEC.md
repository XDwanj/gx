# 任务说明

## 目标

为 `gx callees` 在 `tests/` fixture 体系中补齐多语言测试用例，并让 fixture harness 正式支持 `callees` 命令。

## 约束

- 保持现有 `tests/<language>/<command>/<case>/...` 目录约定。
- 只为当前实现已支持 `callees` 的语言添加 fixtures。
- fixture 输出保持 `--json` 结构，与默认 TOON 语义一致。
- 不为暂未支持 `callees` 的语言伪造成功路径。
- 同步更新覆盖校验，避免后续遗漏 public kind/command 组合。

## 验收标准

1. fixture harness 能执行 `callees` fixture。
2. `tests/` 下为支持 `callees` 的语言提供至少一个基础 fixture。
3. fixture 结果覆盖普通函数调用与成员/方法调用等核心路径。
4. 相关覆盖校验和命令测试通过。
5. `env -u GOROOT go test ./... -timeout 60s` 与 `env -u GOROOT golangci-lint run ./...` 通过。

## 非目标

- 不强行为尚未支持 `callees` 的语言补 fixture。
- 不扩展 `callees` 的语义能力边界。
