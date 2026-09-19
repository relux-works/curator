# TASK-260910-3eu4cy — publish source installs and locks atomically (handoff evidence)

Candidate: uncommitted working tree of `.temp/STORY-260910-1s75e1/worktree`
on top of checkpoint `ee21aec` (sibling TASK-260910-1xs0pj rev3).
Shell: bash; every gate below ran as a standalone process with a real exit
code (exit captured from the binary, never through a pipe). No commits on
the Story branch.

## Changed paths

```
M  cmd/curator/main.go
M  internal/closure/resolve.go
M  internal/install/global.go
M  internal/install/install.go
M  internal/registry/attest.go
M  internal/sourcelock/sourcelock.go
?? cmd/curator/draft_status.go
?? cmd/curator/draft_status_test.go
?? internal/closure/refresh_atomicity_test.go
?? internal/install/draftatomic_test.go
?? internal/install/draftpublish_test.go
?? internal/registry/attest_v5_test.go
?? internal/sourcelock/restore_test.go
```

Production delta is 145 insertions / 10 deletions across 6 files; 1 new
production file (166 lines) and 6 new test files (~1950 lines).

## What it implements (skillfile-sources §2–§4)

This leaf closes the Story: lock, marker, runtime, and adapters publish
as one recoverable transaction per operation with consistent rollback,
boundaries are rechecked at write time, and status/repair/refresh enforce
exact currentness over frozen inputs. It also closes the three identity
consumers 1xs0pj deferred here (scopeStatusDrift, AttestRoot v5,
detectMovedTagsIn via the lock member).

- `internal/sourcelock` (`Restore`): the rollback half of lock
  publication. Prior bytes are restored verbatim through the same
  temp-file, sync, and rename sequence as `Write` — never re-encoded.
- `internal/closure` (`restorePriorFile`): `RefreshDraft` rollback now
  uses the atomic `Restore` instead of a direct `os.WriteFile`, so a
  crash during rollback leaves prior or published bytes, never a torn
  file. A crash *between* the two publication renames leaves the new
  lock beside the old bindings; that generation mismatch fails closed
  (`source_lock_stale` via `Bindings.CheckFresh`) and the next explicit
  attempt re-publishes both files (documented at the call site).
- `internal/install` (`detectMovedTagsIn` + `recordedPackageCommit`): v5
  arm. Legacy identity fields are absent by construction, so the tag
  comparison runs through the locked package commit — the frozen node's
  resolved commit IS the lock member's commit. A tag the current lock
  resolved to a new commit warns exactly like a legacy moved tag, and
  StrictTags refuses it the same way. Local-snapshot arm never matches.
  Legacy markers take the untouched legacy branch.
- `internal/registry` (`AttestRoot` + `v5AttestIdentity`): v5 handling.
  Only network-git re-resolves, through its canonical repository and
  locked commit (the same identity install resolves registries with —
  never a URL parse, which would misclassify a canonical value as
  local). Configured-git is `unattestable` (a source path alone is not
  a registry identity); local-snapshot is `unattestable` (no network
  identity). Legacy path textually untouched.
- `internal/install` (`Result.Attestations`, set in project and global
  attempts): the registry evidence the effective plan selected, exposed
  for the read-only status comparison. Additive field; never authorizes.
- `cmd/curator` (`draft_status.go`, `main.go` wiring): draft status
  currentness. Schema-2 + switch takes the new lane; every other shape
  keeps the legacy surface byte-identically (pinned by test). The lane
  consumes the frozen lock only — no rescan, branch advance, or snapshot
  replacement — and compares, per lock member: installed presence
  (typed Lstat: absent vs unreadable), marker validity, content hash,
  frozen package digest, binding `lock_sha256`, substitution absence
  (the effective plan never substitutes on the draft lane), and recorded
  attestation vs the dry run's effective evidence. A stale lock yields
  per-skill `needs-install` rows (like a legacy unresolvable
  declaration) instead of an opaque error; an unreadable lock yields no
  verdict (nonzero, read-only). Compiled rows reuse the extracted
  `classifyScopeBuilds` (pure extraction; `statusReport` signature and
  behavior unchanged — legacy callers and tests untouched).

