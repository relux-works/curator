# TASK-260910-14hsti — Windows gate-fix handoff (rev6)

Recovery producer run (RUN-260916-487398, root RUN-260916-7fb858) after the
rev5 remote gate FAILED on Test (windows-latest) only (run 35101307516).
Prior evidence: results.md (rev1), _rev2…_rev5, rev1–rev5 CR patches and
validation logs, reviewer verdicts rev2/rev4. This resource appends the
rev6 delta only; the rev5 rework-3 record stands.

## Candidate and contract

- Story worktree base HEAD: `12f1287ee0fb538f9ca004dd53b870e824e5baf2`
  (unchanged; no commits by this producer — the rev5 tree plus the delta
  below, uncommitted, for the handoff snapshot).
- Delta vs rev5 (verified by diffing the worktree against
  `TASK-260910-14hsti_change-request_rev5.patch`): exactly
  `internal/staging/boundaries.go` (doc comment + 2 lines: `os.Stat` →
  `os.Lstat` at both `Within` identity-walk sites) and
  `internal/staging/boundaries_test.go` (+100 lines, pure append: 2 new
  tests). No other file touched this round.
- Spec: curator-spec main `871d11b`, protocol/skillfile-sources.md
  (revision 1 semantics; §2 physical boundaries, §5 diagnostics).
  Diagnostics keep the exact spec §5 classes. Frozen v1 wire schemas
  untouched.
- Shell for all commands below: tool-managed `/bin/sh`; exit codes are
  each gate command's own (stdout redirected to files, `echo exit=$?`
  — no pipes between the gate and its status).

## Root cause of the rev5 Windows failure

Remote gate run 35101307516: every lane green except Test
(windows-latest). The validation-log tail named 2 required platform
cases, but the run's `test-evidence-windows-latest` artifact
(`test/go-test-served.json`, downloaded and grepped for this diagnosis)
shows the true scope — 4 tests, one signature:

- `internal/install/atomicity ::
  TestFailureAtEveryTargetClassRestoresPriorStateInReverseOrder`
  (every project + global subtest): `commit_atomicity_test.go:233: the
  injected fault never fired`, `:250: install after the rollback sweep
  failed`
- `internal/install/atomicity ::
  TestAdapterMirrorLinksAreJournaledAndRestoredExactly` (auto+symlink):
  `:317: the injected fault never fired`
- `internal/install/atomicity ::
  TestStaleAdapterEntryIsRemovedBeforeTheConsumerLedger` (auto+symlink;
  copy passed): `:394: second install failed`
- `internal/install :: TestRemovedSkillCleanedUp`:
  `install_test.go:423` second install failed

Every failure carries the identical install error:

```
source_output_overlap: C:\Users\RUNNER~1\...\NNN\.claude\skills\<skill>:
CreateFile C:\Users\runneradmin\...\NNN\.claude\skills\<skill>: Access is denied.
```

