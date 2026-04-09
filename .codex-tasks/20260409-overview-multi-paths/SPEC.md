# Task Specification

> Scope anchor for the task. Update only when goals or constraints change, and log the reason in PROGRESS.md.

## Task Shape

- **Shape**: `single-full`

## Goals

- Make `gx overview` and `gx o` accept multiple path arguments.
- Keep the existing single-path output contract unchanged for directory, file, and Markdown inputs.
- Define and implement an explicit multi-path output contract for both default and `--json` modes.

## Non-Goals

- Do not change `symbols`, `definition`, or `references` behavior.
- Do not add compatibility aliases or silent fallbacks beyond the explicit multi-path contract.

## Constraints

- Keep Cobra wiring in `cmd/` and business logic in `internal/`.
- Follow Debug-First and keep failures visible.
- Synchronize README, embedded skill guidance, and tests if the public contract changes.

## Environment

- **Project root**: `/Users/xdwanj/Project/Rust/gx`
- **Language/runtime**: Go
- **Package manager**: Go modules
- **Test framework**: `go test`
- **Build command**: `go test ./... -timeout 60s`
- **Existing test count**: `<not counted>`

## Risk Assessment

- [x] External dependencies (APIs, services) — availability confirmed?
- [x] Breaking changes to existing code — impact assessed?
- [x] Large file generation — disk space sufficient?
- [x] Long-running tests — timeout configured?

## Deliverables

- `overview` command supports multiple positional paths.
- Automated coverage for multi-path `overview`.
- Updated public docs for the new contract.

## Done-When

- [ ] `gx overview path1 path2` works for supported path mixes and is covered by tests.
- [ ] Existing single-path behavior remains intact.
- [ ] `go test ./... -timeout 60s` and `golangci-lint run ./...` pass.

## Final Validation Command

```bash
go test ./... -timeout 60s && golangci-lint run ./...
```
