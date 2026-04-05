# Task Specification

> Scope anchor for the task. Update only when goals or constraints change, and log the reason in PROGRESS.md.

## Task Shape

- **Shape**: `single-full`

## Goals

- Keep the current public `kind` set unchanged while mapping Protobuf declarations onto existing kinds.
- Prefer external grammar module dependencies over vendored grammar sources where the upstream Go module is directly consumable.
- Migrate `protobuf`, `typescript`, and `swift` away from vendored grammars, and migrate `zig` too if the upstream module can be consumed safely.
- Remove `solidity` vendored grammar remnants because `gx` no longer supports Solidity.
- Preserve automated coverage across parser-level and fixture-level flows after the migration.

## Non-Goals

- Do not add new public kinds or remove existing ones.
- Do not expose every Protobuf syntax form as a symbol in the first pass.
- Do not keep vendored grammars only for convenience when a stable external module path exists.

## Constraints

- Use the public command name `gx` in messages and user-facing text.
- Keep Cobra wiring in `cmd/` and business logic in `internal/`.
- Preserve behavior across both human-readable and `--json` modes.
- Follow Debug-First: no silent fallbacks, fake success paths, or swallowed errors.
- Use existing upstream grammar packages; do not create a new grammar implementation.
- The user explicitly accepts raising the repository Go version if required by upstream grammar modules.

## Environment

- **Project root**: `/Users/xdwanj/Project/Rust/gx`
- **Language/runtime**: `Go 1.24.5+`
- **Package manager**: `go modules`
- **Test framework**: `go test`
- **Build command**: `go build ./...`
- **Existing test count**: `15`

## Risk Assessment

- [x] External dependencies (APIs, services) — availability confirmed?
- [x] Breaking changes to existing code — impact assessed?
- [x] Large file generation — disk space sufficient?
- [x] Long-running tests — timeout configured?

## Deliverables

- External grammar module integration for `protobuf`, `typescript`, and `swift`.
- A root-cause decision for `zig`: migrate to an external module if directly consumable, otherwise document why vendor must remain.
- Removal of `internal/grammars/solidity/`.
- Stable language registration, symbol extraction queries, and help matrix behavior after migration.
- Parser and fixture tests that continue to pass after dependency changes.

## Done-When

- [ ] `protobuf`, `typescript`, and `swift` no longer rely on vendored local grammar wrappers.
- [ ] `zig` is either migrated to a working external module path or explicitly retained with a documented root cause.
- [ ] `internal/grammars/solidity/` is removed.
- [ ] `gx` still recognizes `.proto` files and preserves the current Protobuf kind mapping.
- [ ] `go test ./... -timeout 60s` passes.
- [ ] `golangci-lint run ./...` passes.

## Final Validation Command

```bash
go test ./... -timeout 60s && golangci-lint run ./...
```

## Demo Flow (optional)

1. Run `gx lang list` and confirm `protobuf` is listed.
2. Run `gx symbols --json --kind struct --name 'HelloRequest' path/to/file.proto`.
3. Run `gx definition --json --kind interface --name 'Greeter' path/to/file.proto`.
