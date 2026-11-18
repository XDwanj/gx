---
name: gx
description: "Semantic code navigation tool. Use only when the user explicitly requests gx or asks to use semantic symbol navigation."
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
gx symbols --kind func --name 'Search' .
gx definition --name 'Search' internal/tmdb
gx references --unique --name 'Search' .
```

To answer "who does F call", run `gx definition` on F and read the body.

Use `--define-in FILE` when a common symbol name needs AI disambiguation against the definition in a specific file. This requires `GX_AI_API_KEY` and `GX_AI_BASE_URL`, accepts optional `GX_AI_MODEL` (set to `typesafe/jev-1.13` to use OpenRouter's decisions API instead of chat completions), and applies to `symbols`, `definition`, and `references`.

```bash
gx references --name 'login' --define-in internal/domain/user.go .
```

## Kinds And Coverage

Public `--kind` values:

- `func`
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

- `bash`: `func`
- `c`: `func`, `struct`, `enum`, `type`
- `cpp`: `func`, `struct`, `class`, `enum`, `module`, `type`
- `go`: `func`, `const`, `struct`, `interface`, `type`
- `java`: `class`, `func`, `const`, `enum`, `interface`, `module`
- `kotlin`: `func`, `const`, `class`, `enum`, `interface`, `type`
- `lua`: `func`
- `python`: `func`, `const`, `class`
- `protobuf`: `struct`, `enum`, `interface`, `func`
- `ruby`: `func`, `class`, `module`
- `rust`: `func`, `const`, `struct`, `enum`, `interface`, `module`, `type`
- `swift`: `func`, `const`, `struct`, `enum`, `class`, `interface`, `module`, `type`
- `typescript`: `func`, `const`, `class`, `enum`, `interface`, `module`, `type`
- `vue`: `module` for the top-level `template`/`script`/`style` blocks, plus `typescript` kinds extracted from `<script>` (requires `typescript` enabled). `references` also resolves `@`/`v-on:`/`:`/`v-*` attribute expressions and `{{ }}` interpolations.
- `wxml`: no symbols. `references` resolves `bind*`/`catch*`/`mut-bind`/`capture-*` event bindings and `{{ }}` expressions (requires `typescript` enabled).
- `zig`: `func`, `struct`, `enum`

## Command Guide

### `gx overview`

Use `overview` before reading code.

- Accepts multiple paths and returns one section per target when more than one path is supplied.
- Directory mode returns per-file symbol summaries.
- File mode returns top-level symbols for that file.
- Directory mode supports `--limit` and `--offset`.
- In multi-path mode, pagination applies independently to each directory target.
- File mode and Markdown outline mode ignore pagination flags.
- Markdown outline rows include `line`, `level`, and `heading`.
- When you know a field name but not the enclosing struct, class, or message name, use `overview` first to find candidate declarations.

Examples:

```bash
gx overview internal/tmdb
gx overview internal/tmdb/search.go
gx overview --full internal/tmdb
gx overview internal/tmdb internal/tmdb/search.go README.md
gx -C /path/to/project overview proto/api.proto
```

### `gx symbols`

Use `symbols` to find declarations.

- Returns definitions, not usages.
- Accepts files, directories, path globs, or a mix.
- `--include` and `--exclude` filter indexed file paths using glob matching. `--include` can temporarily bring back matching files hidden by `.gitignore` or `.ignore`.
- `--name` uses glob matching.
- Default result limit is `100`.
- Use `--limit <N|all>` and `--offset` for broad matches.

Examples:

```bash
gx symbols --name '*Search*' internal/tmdb
gx symbols --kind func --name 'Search' .
gx symbols --include 'internal/**' --exclude '**/*_test.go' --name '*Search*' .
gx symbols --name '*Search*' --limit 20 --offset 20 internal/tmdb
```

### `gx definition`

Use `definition` when you already know the symbol and want the body immediately.

- Returns `{file, line, body}` rows. `body` holds the symbol source; `truncated` and `lines` appear when `--max-lines` cuts the body.
- Accepts files, directories, path globs, or a mix.
- `--include` and `--exclude` filter indexed file paths using glob matching. `--include` can temporarily bring back matching files hidden by `.gitignore` or `.ignore`.
- Default result limit is `5`.
- `--limit` controls how many definitions you get.
- `--max-lines` controls how large each definition body can be.
- For member checks such as "does this struct or message contain field X", prefer `definition | rg` after you narrow to the enclosing declaration. This is especially useful in languages where fields are not indexed as standalone `symbols`.

Examples:

```bash
gx definition --name 'Search' internal/tmdb
gx definition --name 'Client' internal/tmdb/client.go
gx definition --include 'internal/**' --exclude '{**/*_test.go,**/mocks/**}' --name 'Search' .
gx definition --name '*Search*' --limit 1 --max-lines 80 internal/tmdb
gx definition --name 'CreateUserRequest' proto/api.proto -C /path/to/project | rg '\bemail\b'
gx definition --name 'User' internal/model/user.go | rg '\bEmail\b'
```

### `gx references`

Use `references` to find usages and impact.

- Commonly includes the definition site as well.
- Accepts files, directories, path globs, or a mix.
- `--include` and `--exclude` filter indexed file paths using glob matching. `--include` can temporarily bring back matching files hidden by `.gitignore` or `.ignore`.
- `--unique` deduplicates by enclosing function.
- Default result limit is `50`.
- Use `--limit <N|all>` and `--offset` when a common name fans out widely.

Examples:

```bash
gx references --name 'Search' .
gx references --unique --name 'Search' .
gx references --include 'internal/**' --exclude '**/*_test.go' --name 'Search' .
gx references --name 'Search' --limit 25 --offset 25 .
```

### `gx index`

Use `index` to build or incrementally refresh the project index before running queries.

```bash
gx index
gx -C /path/to/project index
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
- Those paths may be directories, files, path globs, or a mix.
- `symbols`, `definition`, and `references` also accept repeatable `--include` and `--exclude` glob filters over indexed file paths. `--include` can temporarily re-index matching ignored files for the current query.
- Prefer the smallest scope that answers the question.
- `-C DIR` runs `gx` as if started in `DIR`.
- A directory excluded by `.gitignore`/`.ignore` (for example `frontend/` at a repo root) silently returns zero results at the parent root. Run `gx -C path/to/subdir` to query inside it, or add `--include 'subdir/**'` to bring the matching files back inside the current query.

