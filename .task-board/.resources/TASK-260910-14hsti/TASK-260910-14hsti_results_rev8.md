# TASK-260910-14hsti — rework-5 handoff (rev8)

Rework-5 producer run after verdict rev7 CHANGES_REQUESTED (F1 HIGH
Durable conversion replaces the original boundary identity). Prior
evidence: results.md (rev1), _rev2…_rev7, rev1–rev7 CR patches and
validation logs, verdicts rev2/rev4/rev6/rev7, probes rev6/rev7. This
resource appends the rev8 delta only; rev7 stands except where fixed
below.

## Candidate and contract

- Story worktree base HEAD: `12f1287ee0fb538f9ca004dd53b870e824e5baf2`
  (unchanged; no commits by this producer — the rev7 tree plus the
  delta below, uncommitted, for the handoff snapshot).
- Delta vs rev7 (same 28 worktree paths, no new files):
  `internal/staging/boundaries.go` (PinnedIdentity pins, Snapshot token
  capture, serialize-only Durable),
  `internal/staging/identity_unix.go` + `identity_windows.go`
  (identityToken), `internal/staging/boundaries_test.go` (3 new tests),
  `internal/transaction/boundaries_test.go` (1 new test). No other file
  touched.
- Spec: curator-spec main `871d11b`,
  protocol/skillfile-sources.md (revision 1 semantics; §2 physical
  boundaries, §5 diagnostics) and protocol/repository-transport.md
  rev1. Proven changes keep exact spec §5 `source_output_overlap`;
  uncapturable/unrecorded identity uses the rework-authorized
  `boundary_identity_unreadable`, never overlap. Frozen v1 wire schemas
  untouched (journal `boundary_proof` shape unchanged; only the values
  it carries are now planning-time).
- Shell for all commands below: tool-managed `/bin/sh`; mutant killer
  exits are the `go test` process exit (PIPESTATUS / direct redirect,
  no pipe masking).

## F1 — Durable serializes the planned identity, never re-stats

`Snapshot` now pins `PinnedIdentity{Info, Token}` per admitted input
and per destination ancestor: the FileInfo the same-process Recheck
compares with `os.SameFile`, plus the durable token captured together
at planning (`identityToken`: unix derives Dev:Ino from the already
stattd `Sys()` with zero I/O; Windows reads the volume/file-index
token adjacent to the pin, since FileInfo carries no public index).
Token capture failure fails closed at planning as
`boundary_identity_unreadable` — no snapshot, no journal, no
publication. `Snapshot.Durable()` copies the stored tokens with zero
filesystem calls; a pin without a token (hand-built only — Snapshot
never produces one) fails closed as `boundary_identity_unreadable`.
Recovery (`verifyDurableIdentity` via `FileIdentity`) still inspects
the live filesystem — that is the comparison side, unchanged.

## Tests (committed, production entry points)

- Staging conversion pair (swap between Snapshot and Durable, then
  `RecheckDurableOne` + restore control):
  `TestDurableKeepsPlannedIdentityAfterParentReplacement` (destination
  ancestors: refuses with overlap+identity gate, restore passes),
  `TestDurableKeepsPlannedAdmittedIdentityAfterSwap` (admitted inputs:
  refuses with overlap+admitted gate, restore passes).
- `TestDurableRefusesPinWithoutCapturedToken` (hand-built empty token
  fails closed as unreadable, never overlap; ancestor + admitted).
- Transaction real-journal restart:
  `TestRecoveryUsesOriginalSnapshotIdentity` — committed twin of the
  rev7 reviewer regression `TestReviewRecoveryUsesOriginalSnapshotIdentity`
  (same sequence through real Engine.Prepare + fresh Engine.Recover:
  Snapshot, swap, Durable, restore, pre-journal control, Prepare,
  swap back, same-process guard control, Recover refuses with
  `source_output_overlap`, live stays old, journal gone).
- Controls where nothing changed (unchanged + legacy): existing
  `TestRecoveryWithUnchangedBoundaryPublishes`,
  `TestRecoveryLegacyJournalSkipsGuard`,
  `TestRecheckDurableOneMatchesRecheck` — all pass unchanged.

## Validation (this producer, final candidate)

