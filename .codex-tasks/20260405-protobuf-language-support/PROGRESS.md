# Progress Log

> Auto-maintained by Taskmaster. Each entry records what happened, why, and what's next.
> This file serves as both decision audit trail and context-recovery anchor.

---

## Session Start

- **Date**: 2026-04-05 22:24
- **Task name**: `20260405-protobuf-language-support`
- **Task dir**: `.codex-tasks/20260405-protobuf-language-support/`
- **Spec**: See `SPEC.md`
- **Plan**: See `TODO.csv` (4 milestones)
- **Environment**: `Go / Cobra / go test`

---

## Context Recovery Block

> If you are resuming this task after compaction, session restart, or context loss,
> read this section FIRST to restore working state.

- **Current milestone**: `none`
- **Current status**: `DONE`
- **Last completed**: `#4 — Run repository validation gates`
- **Current artifact**: `.codex-tasks/20260405-protobuf-language-support/TODO.csv`
- **Key context**: `protobuf`, `typescript`, `swift`, and `zig` now all use external modules. `swift` and `zig` were made consumable by forking and repairing upstream Go packaging in `XDwanj/tree-sitter-swift` and `XDwanj/tree-sitter-zig`. `internal/grammars/` has been removed entirely.
- **Known issues**: None.
- **Next action**: None.

---

## Milestone 1: Confirm integration points and test matrix

- **Status**: DONE
- **Started**: 22:24
- **Completed**: 22:28
- **What was done**:
  - Read `cmd/`, `internal/lang/`, `internal/language/`, and existing tests before any production edits.
  - Confirmed the external `tree-sitter-proto` grammar exposes named nodes needed for `message`, `enum`, `service`, `rpc`, and `package`.
- **Key decisions**:
  - Decision: Keep the current public `kind` set unchanged for Proto support.
  - Reasoning: The user explicitly asked not to add or remove public kinds.
  - Alternatives considered: Adding Proto-specific kinds was rejected to keep CLI contracts stable.
- **Problems encountered**:
  - Problem: The external grammar repository has no semver tags.
  - Resolution: Use a commit-pinned module reference instead of an unpinned branch.
  - Retry count: 0
- **Validation**: `test -f cmd/fixture_test.go && test -f internal/language/language.go` → exit 0
- **Files changed**:
  - `.codex-tasks/20260405-protobuf-language-support/TODO.csv` — advanced milestone state.
  - `.codex-tasks/20260405-protobuf-language-support/PROGRESS.md` — recorded recovery state and milestone details.
- **Next step**: Milestone 2 — Implement Protobuf language registration and extraction

---

## Milestone 2: Implement Protobuf language registration and extraction

- **Status**: DONE
- **Started**: 22:28
- **Completed**: 22:41
- **What was done**:
  - Added `protobuf` to the supported language list and the help kind matrix.
  - Registered `.proto` detection, Protobuf symbol queries, and a local grammar loader.
  - Vendored the external `tree-sitter-proto` parser sources into `internal/grammars/protobuf`.
- **Key decisions**:
  - Decision: Vendor the third-party grammar instead of importing the upstream Go module directly.
  - Reasoning: Direct import upgraded the repository `go` directive from `1.24.5` to `1.26.0`, which was an unnecessary project-wide side effect.
  - Alternatives considered: Keeping the direct module dependency was rejected because it changed the repo toolchain floor.
- **Problems encountered**:
  - Problem: `go get github.com/coder3101/tree-sitter-proto@d65a18ce7c22` upgraded `go.mod` to `go 1.26.0`.
  - Resolution: Removed the module dependency, restored `go 1.24.5`, and vendored the parser sources with a local wrapper.
  - Retry count: 1
- **Validation**: `go test ./internal/lang ./internal/language` → exit 0
- **Files changed**:
  - `internal/lang/lang.go` — added `protobuf` to supported languages.
  - `internal/language/language.go` — registered Protobuf config and local grammar loader.
  - `internal/language/queries.go` — added Protobuf symbol extraction queries.
  - `cmd/help_kind_support.go` — documented Protobuf kind coverage.
  - `internal/grammars/protobuf/protobuf.go` — added local cgo wrapper.
  - `internal/grammars/protobuf/src/*` — vendored third-party parser artifacts.
- **Next step**: Milestone 3 — Add Protobuf fixtures and parser tests

---

## Milestone 3: Add Protobuf fixtures and parser tests

- **Status**: DONE
- **Started**: 22:36
- **Completed**: 22:41
- **What was done**:
  - Added parser-level coverage for Protobuf symbol extraction and `.proto` language detection.
  - Added fixture coverage for Protobuf `symbol` and `definition` flows across `struct`, `enum`, `interface`, and `method`.
