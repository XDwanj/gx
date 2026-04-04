---
name: gx
description: "ALWAYS activate this skill for any task that requires reading, understanding, locating, tracing, reviewing, debugging, or modifying an existing codebase. `gx` is the default first tool for semantic code navigation and must be used before text-based tools for symbol structure, definitions, references, package layout, and architecture exploration. Only do not activate it for non-code files or pure literal text search such as logs, docs, comments, YAML, JSON, SQL, and Markdown."
---

# GX

`gx` is a semantic code navigation tool for AI agents. Its usage pattern is close to LSP-style code intelligence: inspect symbol structure, jump to definitions, trace references, and narrow code reading to the relevant parts instead of scanning whole files up front.

## What `gx` Is Good At

Use `gx` for:

- LSP-style code navigation tasks
- semantic symbol lookup
- package and file structure overview
- definition lookup
- reference tracing
- low-noise context gathering for code understanding

## When `gx` Is Not The Right Tool

`gx` is usually not the right tool when the user:

- only knows a piece of text, not a symbol name
- wants plain keyword or full-text search
- is searching logs, error messages, config keys, comments, or documentation text
- needs to search arbitrary strings inside JSON, YAML, SQL, Markdown, or other non-code content
- is not trying to understand definitions, symbol structure, references, or code layout

## Default Workflow

In most repositories, the best workflow is the same kind of narrowing loop you would use with an LSP:

1. Start with `overview` to understand the package or file surface.
2. Use `symbols` to locate the exact symbol name and kind.
3. Use `definition` to read the implementation body.
4. Use `references` to inspect impact and call sites.

Default examples:

```bash
gx overview internal/tmdb
gx symbols --kind method --name 'Search' .
gx definition --name 'Search' internal/tmdb
gx references --unique --name 'Search' .
```

## Kind Model

`gx` uses one public `kind` vocabulary for commands that support `--kind`:

- `fn`
- `method`
- `const`
- `struct`
- `enum`
- `class`
- `interface`
- `module`
- `type`

Important rules:

- Treat these as `gx`'s public extraction kinds, not as a promise that every language supports every kind.
- `symbols` and `definition` support `--kind`; `references` does not.
- Language support is coverage-based. A missing result may mean the current language extractor does not expose that `kind` yet.
- The help text's per-language lists describe current extraction coverage, not the full syntax space of that language.

Practical implication:

- Narrow by `--name` first.
- Add `--kind` to reduce noise only after you know the target language likely exposes that `kind`.
- If `--kind` returns nothing, retry without it before assuming the symbol does not exist.

## Language Coverage

The current language coverage exposed by `gx symbols -h` and `gx definition -h` is:

- `bash`: `fn`
- `c`: `fn`, `struct`, `enum`, `type`
- `cpp`: `fn`, `method`, `struct`, `class`, `enum`, `module`, `type`
- `go`: `fn`, `method`, `const`, `struct`, `interface`, `type`
- `java`: `class`, `method`, `const`, `enum`, `interface`, `module`
- `lua`: `fn`, `method`
- `python`: `fn`, `const`, `class`
- `ruby`: `method`, `class`, `module`
- `rust`: `fn`, `method`, `const`, `struct`, `enum`, `interface`, `module`, `type`
- `swift`: `fn`, `method`, `const`, `struct`, `enum`, `class`, `interface`, `module`, `type`
- `typescript`: `fn`, `method`, `const`, `class`, `enum`, `interface`, `module`, `type`
- `zig`: `fn`, `struct`, `enum`

Use this table as the first check when a `--kind` query returns nothing.

- If the language and `kind` combination is not listed here, retry without `--kind` or switch to a different `gx` command.
- Treat this table as current extraction coverage, not as a statement about what the language itself can express.

## Command Guide

### `gx overview`

Use `overview` to build a fast mental model before reading code.

- For a directory, it returns a per-file symbol summary.
- For a file, it returns that file's top-level symbols.
- It accepts exactly one path.
- The path can be a file or a directory.
- Directory output supports `--limit`, `--offset`, and `--all`.
- File and Markdown outline modes ignore pagination flags.

Examples:

```bash
gx overview internal/tmdb
gx overview internal/tmdb/search.go
gx overview --full internal/tmdb
```

Use `overview` first when entering an unfamiliar module.

### `gx symbols`

Use `symbols` to find declarations.

