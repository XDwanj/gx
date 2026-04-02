# Local Agent Rules

- Default to Chinese in user-facing replies unless the user explicitly requests another language.
- For this project, the public command name must be `gx`. Mention `cx` only when explaining project origin or upstream source, never as the primary invocation shown to users.
- When diagnosing indexing failures, check grammar install state and the current cache contents together; do not conclude "missing grammar" without ruling out a stale empty cache.
- `golangci-lint run ./...` is a hard gate for this repository. Do not leave lint failures unresolved; either fix them or add explicit config exclusions for generated or third-party code with a clear reason.
- Before changing code, read the relevant entrypoints and existing tests first; keep Cobra command wiring in `cmd/` and business logic in `internal/`.
- Prefer extending existing `internal/app`, `internal/query`, `internal/language`, and `internal/output` flows instead of adding parallel abstractions.
- Preserve output behavior across both human-readable and `--json` modes; do not add formatter-specific logic directly in Cobra handlers.
- Follow Debug-First: do not add silent fallbacks, fake success paths, hidden caps, or error-swallowing branches just to keep the command running.
- When behavior changes, add or update automated tests where practical. When running Go tests, use Go's built-in hard timeout: `go test ./... -timeout 60s`.
- Treat files under `internal/grammars/` such as `parser.c`, `grammar.json`, `node-types.json`, `tree_sitter/*`, `scanner.c`, `scanner.h`, and `binding.cc` as generated or third-party sources. Do not edit them unless the task explicitly requires grammar regeneration or vendor updates.
- Keep path handling explicit and normalized; prefer existing helpers for project-root and relative-path resolution before introducing new path logic.
- For short user-facing guidance docs like embedded `skill.md`, prefer example-first structure over introductory explanation; keep conceptual framing to the minimum needed.
