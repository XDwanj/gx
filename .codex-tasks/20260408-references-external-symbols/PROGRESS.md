# Progress

- Reviewed `cmd/references.go`, `internal/query/runtime.go`, and current references tests.
- Verified the current implementation exits early when `findMatchingSymbolNames` returns no local declarations.
- Added `language.FindReferenceNames` and reused the same reference-node traversal as `FindReferences`.
- Updated `references` to fall back to in-scope reference-name discovery when the index has no matching local declarations.
- Added a Go regression test covering an imported external symbol call with no local declaration.
- Validation:
  - `go test ./internal/query -run 'TestReferences(SupportsGlobName|SupportsPipeSeparatedAlternatives|PaginationWritesHint|ScopeFiltersResults|FindsExternalSymbolUsages)$' -timeout 60s`
  - `go test ./... -timeout 60s`
  - `go run . --root /home/xdwanj/Project/DADAO/Golang/backend/fix_wangjie_schedule_delete_device/advertising references --name SplitCutset .`
  - `golangci-lint run ./...` blocked before analysis because the installed binary was built with `go1.25.8` while the module targets `go1.26.0`.
