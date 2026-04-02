# gx — Semantic Code Navigation

Use `gx` to narrow the target before reading a full file.

## Examples

- See project entrypoints and top-level structure: `gx overview .`
- Inspect a directory before opening files: `gx overview internal/query`
- Inspect file structure without reading the whole file: `gx overview cmd/root.go`
- Find functions or types across the project: `gx symbols --kind fn --name '*search*'`
- `gx --name` supports glob patterns such as `prefix*`, `*suffix`, and `*contains*`
- Limit symbol search to one file: `gx symbols --scope internal/query/search.go`
- Read one function or type directly: `gx definition --name Search`
- Resolve a definition from a specific file: `gx definition --name Search --scope internal/query/search.go`
- See where a symbol is used: `gx references --name Search`
- Check refactor blast radius first: `gx references --name Search --unique`

## Workflow

`overview -> symbols -> definition/references -> Read`

Only fall back to Read when you need full-file context.

## Quick Reference

```text
gx overview PATH
gx overview DIR --full
gx symbols [--kind K] [--name GLOB] [--scope PATH]
gx definition --name GLOB [--scope PATH] [--kind K]
gx references --name GLOB [--scope PATH] [--unique]
gx lang list
gx lang add LANG [LANG...]
```

Aliases: `gx o`, `gx s`, `gx d`, `gx r`

Kinds: `fn`, `method`, `struct`, `enum`, `trait`, `type`, `const`, `class`, `interface`, `module`, `event`

Match mode: every `--name` filter uses glob matching, and `--scope` accepts either a file path or a directory path.

## Missing Grammars

Check installed grammars: `gx lang list`

Install what is missing: `gx lang add rust`
