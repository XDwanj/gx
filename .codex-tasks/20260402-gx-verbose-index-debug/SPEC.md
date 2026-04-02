# gx verbose index debug

## Goal

Add a global `--verbose` flag that exposes indexing and query progress, then use
that visibility to diagnose and fix why `gx o cmd/definition.go` appears to
hang with no output.

## Scope

- Add a global verbose flag on the root command.
- Emit debug-first progress logs for root resolution, cache usage, index build
  stages, and slow file processing.
- Reproduce the hanging behavior with verbose output and fix the root cause.
- Add or update automated tests.

## Constraints

- Preserve normal non-verbose output behavior.
- Keep Cobra wiring in `cmd/` and business logic in `internal/`.
- Do not add silent fallbacks; expose failures and progress explicitly.