- **Key decisions**:
  - Decision: Use one symbol fixture with all supported Protobuf kinds plus one definition fixture per advertised kind.
  - Reasoning: This satisfies the fixture matrix while keeping test data compact.
  - Alternatives considered: Duplicating symbol fixtures per kind was unnecessary once coverage was already explicit.
- **Problems encountered**:
  - Problem: None.
  - Resolution: Not applicable.
  - Retry count: 0
- **Validation**: `go test ./cmd ./internal/language -timeout 60s` → exit 0
- **Files changed**:
  - `internal/lang/lang_test.go` — asserted `protobuf` appears in supported languages.
  - `internal/language/language_test.go` — added Protobuf parser and detection tests.
  - `tests/protobuf/...` — added CLI fixture coverage for Protobuf.
- **Next step**: Milestone 4 — Run repository validation gates

---

## Milestone 4: Run repository validation gates

- **Status**: DONE
- **Started**: 22:41
- **Completed**: 22:44
- **What was done**:
  - Ran the full repository test suite with a 60-second Go timeout.
  - Ran `golangci-lint` across the repository.
  - Confirmed `go run . lang list` now shows `protobuf`.
- **Key decisions**:
  - Decision: Add a lightweight command-level check after test and lint.
  - Reasoning: It directly confirms the user-visible language registry changed as intended.
  - Alternatives considered: Relying only on unit tests would have been sufficient technically, but less direct for the CLI-facing claim.
- **Problems encountered**:
  - Problem: None.
  - Resolution: Not applicable.
  - Retry count: 0
- **Validation**: `go test ./... -timeout 60s && golangci-lint run ./...` → exit 0
- **Files changed**:
  - `.codex-tasks/20260405-protobuf-language-support/TODO.csv` — closed final milestone.
  - `.codex-tasks/20260405-protobuf-language-support/PROGRESS.md` — recorded final validation and recovery state.
- **Next step**: None

---

## Milestone 1: Update Task Scope For External Grammar Migration

- **Status**: DONE
- **Started**: 00:04
- **Completed**: 00:04
- **What was done**:
  - Rewrote the task specification to cover external grammar migration, zig investigation, and Solidity cleanup.
  - Reset `TODO.csv` to track the new four-step execution plan.
- **Key decisions**:
  - Decision: Continue in the same task directory instead of starting a second task.
  - Reasoning: The new work is a direct continuation of the just-finished Protobuf support task.
  - Alternatives considered: Creating a new task directory would have duplicated context and validation history.
- **Problems encountered**:
  - Problem: None.
  - Resolution: Not applicable.
  - Retry count: 0
- **Validation**: `test -f .codex-tasks/20260405-protobuf-language-support/SPEC.md && test -f .codex-tasks/20260405-protobuf-language-support/TODO.csv` → exit 0
- **Files changed**:
  - `.codex-tasks/20260405-protobuf-language-support/SPEC.md` — expanded scope and done-when criteria.
  - `.codex-tasks/20260405-protobuf-language-support/TODO.csv` — replaced the completed plan with the new migration plan.
  - `.codex-tasks/20260405-protobuf-language-support/PROGRESS.md` — updated recovery state.
- **Next step**: Milestone 2 — Determine whether zig can drop vendor safely

---

## Milestone 2: Determine Whether Zig Can Drop Vendor Safely

- **Status**: SUPERSEDED
- **Started**: 00:04
- **Completed**: 00:14
- **What was done**:
  - Verified the published `github.com/maxxnino/tree-sitter-zig/bindings/go` artifact only contains `binding.go`, `binding_test.go`, `go.mod`, and `LICENSE`.
  - Verified the submodule `go.mod` declares `module github.com/tree-sitter/tree-sitter-zig`, which does not match the published import path under `github.com/maxxnino/...`.
  - Reproduced the failure when importing `github.com/maxxnino/tree-sitter-zig/bindings/go` in a temporary module.
  - Reproduced Swift root-module failure because the published cache is missing `src/parser.c`.
- **Key decisions**:
  - Decision: Initial conclusion was to keep `zig` vendored unless we controlled a repaired fork.
  - Reasoning: The published Go module is unusable without patching upstream because it both misdeclares the module path and omits parser sources from the distributed submodule.
  - Alternatives considered: `replace` from `github.com/tree-sitter/tree-sitter-zig` to `github.com/maxxnino/tree-sitter-zig` was rejected because the replaced parent module still does not contain `bindings/go`.
- **Problems encountered**:
  - Problem: Zig publishes a broken Go binding layout for direct consumption.
  - Resolution: Document the root cause and keep the existing vendored grammar.
  - Retry count: 0
