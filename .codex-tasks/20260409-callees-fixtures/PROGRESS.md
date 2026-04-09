# 进展记录

## 当前状态

- 任务: 为 `gx callees` 补齐 tests fixture 体系
- 形态: single-full
- 进度: 4/5
- 当前: 运行全量测试与 lint
- 文件: `.codex-tasks/20260409-callees-fixtures/TODO.csv`
- 下一步: 跑 `env -u GOROOT go test ./... -timeout 60s` 与 `env -u GOROOT golangci-lint run ./...`

## 关键决策

- 只为当前已支持 `callees` 的语言补 fixtures。
- 先扩展 harness，再批量补 `tests/<lang>/callees/...`。
- fixture 输出保持 JSON 扁平行结构：`file,line,caller,callee,context`。

## 验证记录

- `env -u GOROOT go test ./cmd -run 'TestLanguageFixtures|TestFixtureCalleesLanguageCoverage' -count=1`：通过