No production change was needed in `internal/transaction`,
`internal/adapters`, `internal/runtimestore`, or `internal/staging`:
the journaled commit, per-write boundary guard, unmanaged-destination
refusal, and source-v1 materialization already implement the contract;
this leaf proves their draft-lane integration at the production entries
(see tests). That is the "extend, don't re-implement" outcome of the
inspection the task requires.

## Design notes for the reviewer

- "One recoverable transaction": resolve (lock+bindings) and install
  (marker+runtime+adapters) are separate CLI invocations, so each is
  one recoverable publication with consistent rollback, and the two are
  bound by the lock generation: mismatched generations fail closed
  (`source_lock_stale`), never reinterpreted; retry re-publishes.
- Content-only mutation of an admitted input *after* staging is benign
  by design: staged bytes were already frozen pre-mutation, so the
  commit publishes a self-consistent snapshot, and the next open
  rehashes and refuses. The write-time guard pins admitted *identity*;
  the admitted-replacement test proves a swapped input refuses with
  `source_output_overlap` and full rollback.
- The retarget test's parent swap moves the journal's own
  `.curator-txn-*` sidecars out of the engine's view and back; the test
  asserts they are the only difference (the stale scan skips dotfiles,
  so they are never installed state), sweeps exactly that prefix, and
  proves the live state byte-identical.

## Tests (production entries named)

Production entry `install.Project` + `closure.RefreshDraft`
(`internal/install/draftatomic_test.go`,
`internal/install/draftpublish_test.go`):

- `TestDraftFailureAtEveryTargetClassRestoresPriorState` (+/−, 7 rows):
  fault at each of context/runtime/shim-canonical/env-file/
  adapter-ledger/removal/consumer → failed, whole-state digest
  (lock+bindings+markers+runtime+shims+env+adapters+consumers)
  byte-identical, reverse-order rollback, no journal; post-sweep
  success commits all 7 classes in order and binds the new lock.
- `TestDraftMovedTagWarnsThroughTheLock` (+/−): same-tag move warns
  `moved tag for review: v1 C1 -> C2` with exact commits; StrictTags
  refuses; install then binds the moved commit.
- `TestDraftRepairRevalidatesLockedSource` (−): tampered frozen
  snapshot → `source_snapshot_changed`; lock and marker byte-identical.
- `TestDraftRepairRestoresDriftedContent` (+): drifted bytes repaired
  from the frozen lock; lock bytes identical; package still bound.
- `TestDraftRefusedRepairPreservesLock` (−): revoked evidence refuses;
  lock and attested marker byte-identical.
- `TestDraftInstallFailsStaleBindingsGeneration` (−/+): new lock + old
  bindings (simulated crash) → `source_lock_stale`, install writes no
  lock bytes, marker stays on v1; next refresh recovers, install binds v2.
- `TestDraftRefreshThenInstallReplacesLockAndMarker` (+/−): refresh
  publishes v2; failed install keeps the v1 marker beside the v2 lock;
  success binds v2. Neither file moves before its own success.
- `TestDraftRetargetFailsAtWriteTimeAndRollsBack` (−): parent replaced
  after the pre-journal recheck → `source_output_overlap` from the
  per-write guard, full rollback, no journal.
- `TestDraftAdmittedReplacementFailsAtWriteTimeAndRollsBack` (−):
  admitted input replaced after staging → `source_output_overlap`,
  full rollback, no journal.
- `TestDraftUnmanagedAdapterDestinationRefuses` (−): foreign file at a
  mirror path → `adapter target already exists and is not managed`;
  foreign bytes, lock intact, nothing staged live.

