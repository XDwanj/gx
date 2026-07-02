---
name: gx
description: "ALWAYS activate this skill for any task that requires reading, understanding, locating, tracing, reviewing, debugging, or modifying an existing codebase. `gx` is the default first tool for semantic code navigation and must be used before text-based tools for symbol structure, definitions, references, package layout, and architecture exploration. Only do not activate it for non-code files or pure literal text search such as logs, docs, comments, YAML, JSON, SQL, and Markdown."
---

# GX

`gx` is a semantic code navigation tool for AI agents. Use it when the task is about code structure, declarations, definitions, callees, references, or package layout. Do not use it for plain text lookup in logs, comments, Markdown, YAML, JSON, SQL, or arbitrary strings.

## Default Workflow

Use the usual narrowing loop:

1. `overview` for package or file surface.
2. `symbols` to find the exact declaration.
3. `definition` to read the implementation body.
4. `callees` to inspect outgoing calls.
5. `references` to inspect impact.
6. `tree` with `--define-in` when you need an AI-pruned in/out tree.

Default examples:

```bash
gx overview internal/tmdb
gx symbols --kind func --name 'Search' .
gx definition --name 'Search' internal/tmdb
gx callees --name 'Search' internal/tmdb
gx references --unique --name 'Search' .
gx tree --name 'Search' --define-in internal/tmdb/search.go .
```

Use `--define-in FILE` when a common symbol name needs AI disambiguation against the definition in a specific file. This requires `GX_OPENAI_API_KEY` and `GX_OPENAI_BASE_URL`, accepts optional `GX_OPENAI_MODEL`, and applies to `symbols`, `definition`, `callees`, `references`, and `tree`. The `tree` command requires `--define-in`.

```bash
gx references --name 'login' --define-in internal/domain/user.go .
gx tree --name 'login' --define-in internal/domain/user.go .
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
- `symbols` and `definition` support `--kind`; `callees` and `references` do not.
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
- `zig`: `func`, `struct`, `enum`

## Command Guide

### `gx overview`

Use `overview` before reading code.

- Accepts multiple paths and returns one section per target when more than one path is supplied.
- Directory mode returns per-file symbol summaries.
- File mode returns top-level symbols for that file.
- Directory mode supports `--limit`, `--offset`, and `--all`.
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
- Use `--limit`, `--offset`, and `--all` for broad matches.

Examples:

```bash
gx symbols --name '*Search*' internal/tmdb
gx symbols --kind func --name 'Search' .
gx symbols --include 'internal/**' --exclude '**/*_test.go' --name '*Search*' .
gx symbols --name '*Search*' --limit 20 --offset 20 internal/tmdb
```

### `gx definition`

Use `definition` when you already know the symbol and want the body immediately.

- Returns the symbol body with a `file:line` header in terminal output. `--json` keeps separate `file` and `line` fields.
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
- Use `--limit`, `--offset`, and `--all` when a common name fans out widely.

Examples:

```bash
gx references --name 'Search' .
gx references --unique --name 'Search' .
gx references --include 'internal/**' --exclude '**/*_test.go' --name 'Search' .
gx references --name 'Search' --limit 25 --offset 25 .
```

### `gx callees`

Use `callees` to inspect calls made inside a function body when the callee is defined in the same source directory and current query scope.

- Returns scoped call sites, not resolved definitions.
- Accepts files, directories, path globs, or a mix.
- `--include` and `--exclude` filter indexed file paths using glob matching. `--include` can temporarily bring back matching files hidden by `.gitignore` or `.ignore`.
- Default result limit is `50`.
- Terminal output fields are `file`, `caller`, `callee`, and `context`, with `file` rendered as `path/to/file:line`. `--json` keeps separate `file` and `line` fields.

Examples:

```bash
gx callees --name 'Search' .
gx callees --name '*Search*' internal/tmdb
gx callees --exclude '{**/*_test.go,**/mocks/**}' --name 'Search' .
gx callees --name 'Search' --limit 25 --offset 25 .
```

### `gx tree`

Use `tree` to inspect AI-pruned incoming calls, outgoing calls, or both for a function.

- Requires `--name` and `--define-in`.
- Defaults to `--direction both` and `--depth 8`.
- Runs AI pruning in parallel with an in-process limit of 256 API requests.
- `--direction in` shows functions that call into the target.
- `--direction out` shows scoped project functions called by the target.
- Tree expansion only links functions defined in the same source directory and current query scope.
- `--verbose` shows tree expansion plus AI cache/API pruning progress.
- Does not support `--limit`, `--offset`, or `--all`; use `--depth` to control output size.
- Accepts files, directories, path globs, or a mix.
- `--include` and `--exclude` filter indexed file paths using glob matching.
- Terminal and `--json` output use nested `in` / `out` objects. Each node includes `file` as `path/to/file:line` and `symbol`; `cycle` appears only when recursion finds a cycle. Call context is used for AI pruning but is not printed.

Examples:

```bash
gx tree --name 'Search' --define-in internal/tmdb/search.go .
gx tree --name 'Search' --define-in internal/tmdb/search.go --direction in .
gx tree --name 'Search' --define-in internal/tmdb/search.go --direction out --depth 2 .
gx tree --exclude '{**/*_test.go,**/mocks/**}' --name 'Search' --define-in internal/tmdb/search.go .
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

- `overview`, `symbols`, `definition`, `callees`, and `references` accept multiple paths.
- Those paths may be directories, files, path globs, or a mix.
- `symbols`, `definition`, `callees`, and `references` also accept repeatable `--include` and `--exclude` glob filters over indexed file paths. `--include` can temporarily re-index matching ignored files for the current query.
- Prefer the smallest scope that answers the question.
- Use `-C` when you need to run `gx` against another directory context.

Examples:

```bash
gx overview internal/tmdb internal/tmdb/search.go README.md
gx symbols --name 'Search' .
gx -C /path/to/project symbols --name 'Search' .
gx symbols --include 'internal/**' --exclude '**/*_test.go' --name 'Search' .
gx symbols --name 'Search' internal/tmdb/search.go internal/tools
gx definition --name 'Search' internal/tmdb/search.go
gx callees --name 'Search' internal/tmdb/search.go internal/tools
gx references --unique --name 'Search' internal/tmdb/search.go internal/tools
```

## Output Mode

Use default output for terminal reading. Use `--json` when another tool or script will consume the result, or when an agent needs structured fields.

Examples:

```bash
gx definition --json --name 'Search' internal/tmdb
gx callees --json --name 'Search' .
gx references --json --unique --name 'Search' .
```

## Practical Heuristics

- Start from a package directory, not the repo root, when using `overview`.
- Use `symbols` before `definition` if the name or `kind` may be ambiguous.
- Use `callees` after `definition` when you need the outgoing call list for one implementation.
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
