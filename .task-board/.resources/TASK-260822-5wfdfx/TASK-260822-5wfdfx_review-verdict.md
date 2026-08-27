# TASK-260822-5wfdfx — review verdict: ACCEPTED

Reviewer run `RUN-260822-694018`, read-only. No file in `internal/skillspec` was
modified by this review; the mutation checks below ran against a throwaway copy under
`.temp/TASK-260822-5wfdfx-mutation/`, never the live checkout.

## Verdict

**accepted.** The acceptance criteria hold, the delivered change is the right one, and
every claim the implementer made about it reproduces independently. Three findings are
carried forward as notes, none of them a defect in the delivered code.

## What was delivered

`internal/skillspec/parse.go` (+34/-3) and `internal/skillspec/parse_test.go` (+203).

The task brief assumed a test-only change. The parse-error half of the contract was
indeed already correct at HEAD — `Load` returned `loadCskSkill(cskPath)` error included,
so a present-but-malformed manifest was already terminal. The real defect was one level
earlier, in **presence detection**: `Load` probed with `os.Stat` and treated *every*
stat failure as absence. A dangling `csk-skill.json` symlink therefore loaded
`agents/runtime.json` and returned a spec with `SourceFile: agents/runtime.json` and the
legacy commands; an unreadable snapshot directory returned the empty pure-context spec,
silently stripping every command the skill has.

The fix routes both manifest probes through `manifestPresent`, which uses `os.Lstat` and
returns an error for any failure that is not `fs.ErrNotExist`. This is not scope creep:
the story clause reads "a present-but-unreadable **or** newer-schema canonical manifest
must fail loud", and the AC reads "fallback reachable only when no manifest file exists
at all". Presence detection is exactly what those two sentences are about.

## Independent verification

### Acceptance criteria, checked one by one

| AC clause | Evidence |
|---|---|
| Regression test in `internal/skillspec` | 8 new tests, all present in `parse_test.go` |
| schema 99 + `runtime.json` alongside errors with upgrade hint, no fallback | `TestNewerSchemaVersionReportsUpgradeHint` asserts `*verr.Error`, path `schema_version`, message carries `UpgradeHint` and `99`; `TestBrokenCskSkillNeverFallsBackToLegacy/newer_schema_version` asserts error + **nil** spec with a valid fallback planted |
| Fallback reachable only when no manifest exists at all | `TestLegacyFallbackNeedsAbsentManifest` is the positive control: same payload, manifest removed, loads with `SourceFile == "agents/runtime.json"` |
| `go test` green | `go test -count=1 ./...` exit **0**, 31 packages ok, 0 FAIL |

All 8 tests were confirmed to actually **run** on this host — verbose output shows 8
PASS, zero SKIP. Every platform-dependent test carries a `t.Skip` guard
(mode-000 files, mode-000 directories, `os.Symlink`) so a Windows or root lane degrades
to a skip rather than a false failure.

### Mutation checks — reproduced, not taken on trust

Both were re-run in a throwaway module copy.

**Mutant 1 — revert `parse.go` to HEAD (`os.Stat` presence).** Exit 1. Exactly three
tests fail, and they are exactly the three presence tests:
`TestDanglingCskSkillSymlinkNeverFallsBackToLegacy` (leaked
`SourceFile:agents/runtime.json` with the legacy `fallback` command),
`TestUnreadableSnapshotNeverDegradesToEmptySpec` and
`TestDanglingLegacyManifestNeverDegradesToEmptySpec` (both leaked the empty spec).
The schema-99 tests stayed green, which independently confirms the implementer's
finding that the parse-error half was already correct.

**Mutant 2 — swallow the manifest parse error** (`if spec, err := loadCskSkill(cskPath);
err == nil { return spec, nil }`). Exit 1, 19 failures including all five
`TestBrokenCskSkillNeverFallsBackToLegacy` subtests and
`TestNewerSchemaVersionReportsUpgradeHint`.

Both mutants match the implementer's attached `TASK-260822-5wfdfx_mutant-2-stat-presence.log`
line for line. The tests bite.

### Gates, re-run standalone with real exit codes

| Command | Exit |
|---|---|
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `GOOS=windows go vet ./internal/skillspec/` | 0 |
| `GOOS=linux go vet ./internal/skillspec/` | 0 |
| `gofmt -l cmd internal` | 0, no output |
| `go test -count=1 ./...` | 0, 31 ok, 0 FAIL |
| `golangci-lint run` | 0, "0 issues." |

### Blast radius

