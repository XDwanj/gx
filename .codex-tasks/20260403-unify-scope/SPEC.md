# Task Specification

## Task Shape

- **Shape**: `single-full`

## Goals

- Unify `definition`, `symbols`, and `references` to use a single `--scope` flag.
- Make `--scope` support both file and directory paths with consistent hard-filter semantics.
- Update tests and user-facing docs to reflect the new scope behavior.

## Non-Goals

- Do not change non-path filters such as `--name`, `--kind`, or `--unique`.
- Do not add silent fallback behavior for invalid scope paths.

## Constraints

- Preserve existing user changes already present in the worktree.
- Follow the repo Debug-First policy: invalid scope paths must fail explicitly.
- Keep command wiring in `cmd/` and business logic in `internal/query/`.

## Environment

- **Project root**: `/Users/xdwanj/Project/Rust/gx`
- **Language/runtime**: Go
- **Package manager**: Go modules
- **Test framework**: `go test`
- **Build command**: `go test ./... -timeout 60s`
- **Existing test count**: `runtime_test.go` plus repository test suite

## Risk Assessment

- [x] External dependencies (APIs, services) — availability confirmed?
- [x] Breaking changes to existing code — impact assessed?
- [x] Large file generation — disk space sufficient?
- [x] Long-running tests — timeout configured?

## Deliverables

- Unified `--scope` CLI flags in `cmd/definition.go`, `cmd/symbols.go`, and `cmd/references.go`
- Shared scope filtering logic in `internal/query/`
- Updated tests and user-facing guidance

## Done-When

- [ ] All three commands accept `--scope` instead of `--from` / `--file`
- [ ] `--scope` supports both file and directory filtering consistently
- [ ] Automated tests cover file and directory scope behavior
- [ ] `go test ./... -timeout 60s` and `golangci-lint run ./...` pass

## Final Validation Command

```bash
go test ./... -timeout 60s && golangci-lint run ./...
```
