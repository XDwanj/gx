# Task Specification

> Scope anchor for the task. Update only when goals or constraints change, and log the reason in PROGRESS.md.

## Task Shape

- **Shape**: `single-full`

## Goals

- Add ripgrep-style `.ignore` file support to index traversal so ignored files and directories are excluded from `gx` indexing.
- Support directory-scoped `.gitignore` files alongside `.ignore` during index traversal.
- Preserve existing `.gx-ignore` directory skipping.
- Cover the new filtering behavior with automated tests and document the supported ignore sources.

## Non-Goals

- Do not add `.rgignore`, global gitignore, or parent-directory ignore discovery above the project root.
- Do not change language detection, grammar installation, or output formatting behavior.

## Constraints

- Keep `gx` as the public command name in docs and user-facing text.
- Follow Debug-First behavior: no silent fallback paths or fake success states.
- Do not edit generated or third-party sources under `internal/grammars/`.

## Environment

- **Project root**: `/Users/xdwanj/Project/Rust/gx`
- **Language/runtime**: Go
- **Package manager**: Go modules
- **Test framework**: `go test`
- **Build command**: `go test ./... -timeout 60s`
- **Existing test count**: `36`

## Risk Assessment

- [x] External dependencies (APIs, services) — availability confirmed?
- [x] Breaking changes to existing code — impact assessed?
- [x] Large file generation — disk space sufficient?
- [x] Long-running tests — timeout configured?

## Deliverables

- Updated ignore traversal in `internal/index/index.go`
- Tests covering `.ignore` support and interaction with existing ignore mechanisms
- README updates describing `.ignore` support

## Done-When

- [ ] `walk` applies root `.gitignore`, layered `.ignore`, and directory `.gx-ignore` correctly during indexing.
- [ ] `walk` applies directory-scoped `.gitignore`, layered `.ignore`, and directory `.gx-ignore` correctly during indexing.
- [ ] New tests verify root and nested `.ignore` behavior without regressing existing ignore behavior.
- [ ] Documentation mentions `.ignore` alongside existing ignore mechanisms.
- [ ] `go test ./... -timeout 60s` and `golangci-lint run ./...` both pass.

## Final Validation Command

```bash
go test ./... -timeout 60s && golangci-lint run ./...
```

## Demo Flow (optional)

1. Add `.ignore` files to a project tree.
2. Run a `gx` query command that loads or rebuilds the index.
3. Confirm ignored files do not appear in the indexed results.