Production entry `closure.RefreshDraft` / `sourcelock`
(`internal/closure/refresh_atomicity_test.go`,
`internal/sourcelock/restore_test.go`):

- `TestRefreshDraftFirstWriteFailurePreservesPriorFiles` (−): lock
  rename cannot stage → prior lock+bindings byte-identical, no temp
  files, prior snapshot still opens. (POSIX permission vector with the
  declared host-capability skip probe.)
- `TestRefreshDraftCrashedGenerationFailsClosedAndRetryRecovers`
  (−/+): new lock + old bindings → `source_lock_stale`; next attempt
  re-publishes both; recovered generation fresh and stored.
- `TestRestoreReturnsExactPriorBytes` (+): rollback bytes verbatim, no
  temp files.

Production entry `registry.AttestRoot`
(`internal/registry/attest_v5_test.go`):

- Network-git audited (+) / revoked (−, `HasRevocation`) through the
  package; local + configured-git `unattestable` with exact details
  (−); legacy marker unchanged (+); malformed shapes refused at the
  helper seam (4 rows).

Production entry `curator status` (`cmd/curator/draft_status_test.go`):

- CLI: up-to-date + `--check` 0 + JSON shape + read-only proof
  (project and home digests identical around all three checks);
  needs-install after refresh-without-install (`--check` 1, then
  install → current); stale manifest → needs-install (`--check` 1);
  content-drift; not-installed; live-mutation-without-refresh stays
  current (frozen inputs); invalid marker.
- `TestClassifyDraftMemberComparesExactCurrentness`: 12 rows — current
  (plain + attested), package mismatch (both arms), lock mismatch,
  attestation registry/status/key/absent-in-either-direction,
  substituted.
- Presence rows: legacy marker → needs-install; content-drift;
  not-installed; unreadable parent → unresolvable (declared skip
  probe). Diversion guard: only schema-2 + switch diverts. Missing
  lock → no verdict (fail closed).

## Narrowing mutants (each restored; tree verified identical after)

| id | mutant | killer | result |
|---|---|---|---|
| M1 | status compares attestation to itself | attestation registry/status/key/recorded-absent rows | KILLED exit 1 |
| M2 | status compares package digest to itself | package-mismatch (both arms) | KILLED exit 1 |
| M3a/b | per-write `BoundaryCheck` removed (proof kept) | retarget + admitted tests | SURVIVED exit 0, **equivalent**: the durable proof covers the same-process commit — redundant defense layer |
| M3c/d | M3 + durable `BoundaryProof` removed | retarget + admitted tests | KILLED exit 1 (install wrongly succeeds) |
| M4 | moved-tag v5 arm removed | `TestDraftMovedTagWarnsThroughTheLock` | KILLED exit 1 |
| M5 | v5 network-git forced `unattestable` | audited + revoked + helper rows | KILLED exit 1 |
| M6 | `restorePriorFile` no-op | both pre-existing second-write tests | KILLED exit 1 |
| M7 | `draftLockIsStale` forced false | stale-manifest CLI test (gets error, not rows) | KILLED exit 1 |
| M8 | substituted check narrowed to NUL | substituted row | KILLED exit 1 |

No genuine survivor: M3a/b is the documented redundant-layer shape
(also seen on 1xs0pj), demonstrated by the combined kill.

## Independently executed checks (real exit codes)

Build/vet/lint:

- `go build ./...`-equivalent over touched trees: exit 0
- `go vet` cmd/curator + install + closure + registry + sourcelock: exit 0
- `gofmt -l cmd internal`: no output; `git diff --check`: clean
- `golangci-lint run` over the 5 touched trees: exit 0, 0 issues
  (one QF1001 De Morgan finding fixed during the run)

New tests (prebuilt binaries, `-count=1`):

- install sweep `TestDraftFailureAtEveryTargetClassRestoresPriorState`: exit 0 (87.9s)
- install publish rows (9 tests, one `-run` each): exit 0
  (0.2s–24.3s each)
