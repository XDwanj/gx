# Progress Log

> Auto-maintained by Taskmaster. Each entry records what happened, why, and what's next.
> This file serves as both decision audit trail and context-recovery anchor.

---

## Session Start

- **Date**: 2026-04-02 22:22
- **Task name**: `20260402-rg-ignore-support`
- **Task dir**: `.codex-tasks/20260402-rg-ignore-support/`
- **Spec**: See `SPEC.md`
- **Plan**: See `TODO.csv` (3 milestones)
- **Environment**: Go / Go modules / `go test`

---

## Context Recovery Block

> If you are resuming this task after compaction, session restart, or context loss,
> read this section FIRST to restore working state.

- **Current milestone**: #4 — Extend traversal to nested `.gitignore`
- **Current status**: DONE
- **Last completed**: #4 — Extend traversal to nested `.gitignore`
- **Current artifact**: `.codex-tasks/20260402-rg-ignore-support/TODO.csv`
- **Key context**: `walk` now merges directory-scoped `.gitignore` and `.ignore` files by rewriting scoped rules to root-relative patterns before compiling them. Focused tests, full `go test`, and `golangci-lint` all pass.
- **Known issues**: None.
- **Next action**: Task complete.

---

## Milestone 1: Scaffold task artifacts

- **Status**: DONE
- **Started**: 22:22
- **Completed**: 22:22
- **What was done**:
  - Created the task directory with `SPEC.md`, `TODO.csv`, and `PROGRESS.md`.
- **Key decisions**:
  - Decision: Use `single-full`.
  - Reasoning: This is a code change with multiple ordered steps and explicit validation gates.
  - Alternatives considered: `single-compact`, rejected because persistent recovery artifacts are more appropriate for code changes.
- **Problems encountered**:
  - Problem: None.
  - Resolution: N/A
  - Retry count: 0
- **Validation**: `test -f .codex-tasks/20260402-rg-ignore-support/SPEC.md && test -f .codex-tasks/20260402-rg-ignore-support/TODO.csv && test -f .codex-tasks/20260402-rg-ignore-support/PROGRESS.md` → exit 0
- **Files changed**:
  - `.codex-tasks/20260402-rg-ignore-support/SPEC.md` — task scope, constraints, and acceptance criteria
  - `.codex-tasks/20260402-rg-ignore-support/TODO.csv` — milestone tracking
  - `.codex-tasks/20260402-rg-ignore-support/PROGRESS.md` — recovery and audit log
- **Next step**: Milestone 2 — Implement layered `.ignore` traversal

---

## Milestone 2: Implement layered `.ignore` traversal

- **Status**: DONE
- **Started**: 22:22
- **Completed**: 22:26
- **What was done**:
  - Reworked `walk` to keep per-directory traversal state instead of a single root `.gitignore` matcher.
  - Added support for root and nested `.ignore` files by rewriting each file's rules into root-relative patterns before compiling the aggregate matcher.
  - Preserved `.gx-ignore` directory skipping and `.git` pruning.
  - Added tests for root `.ignore`, scoped `.ignore`, and combined `.gitignore` + `.ignore` behavior.
- **Key decisions**:
  - Decision: Rewrite scoped `.ignore` patterns to root-relative paths and compile one ordered matcher per directory state.
  - Reasoning: The existing dependency cannot distinguish between "no match" and "negated match" across separate ignore files, so preserving rule ordering required a single aggregate matcher.
  - Alternatives considered: Layered per-file matcher stacks, rejected because negation semantics could not be represented correctly with the available API.
- **Problems encountered**:
  - Problem: A test initially assumed `!build/keep.go` could re-include a file inside an ignored directory.
  - Resolution: Corrected the test to match git/ripgrep behavior, where ignored parent directories prevent re-inclusion of descendants.
  - Retry count: 1
- **Validation**: `go test ./internal/index -run 'TestWalk' -timeout 60s` → exit 0
- **Files changed**:
  - `internal/index/index.go` — layered ignore traversal and scoped rule rewriting
  - `internal/index/index_test.go` — `.ignore` coverage and interaction tests
- **Next step**: Milestone 3 — Add docs and run full validation

---

## Milestone 3: Add docs and run full validation

- **Status**: DONE
- **Started**: 22:26
- **Completed**: 22:28
- **What was done**:
  - Updated both READMEs to mention `.ignore` in the indexing flow.
  - Removed an unused helper flagged by `golangci-lint`.
  - Ran full repository tests and lint.
- **Key decisions**:
  - Decision: Keep documentation changes minimal and focused on the indexing flow.
  - Reasoning: The user requested support for `.ignore`; the existing flow diagram is the most direct place to surface the new behavior.
  - Alternatives considered: Broader README expansion, deferred because the current docs already describe indexing at a high level.
- **Problems encountered**:
  - Problem: `golangci-lint` flagged `compileIgnore` as dead code after the traversal refactor.
  - Resolution: Removed the unused helper and reran all validation.
  - Retry count: 1
- **Validation**: `go test ./... -timeout 60s && golangci-lint run ./...` → exit 0
- **Files changed**:
  - `README.md` — indexing flow now mentions `.ignore`
  - `README.zh-CN.md` — 索引流程现在提到 `.ignore`
  - `internal/index/index.go` — removed unused helper after lint feedback
- **Next step**: Task complete

---

## Milestone 4: Extend traversal to nested `.gitignore`

- **Status**: DONE
- **Started**: 22:31
- **Completed**: 22:36
- **What was done**:
  - Updated traversal state building so every directory loads its own `.gitignore` before its `.ignore`.
  - Added tests proving nested `.gitignore` files affect only their subtree and that same-directory `.ignore` rules can override `.gitignore`.
  - Re-ran focused walk tests, full repository tests, and lint.
- **Key decisions**:
  - Decision: Keep `.gitignore` earlier than `.ignore` in the per-directory rule order.
  - Reasoning: This preserves the existing intent that `.ignore` remains the more explicit local override layer.
  - Alternatives considered: Appending `.gitignore` after `.ignore`, rejected because it would make VCS rules override user-local ignore rules in the same directory.
- **Problems encountered**:
  - Problem: The original task spec declared nested `.gitignore` support out of scope.
  - Resolution: Updated the task spec and tracking artifacts to reflect the approved scope expansion.
  - Retry count: 0
- **Validation**: `go test ./internal/index -run 'TestWalk' -timeout 60s && go test ./... -timeout 60s && golangci-lint run ./...` → exit 0
- **Files changed**:
  - `internal/index/index.go` — nested `.gitignore` files are now loaded per directory
  - `internal/index/index_test.go` — added nested `.gitignore` and override coverage
  - `.codex-tasks/20260402-rg-ignore-support/SPEC.md` — scope updated for nested `.gitignore`
  - `.codex-tasks/20260402-rg-ignore-support/TODO.csv` — appended milestone 4
  - `.codex-tasks/20260402-rg-ignore-support/PROGRESS.md` — logged scope extension and validation
- **Next step**: Task complete

---

## Final Summary

- **Total milestones**: 4
- **Completed**: 4
- **Failed + recovered**: 0
- **External unblock events**: 0
- **Total retries**: 2
- **Files created**: 3
- **Files modified**: 5
- **Key learnings**:
  - The existing ignore library must be wrapped to honor per-directory `.gitignore` and `.ignore` scope while preserving rule ordering.
- **Recommendations for future tasks**:
  - Keep ignore semantics centralized in traversal helpers so indexing and cache invalidation stay aligned.
