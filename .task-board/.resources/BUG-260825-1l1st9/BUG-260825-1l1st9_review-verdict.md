# BUG-260825-1l1st9 — review verdict: ACCEPTED

Reviewer run `RUN-260825-723452` (not goal-bound). Read-only review; no code changed.

Scope reviewed: PR #41 (`fix/BUG-260825-1l1st9-marker-v4-compiled-banding`),
head `eb58395`, merged as `5cbd1b8` on `main` at 2026-08-25T00:36:15Z.
Merge is a clean no-op: `tree(5cbd1b8) == tree(eb58395) == d298308`, first
parent `272b203`.

## Review workspace

The STORY-260822-2lvw0e worktree is at `903af23`, which predates
`PolicySchemaVersion` entirely — the delta is not reviewable there. Review was
run in a detached worktree `.temp/BUG-260825-1l1st9/review-wt` at `5cbd1b8`
(submodules initialised). The story branch was not touched.

## Verdict

Accepted. The fix is correct, minimal, complete for its class, and the tests
prove the gate in both directions. Every AC clause is independently verified
below.

## AC clause 1 — status reports project-management and its three compiled commands installed/current on marker v4

Verified live on this machine, without reinstalling:

- `curator v0.14.0-rc.3-48-g5cbd1b8` (`/Users/iv/.local/bin/curator`)
- `~/.curator/global/skills/project-management/.csk-install.json`:
  `schema_version 4`, `skill_schema_version 8`, `commit 3958813`,
  `installed_at 2026-08-24T23:20:46Z`, mtime `Aug 25 03:20:47 2026`.
  The marker predates the merge (00:36:15Z) and was not rewritten, so the flip
  came from the reader, not from a reinstall.
- `curator global status --json` → `project-management: "up-to-date"`; the
  three compiled commands `task-board`, `task-board-tui`, `tb-sessiond` each
  report `"state": "current"`.

## AC clause 2 — regression test pins it

The regression drives the real production entry point, not a helper:
`statusReport` (`cmd/curator/main.go:768`) is called by both
`curator status` (`main.go:677`) and `reportGlobalStatus` → `main.go:1219`.
`TestStatusReportFindsASchema8InstallationCurrent` calls `statusReport`
directly over a marker produced by the real `marker.Write`, and asserts the
writer really emitted schema 2/3/4 for manifest bands 6/7/8 — so a green run
cannot come from a silent fallback to schema 2.

All three fixed readers are production-reachable:

| Reader | Production call site |
| --- | --- |
| `classifySkillBuilds` | `cmd/curator/main.go:807`, inside `statusReport` |
| `markerRefusal` | `cmd/curator/main.go:895`, and `builds.go:362` |
| `marks.absorb` | `markScopes` ← `scopes.Collect` ← `cmd/curator/main.go:1713` |

## AC clause 3 — merged green

PR #41 lanes, all SUCCESS pre-merge: Test (ubuntu/macos/windows), Race
(ubuntu/macos), Lint, Gate self-test (ubuntu/macos/windows), Interop
conformance gate, Naming gate. "Candidate suite" is SKIPPED by design — it is
gated on `workflow_dispatch` with a `candidate_ref`/`candidate_root` input
(`.github/workflows/ci.yml:303-305`), not on PRs.
Implementation commit `564371c` is signed (`%G? = G`).

## Independent mutation verification (reviewer-run, not accepted from the artifact)

Five mutants applied to the merged tree and reverted; every one caught, both
narrowing and widening directions:

| # | Mutant | Result |
| --- | --- | --- |
| 1 | `classifySkillBuilds` narrowed back to `recorded.SchemaVersion != marker.SchemaVersion` (the original defect) | RED — `TestClassifySkillBuilds...` + `TestStatusReportFindsASchema8...`, reproducing the reported string verbatim for schemas 3 and 4 |
| 2 | `marks.absorb` reverted to the old `{2,3}` band | RED — `TestCollectMarksBuildKeys...` + `TestAbsorbKeysBuildLiveness...` |
| 3 | `BuildBearingSchema` widened to `version != LegacySchemaVersion` | RED in all three packages (`internal/marker`, `internal/scopes`, `cmd/curator`) |
| 4 | `NewestSchemaVersion` left stale at `ExternalSchemaVersion` | RED — `TestSupportedSchemaBandIsExactAndNewestIsItsMaximum` |
| 5 | `SupportedSchema` widened to `version >= LegacySchemaVersion` | RED — `TestUnsupportedSchemaErrors` + the band test |

Working tree confirmed clean against `5cbd1b8` after every revert.

## Completeness of the class