Examples:

```bash
gx overview internal/tmdb internal/tmdb/search.go README.md
gx symbols --name 'Search' .
gx -C /path/to/project symbols --name 'Search' .
gx symbols --include 'internal/**' --exclude '**/*_test.go' --name 'Search' .
gx symbols --name 'Search' internal/tmdb/search.go internal/tools
gx definition --name 'Search' internal/tmdb/search.go
gx references --unique --name 'Search' internal/tmdb/search.go internal/tools
```

## Output Mode

Use default output for terminal reading. Use `--json` when another tool or script will consume the result, or when an agent needs structured fields.

List results are wrapped in an envelope: `{"rows": [...], "page": {...}}`. The `page` object carries `total`, `offset`, `next_offset` (absent or 0 when no next page), and `out_of_range` (true when `--offset` exceeds `total`). Use `next_offset` for machine pagination instead of parsing stderr hints.

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
- Narrow the query before reaching for `--limit all`.
- Read `page.next_offset` from the output envelope to move forward through a large result set.

## Response Style

When helping the user with `gx`:

- explain the command sequence in terms of intent, not just syntax
- prefer the smallest scope that answers the question
- explicitly point out when the task is not a good fit for `gx`
- call out shell quoting when patterns use braces or wildcards
- do not turn the explanation into a comparison guide for other search tools
