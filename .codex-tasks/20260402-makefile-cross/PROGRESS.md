# Progress

- 2026-04-02: Inspected current build flow, verified local build, and confirmed tree-sitter grammar packages require cgo for cross-compilation.
- 2026-04-02: Verified local macOS cross-compilation to `darwin/amd64` works with `clang -arch x86_64`.
- 2026-04-02: Added `Makefile` targets for local build, checks, cleanup, and cross-compilation artifact generation.
- 2026-04-02: Updated `.gitignore`, `README.md`, and `README.zh-CN.md` to document the new workflow and cgo toolchain requirements.
- 2026-04-02: Validated `make build`, `make cross-darwin`, `go test ./... -timeout 60s`, and `golangci-lint run ./...`.