Both `Load` consumers already treat its error as fatal —
`internal/closure/closure.go:288` wraps it into "invalid skill manifest for ...", and
`internal/skillcheck/skillcheck.go:32` turns it into a `skill.manifest_invalid` issue.
The new error paths surface through existing channels; nothing degrades silently.

One behavior change worth naming: `errors.Is(ENOTDIR, fs.ErrNotExist)` is **false** on
darwin (verified by probe), so `Load("<a regular file>")` now returns
`cannot determine whether .../csk-skill.json exists: not a directory` instead of the
empty pure-context spec. That is correct — a snapshot that is a regular file is not a
pure context skill — and it converges with `main`, whose `pathExists` uses `os.IsNotExist`
and behaves the same way.

## Findings carried forward (none blocking)

### F1 — the fix does not reach `main` by merging, and `main` still carries the defect

`main` (47 commits ahead of this branch) has already replaced this `Load` entirely:
`agent-skill.json` is canonical, `csk-skill.json` is a legacy read alias, dual files are
compared with `reflect.DeepEqual` and a mismatch is `conflicting_skill_manifests`, and
presence goes through a helper named `pathExists`.

That helper is `os.Stat` + `os.IsNotExist` (`main:internal/skillspec/parse.go:91-100`).
**So `main` still has the exact defect this task just proved and fixed** — a dangling
manifest symlink is `IsNotExist` under `Stat`, so it reads as absence and falls through
to the next manifest, and on `main` there are *three* presence probes to correct, plus
`ManifestSourcePath` which swallows the error outright (`pathExists(...)` with the error
discarded). The port is a real piece of work, not a merge.

The new tests do not port for free either: they hardcode the string `csk-skill.json`
where `main` uses `CanonicalManifestName`/`LegacyManifestName` constants, and `main`'s
resolution has a dual-file conflict path this branch has no concept of.

Recommend a follow-up task against `main` rather than folding it in here — this task's
scope is this branch's code, and it discharged it.

### F2 — `agent-skill.json` is not a phantom; `LOGBOOK.md:220` is wrong about it

The board note and `LOGBOOK.md:220` frame the description's `agent-skill.json` as "a
stale alias in upstream task text, not a rename to chase". That framing is wrong.
`curator-spec/protocol/core.md:35-40` makes `agent-skill.json` the canonical
implementation-neutral filename and `csk-skill.json` a reserved legacy **read** alias
that writers must not emit, and `main` already implements exactly that. The description
was using the spec's name.

The implementer's *decision* was still right — `agent-skill.json` genuinely does not
exist anywhere on this branch, so tests naming it would not compile against anything
real. Only the justification needs correcting, so nobody later treats the canonical spec
filename as upstream noise.

Related, and it narrows F1's residual value: `main`'s `TestCanonicalManifestResolution`
already has an `invalid peer does not fall back` subtest. It covers a malformed peer
between the two *modern* manifests only — it plants no `agents/runtime.json`, tests no
`schema_version` 99, and tests no presence failure. So this task's coverage is
genuinely new, not a duplicate of `main`.

### F3 — a stale outcome artifact contradicts what shipped

Two implementer spawns (`RUN-260822-c1807e`, `RUN-260822-b89c6b`) ran concurrently on
this one task in the same checkout — the same hazard recorded today for
`TASK-260822-96m5pj`. The first spawn's artifact
`TASK-260822-5wfdfx_results.md` states "**No production code was changed** (`git diff`
on `parse.go` is empty)", which the delivered diff falsifies. `LOGBOOK.md` entry 2016
carries the same superseded claim; entry 2025 supersedes it 9 minutes later.

The authoritative artifact is `TASK-260822-5wfdfx_manifest-fallback-fail-loud.md`, and
the board notes carry the correct final story. Recommend the commit-owning mover update
`results.md` in place rather than leave two board artifacts telling different stories.

### Cosmetic

`plantLegacyFallback` is declared between test functions, after its first use — a visible
seam from the two-spawn reconciliation. Compiles and reads fine; not worth a cycle.

## Handoff

Reviewer supplies **no `commit_ack`**. Scope to commit is exactly two files:
`internal/skillspec/parse.go` and `internal/skillspec/parse_test.go`, plus this task's
`LOGBOOK.md` entries. The working tree carries unrelated uncommitted work from sibling
tasks (`cmd/curator/`, `internal/closure/`, `internal/config/`, `internal/install/`,
`internal/skillcheck/`) — the commit-owning mover must not sweep it in.
