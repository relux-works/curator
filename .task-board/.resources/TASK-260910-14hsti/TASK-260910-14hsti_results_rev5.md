# TASK-260910-14hsti — rework-3 handoff (rev5)

Producer run for orchestrator rework-3 (verdict rev4 CHANGES_REQUESTED).
Prior evidence: results.md (rev1), _rev2, _rev3, _rev4, rev1–rev4 CR patches
and validation logs, reviewer verdicts rev2/rev4. This resource supersedes
nothing; it appends the rev5 delta.

## Candidate and contract

- Story worktree base HEAD: `12f1287ee0fb538f9ca004dd53b870e824e5baf2`
  (unchanged; no commits by this producer — 26 changed/new paths,
  uncommitted, for the handoff snapshot).
- Spec: curator-spec main `871d11b`, protocol/skillfile-sources.md
  (revision 1 semantics; §2 physical boundaries, §5 diagnostics) and
  protocol/repository-transport.md revision 1. Diagnostics keep the exact
  spec §5 classes. Frozen v1 wire schemas untouched (no schema file in
  the diff; journal JSON layout unchanged — the per-write guard is
  in-memory only).
- Shell for all commands below: tool-managed `/bin/sh`; exit codes are
  the gate command's own (`${PIPESTATUS[0]}` where piped).

## What rework-3 ordered, and where it landed

Finding 2 (per-write recheck) — done, no stub-only proof:
- `internal/transaction/boundaries.go` (new): `BoundaryCheck` type —
  `func(targetIndex int, livePath string) error`, run immediately before
  each live mutation.
- `internal/transaction/types.go`: `Plan.BoundaryCheck` field (nil =
  legacy skip, byte-identical).
- `internal/transaction/engine.go`: engine retains the check at Prepare
  (`guards` map, same mutex); `commitTarget` consults it immediately
  before the backup rename (site 1) and before the install rename
  (site 2), after all verification of the bytes involved; refusal flows
  through the existing commit-error path (PhaseRollingBack, exact
  reverse restoration, journal removed). Rollback never consults the
  guard. `removeJournalDurably` (the single funnel for terminal removal:
  rollback end + committed cleanup) releases the entry.
- `internal/staging/boundaries.go`: `Recheck` refactored (pure code
  motion, identical refusal strings) into `recheckAdmitted` +
  `recheckTarget`, plus the single-target `RecheckOne` the engine
  reaches through the install closure.
- `internal/install/commit.go`: `journalPlan` sets
  `BoundaryCheck: boundaryCheckFor(sorted, targets.boundaries)` over the
  same `Sorted()` slice it journals; the closure fail-closes on index
  drift (journal live path vs staged live path, Abs-normalized like
  buildJournal). The pre-journal whole-plan `Recheck` stays as the
  coarse gate.
- Regression `TestPerWriteGuardRefusesSwappedParentAndRollsBack`
  (install, real engine + real journal + real `stageProjectTargets`):
  v1 committed; v2 commit swaps a destination parent via a
  `PointBeforeBackup` fault (rename + mkdir + move the original child
  back, so preimage verification passes and ONLY the identity guard can
  refuse); asserts `source_output_overlap`, earlier target rolled back
  to v1 bytes (proves restoration, not just refusal), refused target
  never published v2, fault hook provably fired, journal gone.

Finding 1a (adapter destination planning live) — done, ungated as ordered:
- `internal/install/targets.go`: `nodeSnapshots` — admitted inputs are
  the closure node snapshots (staging already reads them, so a missing
  one fails with `source_member_missing` before anything is derived).
- `internal/install/install.go` (`stageProjectTargets`) and
  `internal/install/global.go` (`stageGlobalTargets`, the second
  production adapter caller): populate `Group.Admitted` /
  `StageGlobal(..., admitted)` and `targets.attachBoundaries(admitted)`.
- `internal/adapters/stage.go`: `StageGlobal` takes `admitted`;
  stale "draft planning lands later" comments updated (adapters.go,
  stage.go, commit.go `scopeTargets.boundaries`).
