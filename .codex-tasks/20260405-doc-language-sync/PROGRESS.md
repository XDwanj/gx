# Progress

- Confirmed current built-in support includes `protobuf`.
- Updated `README.md` and `README.zh-CN.md` to list `protobuf` and document its kind mapping.
- Updated `internal/skill/skill.md` so embedded `gx skill` output now includes `protobuf`.
- Checked terminal help and existing tests: `gx definition --help` already listed `protobuf`, and protobuf coverage already exists under `tests/` and unit tests.
- Added a `CLAUDE.md` rule to always review `README`, embedded `skill.md`, terminal help, and `tests/` when a public contract changes.
