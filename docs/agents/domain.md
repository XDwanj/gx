# Domain Docs

engineering skills 在探索代码库时按本文件的规则消费领域文档。

## 探索前先读这些

- 仓库根目录的 **`GLOSSARY.md`**；如果存在 **`GLOSSARY-MAP.md`** 则以它为准，它指向每个 context 各自的 `GLOSSARY.md`，按话题读相关的即可。
- **`docs/adr/`**：读与你要改动的区域相关的 ADR。多 context 仓库还要查 `src/<context>/docs/adr/` 下的 context 级决策。

如果这些文件不存在，**静默跳过**：不要提示缺失，也不要主动建议创建。`/domain-modeling` skill（经由 `/grill-with-docs` 和 `/improve-codebase-architecture` 触达）会在术语或决策真正落定的时候懒创建它们。

## 文件结构

Single-context 仓库（大多数仓库）：

```
/
├── GLOSSARY.md
├── docs/adr/
│   ├── 0001-event-sourced-orders.md
│   └── 0002-postgres-for-write-model.md
└── src/
```

Multi-context 仓库（根目录存在 `GLOSSARY-MAP.md` 时）：

```
/
├── GLOSSARY-MAP.md
├── docs/adr/                          ← 系统级决策
└── src/
    ├── ordering/
    │   ├── GLOSSARY.md
    │   └── docs/adr/                  ← context 级决策
    └── billing/
        ├── GLOSSARY.md
        └── docs/adr/
```

## 使用 glossary 的词汇

当你的输出要命名一个领域概念（issue 标题、重构提案、假设、测试名）时，使用 `GLOSSARY.md` 里定义的术语，不要漂移到 glossary 明确回避的同义词。

如果需要的概念不在 glossary 里，这是个信号：要么你在发明项目并不使用的说法（重新考虑），要么确实是个缺口（记下来交给 `/domain-modeling`）。

## 标记 ADR 冲突

如果你的输出与已有 ADR 矛盾，显式指出来，不要静默覆盖：

> _与 ADR-0007（event-sourced orders）冲突，但值得重开讨论，因为……_