- **Validation**: `go mod download -json github.com/maxxnino/tree-sitter-zig/bindings/go@latest` → exit 0, plus temporary-module import reproduction failed as expected
- **Files changed**:
  - `.codex-tasks/20260405-protobuf-language-support/TODO.csv` — advanced milestone state.
  - `.codex-tasks/20260405-protobuf-language-support/PROGRESS.md` — recorded the zig and swift upstream packaging findings.
- **Next step**: Fork and repair the affected upstream grammars so vendor removal becomes possible

---

## Milestone 3: Migrate Supported Vendored Grammars To External Modules

- **Status**: SUPERSEDED
- **Started**: 00:14
- **Completed**: 00:18
- **What was done**:
  - Switched Protobuf to `github.com/coder3101/tree-sitter-proto/bindings/go`.
  - Switched TypeScript/TSX to `github.com/tree-sitter/tree-sitter-typescript/bindings/go`.
  - Upgraded the repository `go` directive to `1.26.0` and added the root module requirements needed by those imports.
  - Removed vendored Protobuf and TypeScript grammar files from `internal/grammars/`.
- **Key decisions**:
  - Decision: Initial conclusion was to keep `swift` vendored unless we controlled a repaired fork.
  - Reasoning: Although the Swift root module contains `bindings/go`, the published module cache is missing `src/parser.c`, so cgo compilation fails.
  - Alternatives considered: Importing the Swift submodule directly was rejected because its published submodule declares a different module path (`github.com/tree-sitter/tree-sitter-swift`) than the available repository path.
- **Problems encountered**:
  - Problem: Some upstream grammar repositories publish module layouts that work in GitHub source view but fail as Go modules.
  - Resolution: Migrate only the grammars whose published module artifacts are actually consumable.
  - Retry count: 0
- **Validation**: `go test ./internal/lang ./internal/language ./cmd -timeout 60s` → exit 0
- **Files changed**:
  - `go.mod` / `go.sum` — added external grammar dependencies and raised the Go version.
  - `internal/language/language.go` — switched Protobuf and TypeScript loaders to external modules.
  - `internal/grammars/protobuf/*` — removed vendored Protobuf grammar files.
  - `internal/grammars/typescript/*` — removed vendored TypeScript grammar files.
- **Next step**: Fork and repair the affected upstream grammars, then finish full vendor removal

---

## Milestone 4: Remove Unsupported Solidity Vendor And Run Repository Gates

- **Status**: SUPERSEDED
- **Started**: 00:18
- **Completed**: 00:20
- **What was done**:
  - Removed `internal/grammars/solidity/` entirely.
  - Confirmed only `swift` and `zig` remain under `internal/grammars/`.
  - Ran the full repository test suite and `golangci-lint`.
- **Key decisions**:
  - Decision: Temporary close-out before the user requested full fork-based vendor removal.
  - Reasoning: At that point the published upstream modules were still not directly consumable.
  - Alternatives considered: None beyond the temporary stop state.
- **Problems encountered**:
  - Problem: None during final validation.
  - Resolution: Not applicable.
  - Retry count: 0
- **Validation**: `go test ./... -timeout 60s && golangci-lint run ./...` → exit 0
- **Files changed**:
  - `internal/grammars/solidity/*` — removed unsupported vendored grammar files.
  - `.codex-tasks/20260405-protobuf-language-support/TODO.csv` — closed the final milestone.
  - `.codex-tasks/20260405-protobuf-language-support/PROGRESS.md` — recorded final validation.
- **Next step**: Re-open the task and complete fork-based cleanup

---

## Milestone 2 Follow-up: Fork And Repair Swift/Zig Upstream Go Packaging

- **Status**: DONE
- **Started**: 00:23
- **Completed**: 00:38
- **What was done**:
  - Forked `alex-pinkus/tree-sitter-swift` to `XDwanj/tree-sitter-swift`.
  - Forked `maxxnino/tree-sitter-zig` to `XDwanj/tree-sitter-zig`.
  - Patched the Swift fork root `go.mod` to use module path `github.com/XDwanj/tree-sitter-swift`, upgraded it to `go 1.26.0`, fixed Go tests to import the fork path, and added the missing `src/parser.c` and `src/tree_sitter/parser.h`.
  - Patched the Zig fork to publish a root Go module `github.com/XDwanj/tree-sitter-zig`, removed the broken nested `bindings/go/go.mod`, upgraded tests to `github.com/tree-sitter/go-tree-sitter`, and fixed imports to the fork path.
  - Pushed Swift commit `99c92aa3b350948da01eaa207824443908eb3b5f` and Zig commit `52de00bd38c51489f26bd85964563f40be677edc`.
- **Key decisions**:
  - Decision: Fix the module layouts in personal forks instead of keeping local vendor files.
  - Reasoning: This preserves the no-vendor goal while avoiding brittle local `replace` hacks in `gx`.
  - Alternatives considered: Continuing to vendor Swift/Zig was rejected after the user explicitly asked to remove the vendor directory.