`grep` over the merged tree for install-marker-schema references outside
`internal/marker` returns exactly one hit — `cmd/curator/builds.go:555`, the
operator message naming `marker.NewestSchemaVersion`. No hand-rolled schema
inequality survives. The remaining per-schema comparisons inside
`internal/marker` are validation rules (each schema pins a distinct
`skill_schema_version` band and build-record shape), not banding checks.

The second, unreported defect is real and correctly rated: `marked.builds`
feeds `Referenced` in `cache.Sweep` (`internal/scopes/gc.go:114`), so an
unmarked marker-v4 cache key would have been swept out from under a live
schema-8 installation.

## Tests corrected, not deleted

No test function was removed by the diff (`git diff | grep '^-.*func Test'` is
empty; 16 removed test lines total, all in-place edits). Two tests that
asserted the defect were corrected:

- `TestMarkerRefusalSeparatesUnsupportedFromInvalid` pinned `schema_version 3`
  as `unsupported-marker`; now `invalid-marker`, with schema 5 taking over the
  unsupported case, and a new assertion that the detail names the real newest
  readable schema.
- The status matrix case `marker schema cannot be read by this manager` was
  anchored to `marker.SchemaVersion + 1` (written + 1) and is now anchored to
  `marker.NewestSchemaVersion + 1` (readable band + 1), which is the only
  anchor that stays a negative test as the read band widens. A sibling case
  covers a readable-but-invalid document.

## Conformance-gap artifact — claims spot-checked against the pin

Every factual claim in `BUG-260825-1l1st9_conformance-gap.md` was verified
against `curator-spec` at `0ed5c691` (the committed `SPEC_PIN` in
`.github/workflows/ci.yml:44`):

- `status_cases` has exactly two cases, named `compiled-installation-current`
  and `compiled-currentness-failure-matrix`. ✓
- `compiled-installation-current.validated` does list `marker-schema`, and the
  case does not parameterise the schema. ✓
- `compiled-currentness-failure-matrix.independent_conditions` has 14 entries,
  none marker-schema in either direction. ✓
- `fixtures/go-build-skill/.csk-install.json` has a single key,
  `package_marker` — no `schema_version` to vary. ✓
- `gc_cases` holds `locked-mark-and-sweep-compiled-cache` and
  `post-commit-gc-failure-is-maintenance-warning`; neither is
  marker-schema-parameterised. ✓

## Reviewer gates rerun locally (darwin/arm64, `5cbd1b8`)

| Gate | Result |
| --- | --- |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `gofmt -l cmd internal` | clean |
| `golangci-lint run ./...` | 0 issues |
| `go test ./internal/marker/ ./internal/scopes/` | ok |
| `go test ./cmd/curator -run 'TestClassifySkillBuildsAcceptsEveryBuildBearingMarkerSchema\|TestStatusReportFindsASchema8InstallationCurrent\|TestMarkerRefusalSeparatesUnsupportedFromInvalid'` | PASS |
| `go test ./cmd/curator -run TestCompiledProjectStatusRepairRollbackRecovery` | PASS (193.9s, 15 currentness sub-cases incl. both new schema cases) |
| `go test ./internal/{marker,scopes,interop,registry,ui}` | ok (every `internal` package that imports `marker`, except `internal/install`) |

Explicitly **not** rerun locally, accepted from the PR lanes: `internal/install`
and the remaining `internal` packages, plus the Windows, Linux, race, interop
and naming lanes. `.github/ci/test-gate.sh` partitions the whole package set
into served/deferred/excluded stages and runs all three, so the three green
Test lanes plus two green Race lanes do cover `./...` on ubuntu, macos and
windows. A local `go test ./internal/...` was started and exceeded the
reviewer's bounded-call budget (`internal/install` performs real toolchain
builds); it was stopped rather than backgrounded, and its result is not claimed
here. This machine has no Linux or Windows runner.

## Non-blocking follow-ups (not grounds to withhold acceptance)

1. The conformance gap is recorded in the board artifact and `LOGBOOK.md`
   (entry 0417, `STATUS: pending`), but `curator-spec` has no `.task-board`, so
   there is no tracked item on the suite owner's side. Someone should carry the
   three requested vectors into whatever tracker that repo uses; otherwise the
   suite stays blind to this class.
2. Pre-existing and out of scope for this delta: `marker.Write` maps
   `SkillSchemaVersion >= 8` to `PolicySchemaVersion`, while `validMarker`
   requires `SkillSchemaVersion == 8` for schema 4. A future skill schema 9
   would therefore fail to write rather than being refused earlier with a clear
   reason. Worth a look when schema 9 is designed.

## Handoff

Reviewer-archetype run; no `commit_ack` supplied. The PR scope is already
committed and merged as `5cbd1b8` on `main`. The commit-owning mover makes the
final `done` transition with `commit_ack=scope_committed`, citing this artifact.