- closure `TestRefreshDraft*` (5 incl. 3 pre-existing): exit 0
- sourcelock `TestRestoreReturnsExactPriorBytes`: exit 0 (whole package: exit 0)
- registry attest rows (5 tests): exit 0 (whole package: exit 0)
- cmd classifier + presence + guards: exit 0; CLI status rows (7 tests): exit 0

Legacy regression (prebuilt binaries, `-count=1`):

- cmd legacy status surface (8 tests) + 5 re-pointed CLI resolve tests: exit 0
- install `TestLegacyInstallUntouchedWhenDraftOff`, `TestMovedTagWarningAndStrict`,
  `TestGlobalStrictTagsDetectMovedTag`: exit 0
- install commit/rollback suite (12 `-run` patterns): exit 0
- install `TestDraft*` (all 67 incl. wave-1–3 suites): exit 0 in 5
  sequential chunks (16+13+9+14+5). One combined single-run attempt
  timed out at 9m50s under host contention (load ~9, stuck git child
  in fixture setup — setup code this leaf does not touch); rerun in
  chunks is green, including the test that was stuck.
- closure `TestRefreshDraft*` (pre-existing 3) + `TestOpenDraftFrozen*` (4): exit 0

## Bounds (not findings)

- Whole-repository suite, race lanes, and Windows/Ubuntu platform rows
  are left to the hosted gate on the exact candidate tree (this host
  runs narrow packages only per the wave note). Two new tests carry
  POSIX permission vectors behind declared host-capability skip probes
  (`this process can write through a read-only directory`,
  `this environment can read a mode-000 directory`); no new skip class.
- Draft `status --attest` rendering path is unchanged and covered at
  the `AttestRoot` entry; no separate CLI attest test was added.
- `Result.Attestations` for the global scope is set for consistency
  but consumed only by the project draft status (global has no draft
  lane); `status --attest` for configured-git stays conservatively
  `unattestable` (no canonical identity in the marker).
- The combined single-process `TestDraft*` timeout above is
  environmental (host contention); evidence is the green chunked rerun.
# Revision 2 appendix (rework 1: review rev1 → revision 2)

Candidate: same uncommitted working tree, checkpoint `ee21aec` unchanged.
Shell: bash; exit codes captured from the binary (redirect to file, `echo
$?`), never through a pipe. No commits on the Story branch. Everything
the reviewer confirmed (transactional publication, per-write recheck,
refresh/frozen inputs, unmanaged-file protection, retarget/copy failure
paths) is untouched; only F1–F4 below changed.

## Rulings applied (orchestrator, binding)

- F1 → OPTION A. Dropped the v5 arm of `detectMovedTagsIn`
  (`internal/install/install.go`); `recordedPackageCommit` deleted. The
  legacy v4 predicate (`recorded.Ref == node.Resolved.Ref`) is textually
  unchanged. Call site + gate state the bound: on the draft lane install
  never re-resolves refs, so a tag move is observable only at explicit
  refresh.
- F2 → RECOMMENDED FIX. Removed the stale-lock diversion in
  `cmd/curator/main.go` (stale lock takes the existing failed-dry-run
  path: `printResult`, refusal printed, exit non-zero);
  `draftLockIsStale` deleted; `draftStatusDrift` returns the `CheckStale`
  error, never rows.

## Changed paths (revision 2 delta on top of revision 1)

```
M  cmd/curator/main.go            (diversion removed, comment updated)
M  cmd/curator/draft_status.go    (stale helper deleted, stale→error)
M  cmd/curator/draft_status_test.go (stale tests re-pointed, +5 tests)
M  internal/install/install.go    (v5 arm dropped, bound comments)
M  internal/install/draftpublish_test.go (moved-tag test → 2 negative rows + legacy control + F3 row)
```

No other file touched; no new files; no new skip class.

## Tests (all at the production entry)