- Production-entry tests, no canned scopeTargets: direct
  `stageProjectTargets`/`stageGlobalTargets` populate + refuse tests
  (symlink mirror into snapshot; project-inside-snapshot; missing
  snapshot); `runCommit` with the real staging function + stub journal
  (one journaled commit; refusal before journaling); full `Project()` /
  `Global()` e2e (install, relink one ledger-claimed mirror entry into
  the cache snapshot, reinstall refuses with `source_output_overlap`,
  snapshot bytes intact). The e2e positive control (first install
  succeeds with the guard live) proves no false refusal on the real path.

Finding 1b (local acquisition wired, draft-gated) — done:
- `internal/envprofile/envprofile.go`: `Policy.DraftSourcesV1`
  (default false; no machine-config surface today — the schema-2
  install-integration leaf owns the user-facing switch).
- `stateForPath` (the single funnel for both fresh path captures:
  `installLocked` root branch and `resolveOverlay` path branch, hence
  install/reinstall/update overlay funnels) runs `admitPathSource`
  first when the switch is on, before the store traverses anything.
- `admitPathSource` mapping: operand = source root at `.`
  (absolutized; operands forbid `--directory`), home = owning root
  (home-as-source needs root_inputs profiles do not carry, so refused),
  outputs = profile store + context store + profile git cache
  (`profileReposDir` helper split from `gitManager.reposDir`).
  Refusals return unwrapped so the spec §5 class leads;
  `preservePathDiag` is untouched, so switch-off is byte-identical.
- Tests through public `Install`: switch-off legacy pin (store-internal
  operand installs, exit 0), switch-on admit (root + overlay),
  switch-on refuse (root + overlay, nothing published), home-mapping unit.

## Validation (this producer, final candidate)

| Command | Exit | Evidence |
|---|---|---|
| `go test -count=1 -p 1 ./internal/snapshot/ ./internal/staging/ ./internal/privatedir/ ./internal/adapters/` | 0 | all ok (2.09s / 0.37s / 0.36s / 0.43s) |
| `go test -count=1 -p 1 ./internal/install/` (full) | 0 | ok, 238.6s — guard live on every install path, zero regressions |
| `go test -count=1 -p 1 ./internal/envprofile/` (full) | 0 | ok, 237.2s |
| `go test -count=1 -p 1 ./internal/transaction/` (full) | 0 | ok, 69.2s |
| install new-test set post-restore (11 tests) | 0 | 11/11 PASS, 12.5s |
| transaction boundary subset post-restore (4 tests) | 0 | ok, 3.3s |
| envprofile boundary subset post-restore (6 tests) | 0 | ok, 6.6s |
| adapters + staging full post-restore | 0 | ok (0.91s / 0.39s) |
| `go vet` on snapshot/staging/privatedir/adapters/install/transaction/envprofile | 0 | clean |
| `gofmt -l` on the same 7 packages | 0 | empty |
| `GOOS=windows go build` on 6 touched packages | 0 | cross-compiles |
| `GOOS=windows go vet` on transaction/install/envprofile | 0 | clean |
| `golangci-lint run` on the 7 packages | 0 | `0 issues.` (one revive unused-parameter in a new test fixed first) |

No new skip reasons: new skips reuse the verbatim
`symlinks unavailable: %v` string, which matches the declared
`host-capability` class in `.github/ci/skip-classes.tsv` (line 69).
The full landing suite was not run manually; the handoff owns the single
remote-gate execution. No installs, daemon restarts, tags, releases,
LOGBOOK.md edits (prohibited; findings live here), runtime-home changes,
live credential export, or ax calls. Board CLI: the default wrapper
stalls on this host (as verdict rev4 noted); all board writes used
`task-board-main-6cb09a23-curatorlike --no-update-check`.

## Measured negative evidence (this run)

Bytes snapshotted with `shasum -a 256` before mutation; each mutant
reverted and `shasum -c` verified OK after (4/4 files). Post-restore
suites rerun green (table above). Every mutant is a narrowing with a
passing sibling, not a deletion.

