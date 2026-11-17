# Issue tracker: Local Markdown

本仓库的 issue 和 spec 以 markdown 文件形式存放在 `.scratch/` 下。

## 约定

- 一个 feature 一个目录：`.scratch/<feature-slug>/`
- spec 文件是 `.scratch/<feature-slug>/spec.md`
- 实现 issue 一票一文件，放在 `.scratch/<feature-slug>/issues/<NN>-<slug>.md`，编号从 `01` 开始，不要把多张票合并写进一个文件
- Triage 状态记录在每张 issue 文件顶部附近的 `Status:` 行（角色字符串见 `triage-labels.md`）
- 评论和对话历史追加到文件底部的 `## Comments` 标题下

## 当 skill 要求 "publish to the issue tracker"

在 `.scratch/<feature-slug>/` 下新建文件（目录不存在则创建）。

## 当 skill 要求 "fetch the relevant ticket"

读取引用路径对应的文件。用户一般会直接给路径或 issue 编号。

## Wayfinding 操作

供 `/wayfinder` 使用。**map** 是一个文件，**child** 一票一文件。

- **Map**：`.scratch/<effort>/map.md`（正文是 Notes / Decisions-so-far / Fog）。
- **Child ticket**：`.scratch/<effort>/issues/NN-<slug>.md`，编号从 `01` 开始，问题写在正文里。`Type:` 行记录票类型（`research`/`prototype`/`grilling`/`task`）；`Status:` 行记录 `claimed`/`resolved`。
- **Blocking**：顶部附近的 `Blocked by: NN, NN` 行。一张票在它列出的所有文件都变为 `resolved` 后才算解除阻塞。
- **Frontier**：扫描 `.scratch/<effort>/issues/` 下未关闭、未被阻塞、未被认领的文件，编号最小者优先。
- **Claim**：开始任何工作前，先把 `Status:` 改为 `claimed` 并保存。
- **Resolve**：在 `## Answer` 标题下追加答案，把 `Status:` 改为 `resolved`，然后往 `map.md` 的 Decisions-so-far 里追加一条上下文指针（gist + 链接）。
