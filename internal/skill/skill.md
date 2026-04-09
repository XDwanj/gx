---
name: gx
description: "ALWAYS activate this skill for any task that requires reading, understanding, locating, tracing, reviewing, debugging, or modifying an existing codebase. `gx` is the default first tool for semantic code navigation and must be used before text-based tools for symbol structure, definitions, references, package layout, and architecture exploration. Only do not activate it for non-code files or pure literal text search such as logs, docs, comments, YAML, JSON, SQL, and Markdown."
---

# GX

`gx` is a semantic code navigation tool for AI agents. Use it when the task is about code structure, declarations, definitions, references, or package layout. Do not use it for plain text lookup in logs, comments, Markdown, YAML, JSON, SQL, or arbitrary strings.

## Default Workflow

Use the usual narrowing loop:

1. `overview` for package or file surface.
2. `symbols` to find the exact declaration.
3. `definition` to read the implementation body.
4. `references` to inspect impact.

Default examples:

```bash
gx overview internal/tmdb
gx symbols --kind method --name 'Search' .
gx definition --name 'Search' internal/tmdb
gx references --unique --name 'Search' .
```

## Kinds And Coverage

Public `--kind` values:

- `fn`
- `method`
- `const`
- `struct`
- `enum`
- `class`
- `interface`
- `module`
- `type`

Rules:

- Treat these as `gx` extraction kinds, not as a promise that every language supports every kind.
- `symbols` and `definition` support `--kind`; `references` does not.
- Narrow by `--name` first, then add `--kind` to reduce noise.
- If a `--kind` query returns nothing, retry without `--kind`.

Current language coverage:

- `bash`: `fn`
- `c`: `fn`, `struct`, `enum`, `type`
- `cpp`: `fn`, `method`, `struct`, `class`, `enum`, `module`, `type`
- `go`: `fn`, `method`, `const`, `struct`, `interface`, `type`
- `java`: `class`, `method`, `const`, `enum`, `interface`, `module`
- `lua`: `fn`, `method`
- `python`: `fn`, `const`, `class`
- `protobuf`: `struct`, `enum`, `interface`, `method`
- `ruby`: `method`, `class`, `module`
- `rust`: `fn`, `method`, `const`, `struct`, `enum`, `interface`, `module`, `type`
- `swift`: `fn`, `method`, `const`, `struct`, `enum`, `class`, `interface`, `module`, `type`
- `typescript`: `fn`, `method`, `const`, `class`, `enum`, `interface`, `module`, `type`
- `zig`: `fn`, `struct`, `enum`

## Command Guide

### `gx overview`

Use `overview` before reading code.

- Accepts multiple paths and returns one section per target when more than one path is supplied.
- Directory mode returns per-file symbol summaries.
- File mode returns top-level symbols for that file.
- Directory mode supports `--limit`, `--offset`, and `--all`.
- In multi-path mode, pagination applies independently to each directory target.
- File mode and Markdown outline mode ignore pagination flags.

Examples:

```bash
gx overview internal/tmdb
gx overview internal/tmdb/search.go
gx overview --full internal/tmdb
gx overview internal/tmdb internal/tmdb/search.go README.md
```

### `gx symbols`

Use `symbols` to find declarations.

- Returns definitions, not usages.
- Accepts files, directories, or a mix.
- `--name` uses glob matching.
- Default result limit is `100`.
- Use `--limit`, `--offset`, and `--all` for broad matches.

Examples:

```bash
gx symbols --name '*Search*' internal/tmdb
gx symbols --kind method --name 'Search' .
gx symbols --name '*Search*' --limit 20 --offset 20 internal/tmdb
```

### `gx definition`

Use `definition` when you already know the symbol and want the body immediately.

- Returns the symbol body with source file and line.
- Accepts files, directories, or a mix.
- Default result limit is `5`.
- `--limit` controls how many definitions you get.
- `--max-lines` controls how large each definition body can be.

Examples:

```bash
gx definition --name 'Search' internal/tmdb
gx definition --name 'Client' internal/tmdb/client.go
gx definition --name '*Search*' --limit 1 --max-lines 80 internal/tmdb
```

### `gx references`

Use `references` to find usages and impact.

- Commonly includes the definition site as well.
- Accepts files, directories, or a mix.
- `--unique` deduplicates by enclosing function.
- Default result limit is `50`.
- Use `--limit`, `--offset`, and `--all` when a common name fans out widely.

Examples:

```bash
gx references --name 'Search' .
gx references --unique --name 'Search' .
gx references --name 'Search' --limit 25 --offset 25 .
```

### `gx cache`

Use `cache` when the index seems stale.

```bash
gx cache path
gx cache clean
```

### `gx lang`

Use `lang` to inspect or manage installed grammars.

```bash
gx lang list
gx lang enable go typescript
gx lang disable ruby
```

## `--name` Matching Rules

Assume `--name` uses glob-style matching.

- Matching is case-sensitive.
- `*` works as a wildcard.
- Brace alternation uses commas, not pipes.
- Quote the pattern so the shell does not expand it first.

Good patterns:

```bash
gx symbols --name 'Search' .
gx symbols --name '*Search*' .
gx symbols --name '{Ali*Pay,Wechat*Pay}' .
gx symbols --name '*{Ali*Pay,Wechat*Pay}*' .
```

Bad patterns:

```bash
gx symbols --name '{AliPay|WechatPay}' .
gx symbols --name {AliPay,WechatPay} .
```

Why they are bad:

- `|` is not the alternation syntax `gx` accepts here.
- Unquoted braces are expanded by the shell before `gx` sees them.

## Path Rules

- `overview`, `symbols`, `definition`, and `references` accept multiple paths.
- Those paths may be directories, files, or a mix.
- Prefer the smallest scope that answers the question.
- Use `-C` when you need to run `gx` against another directory context.

Examples:

```bash
gx overview internal/tmdb internal/tmdb/search.go README.md
gx -C /path/to/project symbols --name 'Search' .
gx symbols --name 'Search' internal/tmdb/search.go internal/tools
gx definition --name 'Search' internal/tmdb/search.go
gx references --unique --name 'Search' internal/tmdb/search.go internal/tools
```

## Output Mode

Use default output for terminal reading. Use `--json` when another tool or script will consume the result, or when an agent needs structured fields.

Examples:

```bash
gx definition --json --name 'Search' internal/tmdb
gx references --json --unique --name 'Search' .
```

## Practical Heuristics

- Start from a package directory, not the repo root, when using `overview`.
- Use `symbols` before `definition` if the name or `kind` may be ambiguous.
- Treat `--kind` as a precision filter, not as the primary lookup mechanism.
- When results are unexpectedly empty, check help output before assuming the symbol is absent.
- Use `references --unique` for quick impact analysis.
- Narrow the query before reaching for `--all`.
- Use `--offset` to move forward through a large result set.
- Expect a compact paging hint on `stderr` when results are truncated.

## Response Style

When helping the user with `gx`:

- explain the command sequence in terms of intent, not just syntax
- prefer the smallest scope that answers the question
- explicitly point out when the task is not a good fit for `gx`
- call out shell quoting when patterns use braces or wildcards
- do not turn the explanation into a comparison guide for other search tools
