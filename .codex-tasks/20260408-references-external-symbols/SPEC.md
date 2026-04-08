# references external symbol usages

## Goal

Make `gx references --name <symbol>` return in-project usages for external dependency symbols even when the current project index has no local declaration for that symbol.

## Scope

- Update only the `references` query path.
- Keep `symbols` and `definition` behavior unchanged.
- Add regression coverage for a Go file that only imports and calls an external package symbol.

## Validation

- `go test ./internal/query -run TestReferencesFindsExternalSymbolUsages -timeout 60s`
- `go test ./... -timeout 60s`
