# SQLite Index Cache Migration

## Goal

Replace the current JSON index cache with a SQLite-backed implementation while preserving existing `gx` command behavior and output formats.

## Constraints

- Keep command entrypoints and query behavior unchanged.
- Preserve debug-first behavior: surface real storage errors.
- Avoid introducing silent fallbacks or mock behavior.
- Keep the implementation maintainable and small.

## Acceptance Criteria

- Index cache path now points to a SQLite database file.
- `LoadOrBuild` continues to load cached entries or rebuild when stale.
- `gx cache clean` removes the SQLite cache file.
- Tests cover cache path, load/save, and clean behavior.
- `go test ./... -timeout 60s` passes.
- `golangci-lint run ./...` passes.
