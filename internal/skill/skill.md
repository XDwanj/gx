# gx — Semantic Code Navigation

`gx` is derived from `cx`. When `gx` is available in the project, prefer it over reading files directly.

## Escalation hierarchy: directory overview → file overview → symbols → definition / references → read

- **Explore a directory** → `gx overview <dir>` (~20 tokens per entry)
- **Understand a file's structure** → `gx overview <file>` (~200 tokens)
- **Find symbols across the project** → `gx symbols [--kind K] [--name GLOB] [--file PATH]`
- **Read a specific function/type** → `gx definition --name <name>` (~500 tokens)
- **Find all usages of a symbol** → `gx references --name <name>` shows every usage with enclosing function and context
- **Check blast radius before refactoring** → `gx references --name <name> --unique` shows one row per dependent function
- **Fall back to Read tool** only when you need the full file or surrounding context beyond the symbol body

## When to use gx instead of Read

- **Exploring a new codebase** — start with `gx overview .` to see top-level structure, then drill into subdirectories. Cheaper than `ls` + reading files.
- **Before reading a file** — run `gx overview` first. You often don't need the full file.
- **Before editing a function** — `gx definition --name X` gives you the exact text for Edit tool's `old_string` without reading the whole file.
- **Before refactoring** — `gx references --name X --unique` shows which functions depend on X (one row per caller). Use without `--unique` to see every usage with context lines.
- **Understanding how a symbol is used** — `gx references --name X` shows each usage site with the enclosing function and the source line, so you can see if it's called, used as a type, imported, etc.
- **Exploring a codebase** — use `gx symbols` to find what you need across files, then `gx definition` to read specific symbols. Avoid reading file after file.
- **After context compression** — if you previously read a file but the content was compressed out, use `gx overview` to re-orient and `gx definition` for the specific symbols you need. Don't re-read the full file.

## Quick reference

```
gx overview PATH                                    file or directory table of contents
gx overview DIR --full                              directory overview with signatures
gx symbols [--kind K] [--name GLOB] [--file PATH]   search symbols project-wide
gx definition --name NAME [--from PATH] [--kind K]  get a function/type body
gx references --name NAME [--file PATH] [--unique]  find all usages (--unique: one per caller)
gx lang list                                         show supported languages
gx lang add LANG [LANG...]                           install language grammars
```

Short aliases: `gx o`, `gx s`, `gx d`, `gx r`

Symbol kinds: fn, method, struct, enum, trait, type, const, class, interface, module, event

Check signatures for `pub`/`export` to identify public API without reading the file.

## Missing grammars

If gx reports a missing grammar (e.g. `rust grammar not installed — run: gx lang add rust`), install it with `gx lang add rust`. Run `gx lang list` to see what's installed.
