# TASK-260822-5wfdfx — manifest-fallback-fail-loud

## Outcome

The fail-loud contract already held in `internal/skillspec/parse.go`; this task pins it
with regression tests. No production code was changed (`git diff` on `parse.go` is empty).

`Load` (internal/skillspec/parse.go:28) stats the canonical manifest first and returns
`loadCskSkill(cskPath)` **including its error**, so the `agents/runtime.json` branch is
structurally unreachable whenever a manifest file exists.

## Naming note (for the reviewer)

The task description calls the canonical manifest `agent-skill.json`. The repository's
canonical name is **`csk-skill.json`** (internal/skillspec/parse.go:29, types.go doc
comment, existing tests). Tests use the real name; the description name does not exist
anywhere in the code.

## Tests added — internal/skillspec/parse_test.go

1. `TestBrokenCskSkillNeverFallsBackToLegacy` — table test. Each case writes a broken
   `csk-skill.json` **plus a valid `agents/runtime.json`** next to it, then asserts
   `Load` returns an error and a nil spec:
   - `newer schema version` — `{"schema_version": 99}` (the AC case)
   - `malformed json` — truncated object
   - `not a json object` — top-level array
   - `empty file` — zero-byte manifest
   - `valid schema, invalid body` — schema 3 without `capabilities`
2. `TestNewerSchemaVersionReportsUpgradeHint` — schema 99 with a `commands` block that
   the fallback would happily accept. Asserts the error is a `*verr.Error` on path
   `schema_version` whose message contains both `skillspec.UpgradeHint` and `99`.
3. `TestLegacyFallbackNeedsAbsentManifest` — positive control. Same layout, manifest
   removed; the fallback then loads with `SourceFile == "agents/runtime.json"` and the
   `fallback` command intact. This proves the failures above come from manifest
   precedence, not from an unparseable fallback payload.

## Mutation check (proof the tests bite)

Temporarily rewrote `Load` to swallow the manifest error and continue to the fallback:

```go
if spec, err := loadCskSkill(cskPath); err == nil {
    return spec, nil
}
```

Result: all 5 subtests of `TestBrokenCskSkillNeverFallsBackToLegacy` plus
`TestNewerSchemaVersionReportsUpgradeHint` failed, each reporting the leaked
`SourceFile:agents/runtime.json` spec. `parse.go` was restored immediately afterwards
and re-verified clean via `git diff --stat`.

## Gates (each run as a standalone process, real exit codes)

| Command | Exit |
|---|---|
| `gofmt -l internal/skillspec/` | 0, no output |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `go test ./...` | 0 — 31 packages ok/no-test-files, 0 FAIL (`.temp/TASK-260822-5wfdfx-gotest.log`) |
| `golangci-lint run` | 0 — "0 issues." (`.temp/TASK-260822-5wfdfx-lint.log`) |