F1 — `install.Project` (`internal/install/draftpublish_test.go`):

- `TestDraftDeclaredTagBumpIsNotAMovedTag` (−): declared bump v1→v2
  (v1 asserted untouched) → StrictTags ok, no "moved tag" in advisory
  or strict messages, marker binds v2. Adapted from reviewer probe A.
- `TestDraftSameTagMoveInstallsWithoutWarning` (−): same-tag move →
  refresh binds the new commit in the lock; install under StrictTags
  ok, no warning, marker binds the moved commit.
- `TestLegacyDeclaredTagBumpIsNotAMovedTag` (control): same bump on
  the frozen v1 lane passes StrictTags (reviewer probe A legacy half).

F2 — `curator status` (`cmd/curator/draft_status_test.go`):

- `TestDraftStatusStaleManifestRefuses` (re-pointed): stale lock →
  `source_lock_stale` on stderr, exit non-zero, no per-member rows;
  `--check` non-zero.
- `TestDraftStatusStaleEmptyLockRefuses` (zero-member row): resolve
  with `"skills":[]` (succeeds through the CLI, lock has 0 members),
  then declare `review` → refusal + `--check` non-zero. No skip.
- `TestDraftStatusStaleSwappedSelectionRefuses` (swapped row): lock
  {review}, manifest `include:["other"]` → refusal mentions the stale
  lock, stdout never says "review needs-install"; `--check` non-zero.
- `TestDraftStatusDriftFailsClosedOnStaleLock` (unit): stale lock is
  an error from `draftStatusDrift`, never rows.

F3 — `install.Project` dry run (`internal/install/draftpublish_test.go`):

- `TestDraftDryRunExposesResolvedAttestations`: dry run carries the
  injected `ResolveAttest` map on `Result.Attestations`
  (`reflect.DeepEqual`); failed resolution → failed + nil. Kills MR7.

F4 — `curator status` (`cmd/curator/draft_status_test.go`):

- `TestDraftStatusLockOnlyChangeNeedsInstall`: widen include to
  `["review","other"]`, refresh, no install → `review needs-install`,
  `--check` non-zero (reviewer probe D adopted verbatim in shape).
  Kills MR1 at the CLI.
- `TestDraftStatusLegacyMarkerNeedsInstall` (optional row): v4 marker
  in a draft project → `needs-install`, `--check` non-zero. Kills MR8
  at the CLI.

## Narrowing mutants (each reverted; `grep` for mutant residue clean)

| id | mutant | killer | result |
|---|---|---|---|
| MR1 | status drops the `lock_sha256` comparison | `TestDraftStatusLockOnlyChangeNeedsInstall` (CLI) | KILLED exit 1 |
| MR7 | `result.Attestations = nil` (`install.go:532`) | `TestDraftDryRunExposesResolvedAttestations` | KILLED exit 1 |
| MR8 | legacy marker in a draft project reported current | `TestDraftStatusLegacyMarkerNeedsInstall` (CLI) | KILLED exit 1 |
| M-v5 | v5 moved-tag arm restored | `TestDraftDeclaredTagBump…` + `TestDraftSameTagMove…` (both) | KILLED exit 1 |
| M-div | stale-lock diversion restored (both halves) | all three `…Stale…Refuses` CLI rows | KILLED exit 1 |

