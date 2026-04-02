# Makefile and Cross Compilation Support

## Goal

Add a project Makefile that standardizes local build, test, lint, clean, and cross-compilation artifact generation for `gx`.

## Constraints

- Keep the public binary name `gx`.
- Respect the repository's existing Go test and lint commands.
- Surface real cross-compilation toolchain requirements instead of silently skipping targets.
- Support cgo-based builds used by tree-sitter grammars.

## Acceptance Criteria

- A `Makefile` exists with `build`, `test`, `lint`, `clean`, and cross-compilation targets.
- Cross-compilation artifacts are written under `dist/`.
- The Makefile explicitly handles cgo compiler requirements for cross-compiling.
- `.gitignore` ignores generated build artifacts.
- README documents the new build workflow and cross-compilation requirements.
- `make build`, `go test ./... -timeout 60s`, and `golangci-lint run ./...` pass.
