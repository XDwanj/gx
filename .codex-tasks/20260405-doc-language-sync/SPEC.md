# Doc Language Sync

## Goal

Synchronize repository docs with the current built-in language support exposed by `gx`.

## Scope

- Update `README.md`
- Update `README.zh-CN.md`
- Update `internal/skill/skill.md`

## Source Of Truth

- `internal/lang/lang.go`
- `cmd/help_kind_support.go`

## Validation

- Review `git diff -- README.md README.zh-CN.md internal/skill/skill.md`
- Verify `protobuf` appears in the expected language lists and coverage sections