| Command | Exit | Evidence |
|---|---|---|
| `go test -count=1 ./internal/staging/ ./internal/snapshot/ ./internal/privatedir/ ./internal/adapters/` | 0 | ok 0.718s/1.764s/0.959s/1.654s, incl. 3 new staging tests |
| `go test -count=1 -p 1 ./internal/transaction/ -run 'TestCommit.*Boundary\|TestCommitRollbackNever\|TestRecovery'` | 0 | ok 7.703s, incl. new restart regression (quartet alone 4/4 PASS, 3.159s) |
| `go test -count=1 -p 1 ./internal/install/ -run 'Test(StageProjectTargets\|StageGlobalTargets\|RunCommit.*(Boundar\|Recheck\|RealStaging)\|PerWriteGuard\|ProjectInstallRefusesAdapter\|GlobalInstallRefusesAdapter)'` | 0 | ok 14.620s |
| `go test -count=1 -p 1 ./internal/envprofile/ -run 'Test(DraftPath\|LegacyPath\|AdmitPath)'` | 0 | ok 7.246s, 6/6 PASS |
| `go test -count=1 ./internal/install/ -run TestRemovedSkillCleanedUp` | 0 | ok 11.170s (Windows-failing control, local lane) |
| `go vet` on snapshot/staging/privatedir/adapters/envprofile/install/transaction | 0 | clean |
| `gofmt -l` on same 7 packages | 0 | empty |
| `git diff --check` | 0 | clean |
| `GOOS=windows go build ./internal/staging/ ./internal/transaction/ ./internal/install/` | 0 | cross-compiles (new Windows identityToken) |
| `GOOS=windows go vet ./internal/staging/` | 0 | clean |

Full landing suite not run manually; handoff owns the single
remote-gate execution. No installs, daemon restarts, tags, LOGBOOK
edits (prohibited; findings here), runtime-home changes, live
credential export, or ax calls.

## Measured negative evidence (this run)

`internal/staging/boundaries.go` snapshotted with `shasum -a 256`
(`e9b69a22…ba75`) before mutation; each mutant reverted and shasum
verified OK after. Post-restore suites green (staging exit 0,
transaction TestRecovery exit 0). Both single-site narrowings with
passing siblings.

| # | Narrowing mutation | Killer (exit) | Sibling |
|---|---|---|---|
| M1 (ancestors) | Durable ancestors loop re-stats via `FileIdentity(ancestor)` instead of `pin.Token` | staging `TestDurableKeepsPlannedIdentityAfterParentReplacement` exit 1 (`converted-after-swap durable err = <nil>`); transaction `TestRecoveryUsesOriginalSnapshotIdentity` exit 1 (`recovery published across replaced parent: live="new"` — the exact reviewer symptom) | admitted durable test exit 0 |
| M2 (admitted) | Durable admitted loop re-stats via `FileIdentity(canonical)` instead of `pin.Token` | staging `TestDurableKeepsPlannedAdmittedIdentityAfterSwap` exit 1 (`converted-after-swap admitted err = <nil>`) | parent-replacement durable test exit 0 |

2/2 killed locally. The rev7 reviewer probe
(`TestReviewRecoveryUsesOriginalSnapshotIdentity`) is subsumed by the
committed twin above (same sequence, same assertions, committed name
without the Review prefix per the rev7 twin convention).

## Coverage map (rev8 delta → entry points)

- Planning capture: `Plan.Snapshot→identityToken` (+`snapshotAncestors`)
  — every admitted input and destination ancestor carries its durable
  token from planning; uncapturable fails closed before journaling.
- Conversion: `Snapshot.Durable` (pure serialize, zero I/O) —
  `install.attachBoundaries→journalPlan→Prepare` carries the proof;
  crash-recovery re-verifies via `Engine.Recover→…→checkBoundary→
  staging.RecheckDurableOne→verifyDurableIdentity→FileIdentity`.
- Regression paths: staging pair pins both halves (ancestors +
  admitted) at `RecheckDurableOne`; transaction twin pins the full
  restart path through the real journal.

## Bounds (stated, not deferred in-scope work)

- All rev7 bounds carry over (admitted-set definition,
  `ensureDefault`/git-clone funnels, NFD blind spot, inode-reuse
  parity — durable Dev:Ino/FileIndex shares the in-memory SameFile
  reuse window — mid-path-link+case analyzed flip; journal
  forward-compat single-binary upgrade; manager-home trust).
- Inside-Snapshot capture order (unix derives the token from the same
  `stat` buffer with zero window; Windows reads the token adjacent to
  the pin because the index is not public): a swap landing exactly
  between the two adjacent planning syscalls still fails closed at the
  pre-journal in-memory Recheck (`commit.go:658`), so no journal is
  persisted from an inconsistent pin.
- 16k7xy seam (binding ruling: not a finding): the draft
  package-snapshot store and its acquisition entry point belong to
  TASK-260910-16k7xy (capture-and-store-local-package-snapshots).
  Exact call site: `PrepareLocalAcquisition` /
  `LocalAcquisition` in `internal/snapshot/boundaries.go` (this
  leaf validates, enumerates, and reserves staging; 16k7xy owns
  byte capture, hashing, race detection, and store publication).
- Checklist item 12 findings live here (LOGBOOK edits
  prohibited).