- It returns definitions, not usages.
- It accepts multiple paths.
- Paths can be files, directories, or a mix of both.
- `--kind` filters by the public `gx` kind vocabulary.
- Check the Language Coverage table before assuming a `--kind` filter should work for the target language.
- `--name` matches symbol names with glob syntax.
- Default result limit is `100`.
- Use `--limit`, `--offset`, and `--all` to page through broad matches.

Examples:

```bash
gx symbols --name '*Search*' internal/tmdb
gx symbols --kind method --name 'Search' .
gx symbols --name '{Ali*Pay,Wechat*Pay}' internal/service internal/handler
gx symbols --name '*Search*' --limit 20 --offset 20 internal/tmdb
```

### `gx definition`

Use `definition` when you already know the symbol and want the body immediately.

- It accepts multiple paths.
- Paths can be files, directories, or mixed.
- It returns the symbol body with source file and line.
- `--kind` uses the same public kind vocabulary as `symbols`.
- If the Language Coverage table does not list that language and `kind` combination, remove `--kind` and retry.
- `--max-lines` controls output size.
- Default result limit is `5`.
- `--limit` controls how many definitions you get; `--max-lines` controls how large each definition body can be.

Examples:

```bash
gx definition --name 'Search' internal/tmdb
gx definition --name 'Client' internal/tmdb/client.go
gx definition --max-lines 80 --name '{Ali*Pay,Wechat*Pay}' internal/service
gx definition --name '*Search*' --limit 1 --max-lines 80 internal/tmdb
```

### `gx references`

Use `references` to find usages and impact.

- It returns where a symbol is used.
- It commonly includes the definition site as well.
- It accepts multiple paths.
- Paths can be files, directories, or mixed.
- `--unique` deduplicates by enclosing function and is usually better for quick analysis.
- Default result limit is `50`.
- Use `--limit`, `--offset`, and `--all` when a common name fans out widely.

Examples:

```bash
gx references --name 'Search' .
gx references --unique --name 'Search' .
gx references --unique --name '{Ali*Pay,Wechat*Pay}' internal/service internal/handler internal/repo
gx references --name 'Search' --limit 25 --offset 25 .
```

### `gx cache`

Use `cache` when the index seems stale or when you need to inspect where the project index lives.

Examples:

```bash
gx cache path
gx cache clean
```

### `gx lang`

Use `lang` to inspect or manage installed grammars.

Examples:

```bash
gx lang list
gx lang add go typescript
gx lang remove ruby
```

## `--name` Matching Rules

Assume `--name` uses glob-style matching.

Important rules:

- Matching is case-sensitive.
- `*` works as a wildcard.
- Brace alternation works with commas, not pipes.
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

Use these rules when choosing scope:

- `overview` accepts one path only.
- `symbols`, `definition`, and `references` accept multiple paths.
- Those paths may be directories, files, or a mix.
- Prefer narrower scopes first to reduce noise and speed up understanding.

Examples:

```bash
gx symbols --name 'Search' internal/tmdb/search.go internal/tools
gx definition --name 'Search' internal/tmdb/search.go
gx references --unique --name 'Search' internal/tmdb/search.go internal/tools
```

## Output Mode

Use default output when a human is reading the result in the terminal.

Use `--json` when:

- another tool or script will consume the result
- an agent needs structured output
- you want structured fields instead of terminal-oriented formatting

Examples:

```bash
gx definition --json --name 'Search' internal/tmdb
gx references --json --unique --name 'Search' .
```

## Best-Fit Scenarios

This skill is especially useful for:

- navigating a codebase in an LSP-style way from the terminal
- entering an unfamiliar package
- locating a function, method, or type by name
- tracing call sites before a refactor
- estimating impact radius of a code change
- gathering low-noise code context for an agent

## Practical Heuristics

Use these defaults unless the user asks for something different:

- Start from a package directory, not the repo root, when using `overview`.
- Use `symbols` before `definition` if the name or `kind` might be ambiguous.
- Treat `--kind` as a precision filter, not as the primary lookup mechanism.
- When results are unexpectedly empty, check the language coverage model before assuming the symbol is absent.
- Use `references --unique` for quick impact analysis.
- Use `--json` for automation and default output for ad hoc reading.

## Response Style

When helping the user with `gx`:

- explain the command sequence in terms of intent, not just syntax
- prefer the smallest scope that answers the question
- explicitly point out when the task is not a good fit for `gx`
- call out shell quoting when patterns use braces or wildcards
- do not turn the explanation into a comparison guide for other search tools
