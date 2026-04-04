# Task Specification

> Scope anchor for the task. Update only when goals or constraints change, and log the reason in PROGRESS.md.

## Task Shape

- **Shape**: `single-full`

## Goals

- Make `gx symbols` return a declaration index with source coordinates by default.
- Persist symbol line data in the index so queries can reuse stable coordinates.
- Update tests, fixtures, and docs to match the new `symbols` output shape.

## Non-Goals

- Do not add `receiver` or `package` fields in this change.
- Do not change `definition` or `references` output contracts beyond consuming the new stored coordinates internally when useful.

## Constraints

- Keep Cobra wiring in `cmd/` and query logic in `internal/query`.
- Preserve Debug-First behavior and avoid hidden fallbacks.
- Treat this as an intentional breaking change for `symbols`.

## Environment

- **Project root**: `/Users/xdwanj/Project/Rust/gx`
- **Language/runtime**: `Go`
- **Package manager**: `Go modules`
- **Test framework**: `go test`
- **Build command**: `go test ./... -timeout 60s`
- **Existing test count**: `14 Go test files`

## Risk Assessment

- [x] External dependencies (APIs, services) — availability confirmed?
- [x] Breaking changes to existing code — impact assessed?
- [x] Large file generation — disk space sufficient?
- [x] Long-running tests — timeout configured?

## Deliverables

- Updated symbol extraction and index persistence with line metadata.
- Updated `symbols` query output and command behavior.
- Updated automated tests, JSON fixtures, and README docs.

## Done-When

- [ ] `gx symbols` default output includes `file` and `line`.
- [ ] Symbol coordinates are stored in the index and survive cache round-trips.
- [ ] Go tests and `golangci-lint run ./...` pass.

## Final Validation Command

```bash
go test ./... -timeout 60s && golangci-lint run ./...
```
