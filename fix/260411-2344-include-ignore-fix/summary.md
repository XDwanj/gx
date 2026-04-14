# Summary

- Baseline error count: 3
- Final error count: 0
- Categories fixed:
  - include now temporarily re-indexes matching ignored files for the current query
  - normal queries now hide paths according to the current on-disk `.gitignore` and `.ignore` state instead of relying on stale cached visibility
  - `--json` symbol queries now emit `[]` for empty results so machine consumers can distinguish "no matches" without parsing stderr
- Blocked items:
  - the exact launch-manifest `go run` verify path is externally blocked by a local Go toolchain mismatch
- Guard status: passed via `go test ./... -timeout 60s` and `golangci-lint run ./...`
