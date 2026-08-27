# TASK-260822-5wfdfx — manifest fallback must fail loud

Scope: `internal/skillspec`. Prove that a present canonical manifest that cannot be
parsed fails loud and never falls back to `agents/runtime.json`; make the fallback
reachable only when no manifest exists at all.

## Result

Two defect classes were separated:

1. **Broken manifest body** (malformed JSON, `schema_version` the build does not
   know, invalid body). Already fail-loud at HEAD — `Load` returned the parse
   error and never reached the legacy manifest. Now pinned by regression tests
   instead of resting on the shape of one `if`.
2. **Manifest present but not inspectable** (dangling manifest symlink, snapshot
   directory the process cannot read). **Real defect at HEAD**: `Load` probed
   presence with `os.Stat` and treated *every* stat failure as absence, so a
   broken canonical manifest silently produced the legacy spec, or an empty
   pure-context spec. Fixed.

## Change

`internal/skillspec/parse.go`

- `Load` now decides precedence through `manifestPresent`, and any presence
  probe that fails for a reason other than "does not exist" is returned as an
  error instead of falling through to the next manifest.
- `manifestPresent` uses `os.Lstat`, not `os.Stat`, so a manifest symlink whose
  target is gone counts as a **broken manifest**, not an absent one.
- The same rule is applied to `agents/runtime.json`: an unreadable legacy
  manifest is an error, never a silent pure-context skill.

The failure mode this closes is the one recorded in `LOGBOOK.md` §1615 for the
Python implementation: a silently empty `SkillSpec` flips `include_scripts` and
pulls runtime scripts into agent context, so downstream assertions go vacuous
rather than red.

## Tests — `internal/skillspec/parse_test.go`

| Test | Pins |
| --- | --- |
| `TestBrokenCskSkillNeverFallsBackToLegacy` | 5 broken bodies (incl. `schema_version: 99`) with a parseable `agents/runtime.json` planted alongside → error, `spec == nil` |
| `TestNewerSchemaVersionReportsUpgradeHint` | error path is `schema_version`, message carries `UpgradeHint` and names `99` |
| `TestLegacyFallbackNeedsAbsentManifest` | positive control: same legacy payload loads only once the canonical manifest is removed |
| `TestUnreadableCskSkillNeverFallsBackToLegacy` | mode-000 manifest whose body would parse cleanly → `fs.ErrPermission`, no fallback |
| `TestCskSkillDirectoryNeverFallsBackToLegacy` | manifest entry that can never be read as a file → error naming `csk-skill.json` |
| `TestDanglingCskSkillSymlinkNeverFallsBackToLegacy` | presence is `Lstat`, not `Stat` |
| `TestUnreadableSnapshotNeverDegradesToEmptySpec` | unreadable snapshot ≠ pure context skill |
| `TestDanglingLegacyManifestNeverDegradesToEmptySpec` | same rule for the legacy manifest |

All eight run (none skipped) on darwin/arm64.

### Skip guards vs the CI skip-class gate

Three reasons can skip on a host that ignores POSIX modes (Windows, root). Each
was checked against `.github/ci/skip-classes.tsv` and matches an `allow` row of
class `host-capability`, so `platform-case-gate.sh` tolerates them by name:

- `this host cannot create an unreadable file: mode 000 stays readable here`
- `this environment can inspect a mode-000 directory`
- `symlinks unavailable: <err>`

No `.github/ci/platform-cases.tsv` row was added: `internal/skillspec` has no
tier-1 rows, and the reasons above are covered by tier 2.

## Mutation check

Run in a throwaway copy of the module (`.temp/TASK-260822-5wfdfx/mutant`, removed
afterwards) — the live checkout was never left mutated.

| Mutant | `go test ./internal/skillspec/` | Killed by |
| --- | --- | --- |
| M1 — swallow the parse error and fall through to the legacy manifest | **exit 1**, 12 failing cases | every regression test + the pre-existing `mustFail` suite |
| M2 — probe presence with `os.Stat` (exact pre-fix HEAD semantics) | **exit 1** | only the three presence tests; the schema-99 tests still pass, which is the correct diagnosis |

M2's diagnostic is the defect verbatim:

```
--- FAIL: TestDanglingCskSkillSymlinkNeverFallsBackToLegacy
    present manifest degraded into a parsed spec:
    &{SchemaVersion:1 SourceFile:agents/runtime.json ...
      Commands:map[fallback:{Name:fallback Type:script UnixPath:scripts/fallback}]}
```

Raw streams: `mutant-1-swallow-parse-error.log`, `mutant-2-stat-presence.log`.

## Validation (real exit codes, no pipes)

| Command | Exit |
| --- | --- |
| `go test -count=1 ./internal/skillspec/` | 0 |
| `go test -count=1 ./internal/skillspec/ ./internal/skillcheck/ ./internal/closure/ ./internal/install/ ./internal/audit/ ./internal/interop/` (every `skillspec.Load` consumer) | 0 |
| `go test -count=1 ./...` | 0 |
| `go vet ./...` | 0 |
| `GOOS=windows go vet ./internal/skillspec/` / `GOOS=linux …` | 0 / 0 |
| `gofmt -l cmd internal` | 0, no files |
| `golangci-lint run` (full module) | 0, 0 issues |
| `go build -o bin/curator ./cmd/curator` | 0 |

**Not green, not caused by this change:** `make ledger-check` → **exit 2**.
It fails on six pre-existing rows naming packages that are not in the module
(`internal/transaction`, `internal/install/atomicity`, `internal/godriver`) —
work that has not landed yet. `.github/` is untouched by this task and no
`internal/skillspec` row is involved. Log: `ledger-check.log`.

**Not run:** the conformance-root gates (`make ci-test`, `make check-ci`,
`candidate-*`) — they require `CURATOR_CONFORMANCE_ROOT`, which is unset on this
host. `TestPortablePathConformanceVectors` skips for the same reason.

## Notes for review

- The task description names the canonical manifest `agent-skill.json`. In this
  repo it is `csk-skill.json` (`docs/implementation-plan.md:41` treats the file
  names as wire format). The `agent-skill.json` rename exists in the rc.5
  conformance fixture root and is tracked separately (`LOGBOOK.md` §1615,
  `TASK-260720-z9j4c9`); it is not implemented in the Go parser, so the tests
  target the name this build actually reads. When the rename lands, these tests
  follow it with the filename constant.
- Anomaly: two implementer spawns (`RUN-260822-c1807e`, `RUN-260822-b89c6b`) ran
  concurrently on this one task and both edited `internal/skillspec/parse_test.go`
  in the same checkout. The first three tests in the table above arrived from the
  twin run; this run built the presence cases on top and reconciled the file. A
  transient `Load` mutation and a NUL-truncated `internal/config/buildssh_test.go`
  from neighbouring runs briefly reddened `go vet ./...` mid-session; both
  resolved. Final state above was measured on the reconciled tree.
