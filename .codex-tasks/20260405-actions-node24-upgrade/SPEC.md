# Task Spec

## Goal

Upgrade GitHub Actions workflow dependencies that still run on Node.js 20 so repository workflows avoid the GitHub deprecation warning and align with Node.js 24-compatible action majors.

## Scope

- Update `.github/workflows/lint.yml`
- Update `.github/workflows/release.yml`
- Replace `mlugg/setup-zig` with an explicit Zig install step on Linux runners
- Keep existing workflow behavior unchanged apart from action version upgrades

## Validation

- `git diff --check`
- `actionlint .github/workflows/*.yml` if `actionlint` is available
