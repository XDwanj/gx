### L-1: Implemented query-time ignore visibility, include-triggered transient reindexing, and JSON empty-array output; equivalen
- **Strategy:** Implemented query-time ignore visibility, include-triggered transient reindexing, and JSON empty-array output; equivalent built-binary verify reached 0 failures while manifest go-run verify remains externally blocked.
- **Outcome:** keep
- **Insight:** Implemented query-time ignore visibility, include-triggered transient reindexing, and JSON empty-array output; equivalent built-binary verify reached 0 failures while manifest go-run verify remains externally blocked.
- **Context:** goal=修复 gx 路径过滤语义：include 触发按需补索引并覆盖当前 ignore 隐藏，exclude 继续作为查询期排除，普通查询基于当前 ignore 规则实时决定可见性，避免持久化 stale ignored 状态。; scope=CLAUDE.md,README.md,cmd/**,internal/**/*; metric=ignore override failure count; direction=lower
- **Iteration:** include-ignore-fix#2
- **Timestamp:** 2026-04-11T15:43:48Z
