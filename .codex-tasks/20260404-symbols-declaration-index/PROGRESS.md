# Progress Log

> Auto-maintained by Taskmaster. Each entry records what happened, why, and what's next.
> This file serves as both decision audit trail and context-recovery anchor.

---

## Session Start

- **Date**: 2026-04-04
- **Task name**: `20260404-symbols-declaration-index`
- **Task dir**: `.codex-tasks/20260404-symbols-declaration-index/`
- **Spec**: See `SPEC.md`
- **Plan**: See `TODO.csv` (5 milestones)
- **Environment**: `Go / go test / golangci-lint`

---

## Context Recovery Block

- **Current milestone**: `#5 — Run full validation`
- **Current status**: `DONE`
- **Last completed**: `#5 — Run full validation`
- **Current artifact**: `TODO.csv`
- **Key context**: The user corrected the previous design choice and `gx symbols` now keeps only `file + line` as the default declaration index coordinate contract.
- **Known issues**: None in this task scope.
- **Next action**: Deliver summary to the user.

---

## Milestone 1: Create task record and confirm change surface

- **Status**: DONE
- **Started**: 20:52
- **Completed**: 20:53
- **What was done**:
  - Created `.codex-tasks/20260404-symbols-declaration-index/` with `SPEC.md`, `TODO.csv`, and `PROGRESS.md`.
  - Confirmed the change surface across `internal/language`, `internal/index`, `internal/query`, `cmd`, README, and symbol fixtures.
- **Key decisions**:
  - Decision: Treat this as an intentional breaking change for `symbols`.
  - Reasoning: The user explicitly approved breaking the current output contract.
  - Alternatives considered: Adding an opt-in flag such as `--coords`, which was rejected for this task.
- **Problems encountered**:
  - Problem: None.
  - Resolution: N/A
  - Retry count: 0
- **Validation**: `test -f .codex-tasks/20260404-symbols-declaration-index/SPEC.md` → exit 0
- **Files changed**:
  - `.codex-tasks/20260404-symbols-declaration-index/SPEC.md` — task scope and constraints
  - `.codex-tasks/20260404-symbols-declaration-index/TODO.csv` — milestone tracking
  - `.codex-tasks/20260404-symbols-declaration-index/PROGRESS.md` — recovery anchor
- **Next step**: Milestone 2 — Persist symbol coordinates in extraction and index

---

## Milestone 2: Persist symbol coordinates in extraction and index

- **Status**: DONE
- **Started**: 20:53
- **Completed**: 20:55
- **What was done**:
  - Added `Line` and `Column` to `language.Symbol` and `index.Symbol`.
  - Captured coordinates from Tree-sitter `StartPosition()` during extraction.
  - Bumped `IndexVersion` to force cache refresh with the new symbol shape.
- **Key decisions**:
  - Decision: Store line and column in the index instead of computing them on demand.
  - Reasoning: Coordinates are part of the symbol contract now and should be reusable across queries.
  - Alternatives considered: Recomputing line numbers from `ByteStart` during `symbols`, which would duplicate work and still not provide stored columns.
- **Problems encountered**:
  - Problem: None.
  - Resolution: N/A
  - Retry count: 0
- **Validation**: `go test ./internal/language ./internal/index -timeout 60s` → exit 0
- **Files changed**:
  - `internal/language/language.go` — symbol model
  - `internal/language/extract.go` — Tree-sitter coordinate extraction
  - `internal/index/types.go` — persisted index symbol fields
  - `internal/index/index.go` — cache version and symbol mapping
  - `internal/language/language_test.go` — coordinate assertions
- **Next step**: Milestone 3 — Change symbols output shape and update unit tests

---

## Milestone 3: Change symbols output shape and update unit tests

- **Status**: DONE
- **Started**: 20:55
- **Completed**: 20:56
- **What was done**:
  - Changed `SymbolRow` to always emit `file`, `line`, `column`, `name`, `kind`, and `signature`.
  - Switched symbol sorting to source-order within each file.
  - Updated unit tests for single-file, directory, multiple-path, JSON, and path-resolution behavior.
- **Key decisions**:
  - Decision: Always include `file` even for single-file scopes.
  - Reasoning: A declaration index should stay directly citable and machine-consumable in all scopes.
  - Alternatives considered: Keeping the old single-file omission behavior, which would undermine the new contract.
- **Problems encountered**:
  - Problem: None.
  - Resolution: N/A
  - Retry count: 0
- **Validation**: `go test ./cmd ./internal/query -timeout 60s` → exit 0
- **Files changed**:
  - `internal/query/runtime.go` — symbol query output and ordering
  - `internal/query/runtime_test.go` — updated and added query tests
  - `cmd/path_args_test.go` — updated command expectations
- **Next step**: Milestone 4 — Refresh fixture expectations and docs

---

## Milestone 4: Refresh fixture expectations and docs

- **Status**: DONE
- **Started**: 20:56
- **Completed**: 20:57
- **What was done**:
  - Updated English and Chinese README text to describe the new declaration-index output.
  - Regenerated every `tests/*/symbol/*/expected.json` fixture from the current command behavior.
- **Key decisions**:
  - Decision: Regenerate symbol fixtures in bulk instead of editing them manually.
  - Reasoning: The change touches many languages and also changes ordering, so regeneration is less error-prone.
  - Alternatives considered: Hand-editing JSON fixtures, which would be slower and easier to get wrong.
- **Problems encountered**:
  - Problem: The first batch script used `mapfile`, which is unavailable in the system `bash`.
  - Resolution: Rewrote the loop with a portable `while read` array fill.
  - Retry count: 1
