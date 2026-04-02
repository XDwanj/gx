# gx

`gx` is a semantic code navigation tool for AI agents and developers who want to understand a codebase without repeatedly opening full files.

This project is derived from `cx`, but its public command name in this repository is `gx`.

## What it does

- Gives a file or directory overview before you read source files in full.
- Searches symbols across the project with kind and glob filters.
- Prints the body of a symbol directly, so you can inspect one function or type without opening the whole file.
- Finds references for a symbol, with an optional unique-per-caller view to estimate refactor blast radius.
- Manages language grammar availability and index cache state.
- Exposes an embedded `skill` document that tells AI agents how to use the tool efficiently.

## Supported languages

Current built-in language support includes:

`bash`, `c`, `cpp`, `elixir`, `go`, `java`, `lua`, `python`, `ruby`, `rust`, `solidity`, `swift`, `typescript`, `zig`

Use `gx lang list` to see which grammars are currently installed in your local cache.

## Build and run

Build a local binary:

```bash
make build
```

Run directly without building:

```bash
go run . --help
```

Run the standard checks:

```bash
make test
make lint
```

Build cross-compilation artifacts under `dist/`:

```bash
make cross
```

For non-Darwin targets, the project needs a cgo-capable cross compiler. The
default `Makefile` configuration expects `zig cc` for Linux and Windows targets.
On macOS, `make cross-darwin` builds both `darwin/arm64` and `darwin/amd64`
artifacts with `clang`.

## Quick start

See the top-level command list:

```bash
gx --help
```

Explore a directory before opening files:

```bash
gx overview cmd
```

Inspect the structure of a single file:

```bash
gx overview cmd/root.go
```

Find matching symbols across the project:

```bash
gx symbols --name 'new*'
```

Read one symbol body directly:

```bash
gx definition --name buildRuntime --max-lines 40
```

Find all usages of a symbol:

```bash
gx references --name buildRuntime
```

Estimate blast radius by caller:

```bash
gx references --name buildRuntime --unique
```

## Commands

### Navigation

- `gx overview <path>`: Show a table of contents for a file or directory.
- `gx overview <dir> --full`: Show a fuller per-file directory overview.
- `gx symbols [--file PATH] [--name GLOB] [--kind KIND]`: Search symbols across the project.
- `gx definition --name NAME [--from PATH] [--kind KIND] [--max-lines N]`: Print a symbol body.
- `gx references --name NAME [--file PATH] [--unique]`: Find symbol usages.

Short aliases: `gx o`, `gx s`, `gx d`, `gx r`

Supported symbol kinds:

`fn`, `method`, `struct`, `enum`, `trait`, `type`, `const`, `class`, `interface`, `module`, `event`

### Language management

- `gx lang list`: Show supported languages and install status.
- `gx lang add <languages...>`: Mark one or more grammars as installed in the local cache.
- `gx lang remove <languages...>`: Remove grammars from the local cache manifest.

### Cache management

- `gx cache path`: Print the cache file path for the current project index.
- `gx cache clean`: Remove the cached index for the current project.

### Agent integration

- `gx skill`: Print the embedded agent skill guide to stdout.

## Output modes

By default, `gx` prints a compact human-readable table format.

For machine-readable output, add `--json`:

```bash
gx symbols --name 'new*' --json
```

Use `--root <path>` if you want to query a project other than the current working tree.

## Typical workflow

1. Start with `gx overview .` or `gx overview <dir>`.
2. Narrow down with `gx symbols`.
3. Read the exact target with `gx definition`.
4. Check impact with `gx references --unique`.
5. Fall back to reading full files only when you need wider context.
