# Markdown Overview Support

## Goal

Add Markdown support for `gx overview` only.

## Scope

- Support `.md` and `.markdown` files in `overview`
- Extract and display heading structure for levels `h1` through `h6`
- Keep `symbols`, `definition`, and `references` unsupported for Markdown
- Add automated tests and update user-facing docs

## Non-Goals

- No Markdown indexing
- No Markdown symbol extraction
- No Markdown definition or reference queries
- No fenced code block language parsing