Commands (each `go test -p 1 <pkg> -run <mask> -count=1
-timeout=400s`, exit from the binary): MR1 `./cmd/curator`
`-run TestDraftStatusLockOnlyChangeNeedsInstall` → 1 ("review
up-to-date" under the mutant); MR7 `./internal/install` `-run
TestDraftDryRunExposesResolvedAttestations` → 1 (`Attestations =
map[]`); MR8 `./cmd/curator` `-run
TestDraftStatusLegacyMarkerNeedsInstall` → 1 ("review up-to-date");
M-v5 `./internal/install` `-run
'TestDraftDeclaredTagBumpIsNotAMovedTag|TestDraftSameTagMoveInstallsWithoutWarning'`
→ 1 (both FAIL); M-div `./cmd/curator` `-run
'TestDraftStatusStaleManifestRefuses|TestDraftStatusStaleEmptyLockRefuses|TestDraftStatusStaleSwappedSelectionRefuses'`
→ 1 (all three FAIL). Post-mutant `grep -rn
"mutantStale|recordedPackageCommit|draftLockIsStale" cmd/ internal/`:
no output.

## Independently executed checks (real exit codes)

- `go test -p 1 ./cmd/curator -run
  'TestDraftStatus|TestProjectResolve|TestClassifyDraftMember' -count=1
  -timeout=400s` → exit 0 (132.0s)
- `go test -p 1 ./internal/install -run
  'TestDraft|TestLegacyInstallUntouchedWhenDraftOff|TestReview'
  -count=1 -timeout=550s` → exit 0 (479.0s)
- New-row spot runs (all exit 0): F1 triple (33.6s), F3 dry-run
  (0.9s), F2 stale quadruple (8.5s), F4 pair (7.5s)
- `TestLegacyDeclaredTagBumpIsNotAMovedTag` (outside the prescribed
  mask) → exit 0, in the F1 triple run above
- Legacy gate rows `TestMovedTagWarningAndStrict`,
  `TestGlobalStrictTagsDetectMovedTag`,
  `TestLegacyInstallUntouchedWhenDraftOff` → exit 0 (15.7s)
- `go vet ./cmd/curator ./internal/install` → exit 0
- `gofmt -l cmd internal` → no output (exit 0); `git diff --check` → clean
- `golangci-lint run ./cmd/curator/... ./internal/install/...` → exit 0,
  0 issues

## Bounds (not findings)

- The optional F3 `curator status` row through a signed httptest
  registry was not added: no CLI test configures registries today, and
  the wiring is now covered at both production seams — exposure at
  `install.Project` (`TestDraftDryRunExposesResolvedAttestations`,
  kills MR7) and comparison at `classifyDraftMember` (12-row
  `TestClassifyDraftMemberComparesExactCurrentness` incl.
  registry/status/key/absent-both-directions, all green in the suite
  above).
- Whole-repository suite, race lanes, Windows/Ubuntu rows: hosted gate
  on the exact candidate tree only. No new skip class; no fixture
  changes on that axis.
# Revision 3 appendix (rework 2: Windows hang in the rev2 hosted gate)

Candidate: same uncommitted working tree, checkpoint `ee21aec` unchanged,
plus one new test file and one scratch removal (below). Shell: bash; exit
codes captured from the binary (redirect to file, `echo $?`), never through
a pipe. No commits on the Story branch. All semantics accepted so far
(transaction, rollback on injected faults, per-write recheck, stale-lock
refusal, F1/F2 rulings) are untouched: this revision adds no production
change, because the audit below proves rework 1 added no construct that
could hang.

## What the rev2 gate actually shows (forensics, not inference)

Gate run 35418260010 (rev2): ubuntu/macos Test + Race GREEN; Windows Test
FAILED with `panic: test timed out after 1h0m0s` in `internal/install`.
Gate run 35413340404 (rev1): all lanes green, Windows Test in 52m56s.
Evidence: the gate's own `go-test-served.json` streams (downloaded via
`gh run download`, artifacts `test-evidence-windows-latest` for both runs)
and the two Change Request patches. Findings:

1. The victim did not hang. The timeout dump lists
   `TestAuthoritativeCacheRejectionsAreRebuiltNeverAdopted (1m13s)` and
   subtest `cache-wrong-target (5s)` — the subtest had run 5 seconds when
   the package-level 1h alarm fired. The stream shows the parent's first 8
   of 16 subtests PASSed in rev2 at 6.6–13.0s each (rev1: 4.8–12.2s);
   `cache-wrong-target` is 9th of 16 alphabetically, not last — the alarm
   landed mid-suite while the test was progressing normally.
2. The stuck goroutine was making progress, not deadlocked. Its stack is a
   normal commit frame — `install.Project → projectAttempt → runCommit →
   Engine.Commit → commitTarget → saveJournal → validateJournal →
   validateIndependentTargetNamespaces → canonicalNamespacePath →
   filepath.EvalSymlinks → os.Lstat` — sampled `[runnable]`. No mutex,
   channel, lock-file, or lease wait appears in any product frame.
3. The whole Windows suite ran ~1.2–1.35x slower in rev2 than rev1 on every
   package, including packages this leaf does not touch: install
   2894s → 3600s+ (timeout), cmd/curator 1408s → 1858s (+32%),
   install/atomicity 1797s → 2258s (+26%, zero leaf files). Per-test
   Elapsed grew uniformly (e.g. the sweep
   `TestDraftFailureAtEveryTargetClassRestoresPriorState`, byte-identical
   code in both revs: 1257s → 1529s, +272s). Meanwhile macOS rev2 ran
   FASTER than rev1 (install 668s → 547s) and ubuntu mixed
   (136s → 196s). A code-introduced hang cannot slow packages with zero
   changed files, and cannot run faster on macOS; a slower/noisier Windows
   runner explains all three lanes. The suite was already at 48 of the 60
   package-timeout minutes in rev1 — any +25% runner noise tips it over.
4. Rework 1's production delta is provably hang-neutral. Diffing the rev1
   and rev2 patches file-by-file: exactly 5 files differ, of which 3 are
   production, all subtractive — `internal/install/install.go` (F1: the
   v5 moved-tag arm and `recordedPackageCommit` DELETED; the legacy
   predicate textually unchanged, so the legacy lane the victim runs
   executes identical code), `cmd/curator/main.go` + `draft_status.go`
   (F2: the stale-lock diversion and `draftLockIsStale` DELETED, stale →
   the existing refusal path; a different package from the victim). No
   loop, retry, lock/lease wait, or rename/remove/replace sequence was
   added in rework 1 — there is nothing to bound.

## Construct audit (rework-2 item 1)

- `internal/transaction`, `internal/adapters`, `internal/runtimestore`:
  the leaf changed nothing there (no such files in either patch); the
  engine loops (`staging.go` read-until-EOF, `namespace.go`
  walk-to-root with a `parent == prefix` exit) are trunk code, bounded,
  untouched.
- `runWithRestarts` (`install.go:183`, in the victim stack): already
  bounded — `tries >= commit.MaxRestarts` returns a reported failure,
  never an endless loop. Trunk code, untouched.
- `sourcelock.Restore` → `writeFileAtomic`: `tmp.Close()` precedes
  `os.Rename` (handle closed before replace — the Windows-safe order);
  no retry loop. Unchanged since rev1.
- `Result.Attestations` assignments (`install.go:532`,
  `global.go:225`): plain map stores. No waits.
- CLI prompt `for {}` loops (`buildhttpsprompt.go`,
  `buildsshprompt.go`) read stdin to EOF/error in interactive flows only;
  unreachable from the headless gate.

Conclusion: no Windows-only deadlock or unbounded wait exists in the
leaf's code on the rebuild path. The rev2 Windows failure is runner
slowness crossing a fixed 60m package timeout, not a product hang. No
production change, no skip.

## Changed paths (revision 3 delta on top of revision 2)

```
?? internal/install/cache_rebuild_deadline_test.go   (new, committed)
!! internal/install/zz_probe_test.go                 (deleted: stray scratch)
```

- The new file is the bounded regression (below). No new skip class, no
  fixture changes, no production edit.
- `zz_probe_test.go` ("Temporary probe", never in the rev1/rev2 patches)
  was stray scratch in the worktree from a prior attempt, not a deliverable;
  removed so it cannot pollute the rev3 candidate.

## Bounded regression (rework-2 item 3)

`TestCacheWrongTargetRebuildCompletesWithinDeadline`
(`internal/install/cache_rebuild_deadline_test.go`, production entry
`install.Project` via `env.install`): self-contained — no
`CURATOR_CONFORMANCE_ROOT`, no skips — so it runs on every OS in the gate
and locally (the authoritative test skips without the conformance root).
Two subtests, one per non-reusable boundary verdict (`corrupt`,
`untrusted-provenance`, the exact set `nonReusableStatuses` the
authoritative suite assigns by position). Each drives the full
refused-entry → private-rebuild flow with the authoritative assertions
(`ok`, never `BuildCacheHit`, rebuilt exactly once as `alpha`, refused
artifact never adopted, receipt bound to the rebuilt key) inside
`installWithWatchdog`: the install runs in a dedicated goroutine and the
test fails fast with a diagnostic when it does not return within
`cacheRebuildWatchdog = 180s` (~14x the slowest 13.0s gate sample, ~20x
faster than the lane's hour).

Watchdog proofs (temporary edits, each reverted; residue `grep` clean):

- Stall injection: `builder.observe = func(StageRequest) { select {} }`
  (block forever in the fake build seam) + watchdog shortened to 5s →
  exit 1 in ~9s with `install did not return within 5s; refusing to wait
  out the package timeout`. The bound trips on a hang.
- Narrowing mutant (removes the bound): the `<-timer.C` arm replaced by
  an unbounded `<-done`, stall kept → `panic: test timed out after 1m0s`
  under `go test -timeout 60s`, exit 1. Without the bound the hang
  consumes the timeout; the watchdog is the fail-fast mechanism.

## Independently executed checks (real exit codes, macOS)

- New regression `TestCacheWrongTargetRebuildCompletesWithinDeadline`
  (standalone, `-count=1`): exit 0, 14.2s — 2/2 subtests
  (`corrupt` 7.1s, `untrusted-provenance` 7.1s).
- Watchdog proof (temporary injected `builder.observe` hang + 5s
  watchdog): exit 1 in ~9s with `install did not return within 5s;
  refusing to wait out the package timeout`.
- Narrowing mutant (watchdog arm removed, hang kept): exit 1 via
  `panic: test timed out after 1m0s` under `go test -timeout 60s`.
  Both temporary edits reverted; residue `grep` clean.
- Install mask in 5 bounded chunks (union equals the prescribed
  `TestAuthoritativeCacheRejections|TestDraft|TestLegacyInstallUntouchedWhenDraftOff`):
  conformance+legacy exit 0 (3.5s; authoritative SKIPs locally without
  `CURATOR_CONFORMANCE_ROOT`, legacy PASSes); leaf publish 11 exit 0
  (213.4s); sweep exit 0 (129.8s, 7/7 classes); sibling halves exit 0
  (423.8s + 66.8s).
- `go test -p 1 ./cmd/curator -run 'TestDraftStatus|TestProjectResolve'
  -count=1 -timeout=550s`: exit 0 (184.1s).
- `go vet ./internal/install ./cmd/curator`: exit 0.
- `gofmt -l cmd internal`: no output; `git diff --check`: clean.
- `golangci-lint run ./internal/install/...`: exit 0, 0 issues.

## Bounds (not findings)

- Windows is verified by the hosted gate only (this host is macOS). The
  regression carries no skip and uses only helpers the authoritative
  Windows rows already exercise (fake toolchain/cache/builder, `git init`
  fixtures), so it runs on all three lanes.
- `TestAuthoritativeCacheRejectionsAreRebuiltNeverAdopted` SKIPs locally
  (`CURATOR_CONFORMANCE_ROOT` unset); the gate runs it. The new
  self-contained regression is the always-on counterpart.
- Whole-repository suite and race lanes: hosted gate on the exact
  candidate tree only.
