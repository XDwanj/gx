# Epic Specification

## Goal

- Rebuild `cx` as a Go CLI in this repository, using Cobra, with behavior kept as close to 1:1 as practical with the Rust implementation.

## Non-Goals

- Shipping a partial scaffold without working commands
- Introducing fallback-only behavior that hides migration gaps
- Expanding scope beyond the current `cx` feature set unless required for compatibility

## Constraints

- Implementation language must be Go
- CLI framework must be Cobra
- Behavior should match the Rust `cx` command surface and output semantics as closely as practical
- Keep failures explicit and visible during migration
- Prefer maintainable package boundaries so the Go port remains testable

## Risk Assessment

- Tree-sitter grammar loading and query support in Go may differ from the Rust implementation
- Output compatibility is sensitive for TOON, JSON, and plain-text definition rendering
- Incremental index persistence and locking behavior must be re-created carefully
- Cross-language parsing support increases migration surface area

## Child Deliverables

- Bootstrap the Go module, package layout, Cobra command tree, and shared data model
- Port language registry, symbol extraction, reference scanning, and grammar management
- Port index persistence, cache handling, directory walking, and incremental rebuild logic
- Port user-facing commands and reach a verified usable state with tests and command checks

## Dependency Notes

- Child 2 depends on child 1
- Child 3 depends on child 2
- Child 4 depends on child 2 and child 3
- `depends_on` uses `;` as delimiter for multiple IDs

## Child Task Types

- `single-full`
- `batch`

## Done-When

- [ ] Every row in `SUBTASKS.csv` is `DONE`
- [ ] `go test ./...` passes
- [ ] Core commands run successfully in `gx`
- [ ] Behavior has been checked against the Rust `cx` implementation for key flows
