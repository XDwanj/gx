# Task Specification

> Scope anchor for the task. Update only when goals or constraints change, and log the reason in PROGRESS.md.

## Task Shape

- **Shape**: `single-full`

## Goals

- Replace the `internal/index` SQLite cache store implementation from `database/sql` to `sqlx`.
- Preserve the existing cache schema, file format, command behavior, and transaction boundaries.
- Keep ORM scope limited to the index cache persistence path and avoid changing Cobra handlers or output flows.

## Non-Goals

- Introduce a full ORM model layer or repository system across the entire project.
- Change cache file names, table layout, JSON payload shapes, or CLI output semantics.
- Modify generated grammar sources under `internal/grammars/`.

## Constraints

- Keep user-facing command name as `gx`.
- Follow Debug-First: no silent fallback paths, mock behavior, or hidden degradation.
- Preserve behavior across human-readable and `--json` modes by keeping formatter logic out of Cobra handlers.
- Keep path handling and store initialization explicit.

## Environment

- **Project root**: `/Users/xdwanj/Project/Rust/gx`
- **Language/runtime**: Go 1.24.5
- **Package manager**: Go modules
- **Test framework**: Go test
- **Build command**: `go test ./... -timeout 60s`
- **Existing test count**: not counted at task start

## Risk Assessment

- [x] External dependencies (APIs, services) — availability confirmed?
- [x] Breaking changes to existing code — impact assessed?
- [x] Large file generation — disk space sufficient?
- [x] Long-running tests — timeout configured?

## Deliverables

- Updated SQLite cache store implementation using `github.com/jmoiron/sqlx`.
- Tests covering unchanged persistence behavior after the migration.
- Passing repository validation commands.

## Done-When

- [ ] `internal/index` no longer uses `database/sql` directly for cache store access.
- [ ] Existing index persistence behavior remains covered by tests.
- [ ] `go test ./... -timeout 60s` passes.
- [ ] `golangci-lint run ./...` passes.

## Final Validation Command

```bash
go test ./... -timeout 60s && golangci-lint run ./...
```