Mechanism (all three links verified against Go 1.26 runtime source in
this host's GOROOT):

1. Adapter mirror entries are staged with `os.Symlink` carrying a
   destination string relative to the LIVE entry while the link is
   created in the STAGING tree, where it resolves to nothing
   (`internal/adapters/stage.go: stageLink` + `linkDestination`).
   Go's Windows `os.Symlink` stats the destination as resolved from
   the creation directory to choose the link flavor
   (`os/file_windows.go`: `isdir := err == nil && fi.IsDir()`), so
   every published mirror link is a FILE-flagged symlink pointing at
   a DIRECTORY.
2. `staging.Within`'s SameFile ancestor walk used `os.Stat`, which
   FOLLOWS the link: Go's Windows `stat()` reopens a reparse point
   without `OPEN_REPARSE_POINT` and opens the target as a directory,
   which fails with `ERROR_ACCESS_DENIED` through a file-flagged
   link. The `PathError{Op: "CreateFile"}` shape matches
   `os/stat_windows.go` exactly. The long-form (`runneradmin`) path
   in the error proves the failing `Stat` ran on the
   `EvalSymlinks`-resolved spelling — i.e. inside `Within`
   (`Canonicalize` expands the 8.3 `RUNNER~1` component; it is
   Readlink-based and succeeds, which is why only the walk failed).
3. The walk runs only when the canonical string check misses — i.e.
   for DISJOINT destinations — so every second install touching an
   existing mirror link was refused at planning
   (`stage.go:105 ValidateDestinations`, the first production `Within`
   caller; pre-journal `Recheck` and per-write `RecheckOne` funnel
   through the same walk). Baselines passed (entries absent → walk
   skips NotExist), copy mode passed (real dirs), unix passed
   (`Stat`-follow works there), and
   `TestStaleAdapterRemovalRollsBackToTheExactPriorEntry` passed on
   Windows DESPITE the bug (inference: it asserts `failed` +
   unchanged state + no journal without a probe assertion, so a
   pre-journal refusal satisfies it vacuously).

`Within` was the lone follower in the boundary code: `namespaceStat`
(Lstat for entries), `adapterEntryState` ("deliberately Lstat/Readlink
based"), `DigestTarget`, `ValidateDestinations`, `recheckTarget`, and
`walkPruned` all refuse to follow entry links already.

## Fix

`Within` (`internal/staging/boundaries.go`) walks the caller's own
spellings with `os.Lstat` at both sites (root + ancestry) instead of
`os.Stat`. Two lines; comment records the Windows file-link hazard.

Why Lstat-on-originals and not walking the resolved chains: my first
attempt walked the canonical chains, and the existing
`TestValidateDestinationsRefusesEntryInsideAdmitted` caught it the
same run — an entry whose LOCATION sits inside admitted inputs while
resolving outside them is refused through the location arm, whose
detection lives in the unresolved spelling's parent levels, which
resolution erases. Lstat on the originals keeps that arm (parent
levels are plain dirs; only the link level itself changes from
follow-identity to link-identity), while the link-target overlap arms
stay covered by the canonical string comparison plus the callers'
explicit `Readlink` checks (`recheckTarget`, `ValidateDestinations`).
Enumerated source links stay refused downstream (`walkPruned` rejects
any link it meets), so the pruning path cannot go quiet.

Behavior flips vs Stat, all analyzed: (1) file-link crash → correct
answer (the fix); (2) dangling-link case-alias MISS → HIT (genuine
improvement: Stat skips dangling levels via NotExist); (3) mid-path
link + case-mismatched root spelling TRUE → FALSE — unreachable in
production (install spellings are machine-consistent on both sides,
so the string check decides; local-source candidates are
pre-resolved before `Within`; `ValidateDestinations`/recheck chains
are link-free above the final component, whose target is covered by
other arms). (3) is a stated analyzed bound, not an unknown.

## Tests

- `TestWithinNeverFollowsCommittedMirrorLinks` (new): builds the exact
  production shape — link created against an absent target (unresolvable
  at creation ⇒ file-flagged on Windows), target published, link
  renamed into a live adapter root — and pins the disjoint answer
  forward (`Within(admitted, live)` → false, nil) and reverse
  (`Within(live, admitted)` → false, nil), plus a positive control
  (entry member within target → true). Pre-fix it errors on Windows
  (Stat-follow) and passes after; on unix the follow succeeds, so the
  Windows gate lane is its enforcing runner. The four existing
  Windows-failing tests are the production-entry regressions for the
  same hazard (no duplicate install-level test added).
- `TestWithinCatchesDanglingLinkCaseAlias` (new): case-variant
  spellings of one dangling link → (true, nil), plus a positive
  control (dangling link under its parent → true via the string
  check). Fails pre-fix wherever the volume folds case (proven
  locally by both mutants below); skips on case-sensitive Linux
  with `test filesystem is case-sensitive; dangling-link case alias
  needs a folding volume`, which matches the declared
  `host-capability` class (`test filesystem (is case-sensitive|…)`
  in `.github/ci/skip-classes.tsv`). Symlink-creation failure reuses
  the verbatim declared `symlinks unavailable: %v` reason. No
  skip-class file change needed.

## Validation (this producer, final candidate)

| Command | Exit | Evidence |
|---|---|---|
| `go test -count=1 -p 1 ./internal/staging/ ./internal/snapshot/ ./internal/privatedir/ ./internal/adapters/` | 0 | all ok (0.57s / 1.38s / 0.35s / 0.43s); includes the 2 new tests |
| `go test -count=1 ./internal/install/ -run TestRemovedSkillCleanedUp` | 0 | ok, 9.9s (Windows-failing test, local lane) |
| `go test -count=1 ./internal/install/atomicity/ -run 'TestAdapterMirrorLinks…\|TestStaleAdapterEntry…\|TestStaleAdapterRemoval…'` | 0 | ok, 50.6s (3 Windows-failing shapes, local lane) |
| `go test -count=1 ./internal/install/atomicity/ -run TestFailureAtEveryTargetClassRestoresPriorStateInReverseOrder` | 0 | ok, 229.1s (full fault sweep, 12 scenarios) |
| install boundary subset (12 tests: `TestRunCommit*\|TestStageProjectTargets*\|TestStageGlobalTargets*\|TestPerWriteGuard*`) | 0 | ok, 5.7s |
| install e2e guard pair (`TestProject/GlobalInstallRefusesAdapterDestinationOverSnapshot`) | 0 | ok, 8.5s |
| transaction boundary subset (4 tests) | 0 | ok, 3.4s |
| envprofile boundary subset (6 tests) | 0 | ok, 8.6s |
| `go vet` on snapshot/staging/privatedir/adapters/install/transaction/envprofile | 0 | clean |
| `gofmt -l` on the same 7 packages | 0 | empty |
| `GOOS=windows go build ./internal/staging/` | 0 | cross-compiles |
| `GOOS=windows go vet ./internal/staging/` | 0 | clean |
| `golangci-lint run ./internal/staging/` | 0 | `0 issues.` |

The full landing suite was not run manually; the handoff owns the
single remote-gate execution. No installs, daemon restarts, tags,
releases, LOGBOOK.md edits (prohibited; findings live here),
runtime-home changes, live credential export, or ax calls. Board
writes used plain `task-board` (the wrapper stall noted in rev5 did
not reproduce on this run).

## Measured negative evidence (this run)

`internal/staging/boundaries.go` snapshotted with `shasum -a 256`
before mutation; each mutant reverted and `shasum -c` verified OK
after. Post-restore staging suite rerun green (exit 0, above).
Both mutants are single-site narrowings with a passing sibling.

| # | Narrowing mutation | Killer (exit) | Sibling (exit) |
|---|---|---|---|
| M1 | `Within` root site back to `os.Stat` | `TestWithinCatchesDanglingLinkCaseAlias` exit 1 (`within = false, <nil>; want true, nil`) | mirror-links + symlink-alias tests exit 0 |
| M2 | `Within` walk site back to `os.Stat` | `TestWithinCatchesDanglingLinkCaseAlias` exit 1 (same assertion) | same pair exit 0 |

2/2 killed locally. The production-shape test
(`TestWithinNeverFollowsCommittedMirrorLinks`) survives both mutants
on unix by construction (the follow succeeds there) — its enforcing
runner is the Windows gate lane, where the pre-fix tree fails it
with the `CreateFile … Access is denied` refusal (proven by the
four rev5 Windows failures of the same hazard, not by assumption).

## Coverage map (rev6 delta → entry points)

- `ValidateDestinations` (planning, `adapters/stage.go:105` — the
  observed crash site): disjoint mirror entries decide without
  following, both directions — new mirror-links test + the 4
  existing Windows-failing tests through `StageProject`/install.
- Pre-journal `Recheck` and per-write `RecheckOne` (same `Within`
  funnel): unchanged refusal semantics — location-arm test
  (`TestValidateDestinationsRefusesEntryInsideAdmitted`,
  temporarily red during development, green now), full atomicity
  sweep, install/transaction boundary subsets.
- Case semantics: dangling-link case-alias identity now caught —
  new case-alias test (runs on macOS/Windows, declared-skip on
  case-sensitive Linux).

## Bounds (stated, not deferred in-scope work)

- All rev5 bounds carry over unchanged (16k7xy seam, crash-recovery
  guard absence, admitted-set definition, `ensureDefault`/git-clone
  funnels, NFD blind spot, inode-reuse parity, Windows eager pin).
- New: (3) above — mid-path link + case-mismatched root spelling
  flips TRUE → FALSE; unreachable in production per the per-callsite
  argument (machine-consistent install spellings, pre-resolved
  local-source candidates, link-free chains above the final
  component, explicit link-target arms elsewhere).
- New: the mirror-links regression test is Windows-enforced; this
  host (macOS) cannot execute the pre-fix failure mode. Local
  enforcement comes from the dangling case-alias test (both
  mutants) and the unchanged-behavior suites.
- Checklist item 12's findings live in this resource because
  LOGBOOK.md edits are prohibited by campaign rules.