- **Validation**: `go test ./... -timeout 60s` → exit 0
- **Files changed**:
  - `README.md` — English command/docs update
  - `README.zh-CN.md` — Chinese command/docs update
  - `tests/*/symbol/*/expected.json` — regenerated symbol fixtures
- **Next step**: Milestone 5 — Run full validation

---

## Milestone 5: Run full validation

- **Status**: DONE
- **Started**: 20:57
- **Completed**: 20:58
- **What was done**:
  - Ran full repository tests with timeout enforcement.
  - Ran the repository hard-gate lint command.
- **Key decisions**:
  - Decision: Validate with the repository-wide gates rather than only the touched packages.
  - Reasoning: The change intentionally breaks a public output contract and touches many fixtures.
  - Alternatives considered: Stopping after package-local tests, which would miss fixture and lint regressions.
- **Problems encountered**:
  - Problem: None.
  - Resolution: N/A
  - Retry count: 0
- **Validation**: `go test ./... -timeout 60s && golangci-lint run ./...` → exit 0
- **Files changed**:
  - No additional source changes during validation
- **Next step**: Final user summary

---

## Final Summary

- **Total milestones**: 5
- **Completed**: 5
- **Failed + recovered**: 1
- **External unblock events**: 0
- **Total retries**: 1
- **Files created**: 3
- **Files modified**: 78
- **Key learnings**:
  - Persisting coordinates in the index keeps symbol queries cheap and predictable.
  - Making `symbols` source-ordered changes grouped declaration fixtures in multiple languages, so regeneration is the safest update path.
- **Recommendations for future tasks**:
  - If `receiver` or `package` are added later, keep them as additive symbol metadata on top of the current declaration-index contract.

---

## Milestone 2: Remove column from extraction and index

- **Status**: DONE
- **Started**: 21:03
- **Completed**: 21:05
- **What was done**:
  - Removed `Column` from `language.Symbol` and `index.Symbol`.
  - Simplified extraction to persist only `Line` alongside byte ranges.
- **Key decisions**:
  - Decision: Keep line as the only persisted declaration coordinate.
  - Reasoning: The user explicitly rejected `column` as unnecessary for the default workflow.
  - Alternatives considered: Leaving `column` in storage but dropping it from output, which would keep unused metadata around.
- **Problems encountered**:
  - Problem: None.
  - Resolution: N/A
  - Retry count: 0
- **Validation**: `go test ./internal/language ./internal/index -timeout 60s` → exit 0
- **Files changed**:
  - `internal/language/language.go` — removed `Column`
  - `internal/language/extract.go` — line-only extraction
  - `internal/index/types.go` — removed persisted `column`
  - `internal/index/index.go` — removed index mapping for `column`
- **Next step**: Milestone 3 — Remove column from symbols output and unit tests

---

## Milestone 3: Remove column from symbols output and unit tests

- **Status**: DONE
- **Started**: 21:05
- **Completed**: 21:06
- **What was done**:
  - Removed `column` from `SymbolRow` and from JSON/TOON output.
  - Simplified in-file sorting to `line` then `name`.
  - Updated query and command tests to assert the line-only contract.
- **Key decisions**:
  - Decision: Keep `file,line,name,kind,signature` as the public row shape.
  - Reasoning: This preserves the declaration-index value while matching the user's stated need.
  - Alternatives considered: Dropping stored coordinates entirely and recomputing lines from bytes, which would add avoidable work.
- **Problems encountered**:
  - Problem: None.
  - Resolution: N/A
  - Retry count: 0
- **Validation**: `go test ./cmd ./internal/query -timeout 60s` → exit 0
- **Files changed**:
  - `internal/query/runtime.go` — removed `column` from symbols output
  - `internal/query/runtime_test.go` — updated assertions
  - `cmd/path_args_test.go` — updated command expectations
- **Next step**: Milestone 4 — Refresh fixture expectations and docs

---

## Milestone 4: Refresh fixture expectations and docs

- **Status**: DONE
- **Started**: 21:06
- **Completed**: 21:07
- **What was done**:
  - Updated README text in both languages to describe the line-only declaration index.
  - Regenerated all `tests/*/symbol/*/expected.json` fixtures without `column`.
- **Key decisions**:
  - Decision: Regenerate fixtures from the CLI again instead of editing them by hand.
  - Reasoning: The output shape changed across many language suites and regeneration stays consistent with real behavior.
  - Alternatives considered: Manual JSON edits, which were slower and riskier.
- **Problems encountered**:
  - Problem: None after reusing the portable regeneration script.
  - Resolution: N/A
  - Retry count: 0
- **Validation**: `go test ./... -timeout 60s` → exit 0
- **Files changed**:
  - `README.md` — line-only symbol index docs
  - `README.zh-CN.md` — line-only symbol index docs
  - `tests/*/symbol/*/expected.json` — regenerated fixtures
- **Next step**: Milestone 5 — Run full validation

---

## Milestone 5: Run full validation

- **Status**: DONE
- **Started**: 21:07
- **Completed**: 21:07
- **What was done**:
  - Ran full repository tests with timeout enforcement.
  - Ran the repository lint gate.
- **Key decisions**:
  - Decision: Keep the same validation gates as the previous revision.
  - Reasoning: The output contract changed again and needed full confirmation.
  - Alternatives considered: Only rerunning touched-package tests, which would miss fixture regressions.
- **Problems encountered**:
  - Problem: None.
  - Resolution: N/A
  - Retry count: 0
- **Validation**: `go test ./... -timeout 60s && golangci-lint run ./...` → exit 0
- **Files changed**:
  - No additional source changes during validation
- **Next step**: Final user summary