| # | Finding | Narrowing mutation | Killer (exit) | Sibling (exit) |
|---|---|---|---|---|
| M1 | 2 | engine site-1 check only when `index == 0` | `TestCommitBoundaryRefusalAtBackup…` exit 1 (`backedUp = [0 1]`) | install-site refusal PASS |
| M2 | 2 | engine site-2 check only for `KindEntry` | `TestCommitBoundaryRefusalAtInstall…` exit 1 (`err = <nil>`) | backup-site refusal exit 0 |
| M3 | 1a | `attachBoundaries` records nothing when `<2` admitted | `TestPerWriteGuard…` exit 1 (corruption diagnostic instead of boundary) | project e2e refusal exit 0 (planning gate independent) |
| M4 | 1b | `admitPathSource` outputs minus the context store | `TestDraftPathInstallRefusesStoreOverlap` exit 1 (`err = <nil>`) | overlay-profiles refusal exit 0 |
| M5 | 1a | `ValidateDestinations` candidates `{full}` only (location dropped) | `TestValidateDestinationsRefusesEntryInsideAdmitted` exit 1 (`err = <nil>`) | link/parent full-arm test exit 0 |

5/5 killed. Two honest survivals along the way, both fixed by
strengthening tests rather than left standing: M1 first survived
because the scripted guard also fires at install — the test gained
`PointAfterBackup` ordering observations that pin the refusal before
the backup rename; M5 first survived against an install-level test
because the adapter ledger trips the full arm first — the location arm
gained its own isolation test at the adapters level, and the
install-level test was rescoped to the nested-project shape it actually
proves. (One M5 draft did not compile — unused variable — and was
reworked before any evidence was claimed from it.)

## Coverage map (acceptance rows → production entry points)

- Canonicalize paths/symlinks/case: staging unit (prior) +
  `TestProject/GlobalInstallRefusesAdapterDestinationOverSnapshot`
  (symlink attack through `Project`/`Global`) +
  `TestPerWriteGuard…` (rename+mkdir identity swap through runCommit).
- Authored vs managed outputs: adapters unit (prior +
  `TestValidateDestinationsRefusesEntryInsideAdmitted`) +
  `TestStageProject/GlobalTargetsPopulatesBoundaries` (Admitted live).
- Operator root_inputs: snapshot unit (prior) +
  `TestAdmitPathSourceRefusesManagerHome` (home-as-source mapping).
- Overlap rejected before traversal: `TestDraftPathInstall/Overlay…`
  through `Install`; `TestStageProject…` through real staging;
  `TestRunCommitWithRealStaging…` (pre-journal Recheck).
- Recheck at each publication write: engine tests (both sites, real
  journal) + `TestPerWriteGuard…` (real journal, fault injection,
  rollback proven by v1 bytes).
- Unmanaged files: prior unmanaged tests + snapshot-intact assertions
  in every refusal test above.
- Path source `.` with safe subdir: prior
  `TestPrepareLocalAcquisitionAdmitsSafeSubdirectory` +
  `TestDraftPathInstall/OverlayAdmitsOrdinaryDirectory`.

## Bounds (stated, not deferred in-scope work)

- 16k7xy owns byte capture, hashing, race detection, and store
  publication consuming `LocalAcquisition`: no call site exists today
  (I searched; there is no file/function to name). The seam is
  `LocalAcquisition{Physical, Admitted, Files, Staging}` from
  `PrepareLocalAcquisition` in `internal/snapshot/boundaries.go`;
  `admitPathSource` uses it as a pre-traversal admission gate and the
  store re-walks the admitted tree itself.
- Crash recovery replays without the in-memory guard: planning-time
  `FileInfo` cannot round-trip through the journal, so a restarted
  process resumes commit/rollback without per-write rechecks. The
  original attempt checked before each write, and recovery still
  enforces per-target preimage/digest verification. Documented in
  `internal/transaction/boundaries.go`.
- Install-path admitted set = closure node snapshots only. Staged and
  canonical stores receive the run's own writes by design; admitting
  them would false-refuse.
- `ensureDefault`'s direct `EnsureState` captures manager-synthesized
  bytes (it writes `defaultManifest` itself), not an operator path
  source — outside the gate funnel by design. Dev-substitution
  local-path git clones are git acquisition (gitops), not path-source
  snapshot capture — outside the parenthetical scope.
- Carried bounds: NFD-alias blind spot (`staging.Within` doc),
  inode-reuse parity (rev4 results), Windows file-ID eager pin (rev4,
  kept; `pinIdentity` untouched this cycle).
- Checklist item 6 (review-routed branches) is left for the verdict;
  item 12's findings live in this resource because LOGBOOK.md edits
  are prohibited by campaign rules.
