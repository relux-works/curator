# TASK-260910-14hsti — rework-4 handoff (rev7)

Rework-4 producer run after verdict rev6 CHANGES_REQUESTED (F1 HIGH
restart/recovery bypass, F2 MEDIUM inspection failures mislabeled as
overlap). Prior evidence: results.md (rev1), _rev2…_rev6, rev1–rev6 CR
patches and validation logs, verdicts rev2/rev4/rev6, probes rev6. This
resource appends the rev7 delta only; rev6 stands except where fixed
below.

## Candidate and contract

- Story worktree base HEAD: `12f1287ee0fb538f9ca004dd53b870e824e5baf2`
  (unchanged; no commits by this producer — the rev6 tree plus the
  delta below, uncommitted, for the handoff snapshot).
- Delta vs rev6: `internal/staging/boundaries.go` (F2 unreadable +
  DurableSnapshot/Durable/RecheckDurableOne), new
  `internal/staging/identity_unix.go` + `identity_windows.go`
  (FileIdentity), `internal/adapters/boundaries.go` (F2),
  `internal/snapshot/boundaries.go` (F2 + sameFile fail-closed),
  `internal/transaction/types.go` + `engine.go` + `journal.go` +
  `boundaries.go` (F1 durable proof), `internal/install/commit.go`
  (F1 wiring), plus 8 new/extended tests. 28 worktree paths total
  (26 rev6 + 2 identity files). No other file touched.
- Spec: curator-spec main `871d11b`,
  protocol/skillfile-sources.md (revision 1 semantics; §2 physical
  boundaries, §5 diagnostics) and protocol/repository-transport.md
  rev1. Proven overwrites keep exact spec §5
  `source_output_overlap`; inspection failures use the
  rework-authorized `boundary_identity_unreadable`, never overlap.
  Frozen v1 wire schemas untouched (journal `boundary_proof` is
  omitempty; legacy journals without it validate unchanged).
- Shell for all commands below: tool-managed `/bin/sh`; piped gates
  use `set -o pipefail` so the exit is the gate's own.

## F1 — restart/recovery re-verifies the durable proof

Journal is now distinguishable: `Journal.BoundaryProof` (`nil` =
legacy unguarded, non-nil = protected). `Plan.BoundaryProof` carries
`staging.DurableSnapshot` (same canonical spellings as Snapshot with
durable tokens: Dev:Ino on unix via `Sys().(*syscall.Stat_t)`,
Volume:Hi:Lo on Windows via CreateFile with BACKUP_SEMANTICS without
OPEN_REPARSE_POINT + GetFileInformationByHandle, matching Stat).
`buildJournal` deep-copies proof; Check-without-Proof journals an
empty protected marker (`&DurableSnapshot{}`) that fails closed on
recovery, never legacy. `checkBoundary` uses the in-memory guard in
the same process, else re-verifies via `staging.RecheckDurableOne`
per remaining write; nil proof skips (legacy), empty proof
(`len(Targets)==0`) or unrestorable proof fails closed as
`boundary_identity_unreadable` with safe rollback (no further
publication). `validateBoundaryProof` enforces absolute-clean-valid
paths, non-empty tokens, and proof-targets-subset-of-journal; empty
is valid (protected-but-unrestorable). `install.attachBoundaries`
builds Snapshot+Durable together; `journalPlan` sets both
BoundaryCheck and BoundaryProof. No hook to forget: transaction
imports staging directly, so install, envprofile (`lock.go`), and
`cmd/curator` gc recovery all enforce without wiring changes.

## F2 — inspection failures are never overlap

`boundary_identity_unreadable` for every Canonicalize/Stat/Lstat/
Readlink/Within/pinIdentity failure in staging Snapshot/Recheck/
RecheckOne/RecheckDurableOne, adapters ValidateDestinations, and
snapshot Validate/Enumerate/coverage/disjoint (including
`covered` propagation, disjoint fail-closed on non-NotExist, and
`sameFile` returning errors instead of false). Proven changes stay
`source_output_overlap`: retargeted, changed since planning,
vanished (`no longer exists`), is-now-a-link, overwrites admitted
input, not-recorded-at-planning, drift. `recheckIdentity` and
`verifyDurableIdentity` return `(unreadable, err)` so callers prefix
correctly with live-path context preserved.

## Tests (committed, production entry points)

- F1 trio (`internal/transaction/boundaries_test.go`, real
  Engine.Prepare + fresh Engine.Recover, same-spelling swap with
  children moved back so bytes/names identical):
  `TestRecoveryMustRecheckPhysicalBoundary` (refuses with overlap,
  live stays old, journal gone),
  `TestRecoveryWithUnchangedBoundaryPublishes` (succeeds, live new),
  `TestRecoveryLegacyJournalSkipsGuard` (legacy publishes across
  swap, proving no fail-closed regression).
- F2 (`EvalSymlinks: too many links` via self-referential parent,
  asserting unreadable and NOT overlap):
  `TestPlanRecheckIdentityReadErrorIsNotOverlap` and
  `TestRecheckDurableOneIdentityReadErrorIsNotOverlap` (staging),
  `TestValidateDestinationsIdentityReadErrorIsNotOverlap`
  (adapters), `TestValidateLocalPackageIdentityReadErrorIsNotOverlap`
  (snapshot).
- Durable parity: `TestRecheckDurableOneMatchesRecheck` (unchanged
  passes, swapped parent refuses with overlap+identity gate).
- Wiring: `TestStageProjectTargetsPopulatesBoundaries` and global
  twin now assert `Durable.Targets` non-empty.
- Existing suites unchanged in expectation (proven-overlap rows
  still overlap): staging/snapshot/privatedir/adapters full,
  install boundary pattern incl. PerWriteGuard + e2e guard pair,
  envprofile draft/legacy, transaction boundary quartet.