- **Problems encountered**:
  - Problem: `gh` initially picked up an invalid `GITHUB_TOKEN` from the environment.
  - Resolution: Ran `gh` commands with `env -u GITHUB_TOKEN` so the existing keyring login could be used.
  - Retry count: 0
- **Validation**: `cd /tmp/gx-swift-fork-zeH01D && GOFLAGS='' go test ./...` and `cd /tmp/gx-zig-fork-w1J4nB && GOFLAGS='' go test ./...` → exit 0
- **Files changed**:
  - External fork `XDwanj/tree-sitter-swift` — repaired Go packaging and added missing parser sources.
  - External fork `XDwanj/tree-sitter-zig` — repaired Go module layout and tests.
- **Next step**: Migrate `gx` to the repaired fork modules and remove the remaining vendor files

---

## Milestone 3 Follow-up: Migrate Gx To Forked Swift/Zig Modules And Remove Internal/Grammars

- **Status**: DONE
- **Started**: 00:38
- **Completed**: 00:46
- **What was done**:
  - Switched `internal/language/language.go` to import `github.com/XDwanj/tree-sitter-swift/bindings/go` and `github.com/XDwanj/tree-sitter-zig/bindings/go`.
  - Added the two fork modules to `go.mod`.
  - Deleted vendored Swift and Zig grammar sources.
  - Removed the last leftover files under `internal/grammars/`, including `.gx-ignore`.
- **Key decisions**:
  - Decision: Depend on fork root modules with subpackage imports.
  - Reasoning: Both repaired forks now publish a root module that contains the `bindings/go` package plus parser sources.
  - Alternatives considered: Publishing nested Go submodules again was unnecessary once the root module packaging worked.
- **Problems encountered**:
  - Problem: Swift build emits a `TOKEN_COUNT` macro redefinition warning from upstream parser/scanner sources.
  - Resolution: Left the warning visible because builds and tests still pass, and suppressing it would violate the current Debug-First preference.
  - Retry count: 0
- **Validation**: `go test ./internal/lang ./internal/language ./cmd -timeout 60s` → exit 0
- **Files changed**:
  - `go.mod` / `go.sum` — added `XDwanj/tree-sitter-swift` and `XDwanj/tree-sitter-zig`.
  - `internal/language/language.go` — switched Swift and Zig loaders to fork modules.
  - `internal/grammars/*` — fully removed.
- **Next step**: Run full repository validation and command-level checks

---

## Milestone 4 Follow-up: Run Repository Validation Gates

- **Status**: DONE
- **Started**: 00:46
- **Completed**: 00:49
- **What was done**:
  - Ran the full repository test suite.
  - Ran `golangci-lint`.
  - Confirmed `gx lang list` still shows `protobuf`, `swift`, `typescript`, and `zig`.
- **Key decisions**:
  - Decision: Keep the fork module versions pinned to the exact pushed commits.
  - Reasoning: The fork fixes are now part of the build contract and should stay reproducible.
  - Alternatives considered: Depending on `latest` was rejected because it would make the migration non-deterministic.
- **Problems encountered**:
  - Problem: None.
  - Resolution: Not applicable.
  - Retry count: 0
- **Validation**: `go test ./... -timeout 60s && golangci-lint run ./...` → exit 0
- **Files changed**:
  - `.codex-tasks/20260405-protobuf-language-support/TODO.csv` — finalized milestone states.
  - `.codex-tasks/20260405-protobuf-language-support/PROGRESS.md` — recorded final migration details.
- **Next step**: None

---

## Final Summary

- **Total milestones**: 4
- **Completed**: 4
- **Failed + recovered**: 0
- **External unblock events**: 0
- **Total retries**: 0
- **Files created**: 0
- **Files modified**: 12
- **Key learnings**:
  - Published Go-module artifacts for Tree-sitter grammars can diverge materially from the GitHub repository layout, so temporary-module compile checks are a better gate than repository browsing alone.
  - When upstream packaging is broken but the repository layout is otherwise sound, a repaired fork can remove local vendor code cleanly without changing `gx` runtime behavior.
- **Recommendations for future tasks**:
  - None.

## Final Summary

- **Total milestones**: 4
- **Completed**: 4
- **Failed + recovered**: 1
- **External unblock events**: 0
- **Total retries**: 1
- **Files created**: 17
- **Files modified**: 10
- **Key learnings**:
  - Vendoring a third-party Tree-sitter grammar can preserve repository toolchain constraints better than a direct module import when the upstream module has a higher `go` directive.
  - A compact symbol fixture plus per-kind definition fixtures is enough to satisfy the language-level coverage matrix cleanly.
- **Recommendations for future tasks**:
  - None.
