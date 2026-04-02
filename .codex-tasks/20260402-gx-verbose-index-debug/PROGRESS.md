# Progress

- 2026-04-02: Created task artifacts and confirmed current behavior: installed
  Go grammar, empty cache issue previously fixed, but `gx o cmd/definition.go`
  still times out before output because indexing scans large vendored/generated
  grammar trees.
- 2026-04-02: Added global `--verbose` support and index-stage debug logs for
  root resolution, cache path, cache reuse/rebuild decisions, per-file indexing,
  and full-crawl progress.
- 2026-04-02: Reproduced the timeout with `./gx --verbose=true o
  cmd/definition.go`; logs showed the indexer spending most of its time inside
  `internal/grammars/**`, especially large generated `parser.c` files.
- 2026-04-02: Added `internal/grammars/.gx-ignore` so the existing walker skips
  vendored/generated grammar trees during indexing. Command now completes
  quickly and returns symbols for `cmd/definition.go`.
- 2026-04-02: Verification passed with `go test ./... -timeout 60s`,
  `golangci-lint run ./...`, and a fresh-cache run of `./gx --verbose=true o
  cmd/definition.go`.