## Validation (this producer, final candidate)

| Command | Exit | Evidence |
|---|---|---|
| `go test -count=1 ./internal/staging/ ./internal/snapshot/ ./internal/privatedir/ ./internal/adapters/` | 0 | ok 0.385s/2.139s/0.432s/1.641s, incl. 6 new tests |
| `go test -count=1 -p 1 ./internal/transaction/ -run 'TestCommit.*Boundary\|TestCommitRollbackNever\|TestRecovery'` | 0 | 9/9 PASS incl. F1 trio (7.656s) |
| `go test -count=1 -p 1 ./internal/install/ -run 'Test(StageProjectTargets\|StageGlobalTargets\|RunCommit.*(Boundar\|Recheck\|RealStaging)\|PerWriteGuard\|ProjectInstallRefusesAdapter\|GlobalInstallRefusesAdapter)'` | 0 | all PASS incl. PerWriteGuard 4.63s + e2e pair (12.785s) |
| `go test -count=1 -p 1 ./internal/envprofile/ -run 'Test(DraftPath\|LegacyPath\|AdmitPath)'` | 0 | 6/6 PASS (7.026s) |
| `go test -count=1 ./internal/install/ -run TestRemovedSkillCleanedUp` | 0 | ok 9.904s (Windows-failing control, local lane) |
| `go vet` on snapshot/staging/privatedir/adapters/envprofile/install/transaction | 0 | clean |
| `gofmt -l` on same 7 packages | 0 | empty (after -w on 2 files) |
| `git diff --check` | 0 | clean |
| `GOOS=windows go build ./internal/staging/ ./internal/transaction/ ./internal/install/` | 0 | cross-compiles (new Windows FileIdentity) |
| `GOOS=windows go vet ./internal/staging/` | 0 | clean |

Full landing suite not run manually; handoff owns the single
remote-gate execution. No installs, daemon restarts, tags, LOGBOOK
edits (prohibited; findings here), runtime-home changes, live
credential export, or ax calls.

## Measured negative evidence (this run)

`internal/staging/boundaries.go` + `internal/transaction/engine.go`
snapshotted with `shasum -a 256` before mutation; each mutant
reverted and shasums verified OK after. Post-restore suites green
(exit 0, above). Both single-site narrowings with passing siblings.

| # | Narrowing mutation | Killer (exit) | Sibling |
|---|---|---|---|
| M1 (F1) | `checkBoundary` durable path replaced with `return nil` (skip proof, legacy bypass) | `TestRecoveryMustRecheckPhysicalBoundary` exit 1 (`recovery published across replaced parent: live="new"`) | unchanged + legacy recovery tests exit 0 |
| M2 (F2) | staging bytes `Canonicalize` error back to `source_output_overlap` | `TestPlanRecheckIdentityReadErrorIsNotOverlap` exit 1 (`mislabeled as overlap: ... EvalSymlinks: too many links`) | proven-overlap rows exit 0 |

2/2 killed locally. Reviewer probes rev6 (`TestReviewRecovery…`,
`TestReviewIdentity…`) are subsumed by the committed twins above
(same shapes, same assertions plus diagnostic-class pins).

## Coverage map (rev7 delta → entry points)

- Restart recheck: `Engine.Recover→resume→commit→commitTarget→
  checkBoundary→staging.RecheckDurableOne` — F1 trio through real
  journals; `validateBoundaryProof` pins journal shape; install
  `attachBoundaries→journalPlan→Prepare` carries proof (Durable
  assertions).
- Unreadable: staging `Snapshot/Recheck/RecheckOne/RecheckDurableOne`
  + `FileIdentity`, adapters `ValidateDestinations`, snapshot
  `ValidateLocalPackage/ValidateRootInputs/EnumerateInputs/
  checkRootCoverage/checkDeclaredInputsDisjoint/sameFile` —
  4 committed unreadable tests + durable parity; install/
  transaction propagate verbatim.
- Windows: `identity_windows.go` (backup semantics, follow links,
  IsNotExist-mapped vanished) cross-compiles; `Within` Lstat fix
  from rev6 untouched; `TestRemovedSkillCleanedUp` green locally,
  hosted Windows is the enforcing runner.

## Bounds (stated, not deferred in-scope work)

- All rev6 bounds carry over (admitted-set definition,
  `ensureDefault`/git-clone funnels, NFD blind spot, inode-reuse
  parity — durable Dev:Ino/FileIndex shares the in-memory
  SameFile reuse window — Windows eager pin, mid-path-link+case
  analyzed flip).
- Crash-recovery guard absence bound from rev5/6 is CLOSED by this
  rev (durable proof + F1 trio); empty-proof (`Check` without
  `Proof`, test scaffolding only) fails closed on recovery by
  construction.
- 16k7xy seam (binding ruling: not a finding): the draft
  package-snapshot store and its acquisition entry point belong to
  TASK-260910-16k7xy (capture-and-store-local-package-snapshots).
  Exact call site: `PrepareLocalAcquisition` /
  `LocalAcquisition` in `internal/snapshot/boundaries.go` (this
  leaf validates, enumerates, and reserves staging; 16k7xy owns
  byte capture, hashing, race detection, and store publication).
- Journal forward-compat: new journals with `boundary_proof` are
  unreadable by pre-rev7 binaries (`DisallowUnknownFields`);
  backward-compat holds (legacy journals validate, skip as
  legacy). Single-binary upgrade, no rolling.
- Journal tampering (manager-home write) can weaken proof
  subset checks; manager home is trusted, same as live targets.
- Checklist item 12 findings live here (LOGBOOK edits
  prohibited).
