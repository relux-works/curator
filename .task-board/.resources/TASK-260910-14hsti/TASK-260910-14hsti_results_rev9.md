# TASK-260910-14hsti — rework-6 handoff (rev9)

Rework-6 producer run after verdict rev8 CHANGES_REQUESTED (F1 HIGH:
Windows captures inconsistent in-memory and durable identities —
`identityToken` discards the pinned FileInfo and re-opens the path, so
a swap between the two reads plants Info(A)+Token(B)). Prior evidence:
results.md (rev1), _rev2…_rev8, rev1–rev8 CR patches and validation
logs, verdicts rev2/rev4/rev6/rev7/rev8, probes rev6/rev7/rev8. This
resource appends the rev9 delta only; rev8 stands except where fixed
below.

## Candidate and contract

- Story worktree base HEAD: `12f1287ee0fb538f9ca004dd53b870e824e5baf2`
  (unchanged; no commits by this producer — the rev8 tree plus the
  delta below, uncommitted, for the handoff snapshot).
- Delta vs rev8 (same 28 worktree paths, no new files):
  `internal/staging/boundaries.go` (single-inspection
  `captureIdentity` + `SetCaptureHook` seam, token-authoritative
  `PinnedIdentity`, `verifyDurableIdentity` for both same-process and
  recovery checks; `pinIdentity`/`recheckIdentity` removed),
  `internal/staging/identity_unix.go` + `identity_windows.go`
  (`inspectPath`/`tokenOfInspection`/`FileIdentity` split;
  `identityToken` removed), `internal/staging/boundaries_test.go`
  (2 new capture-window tests), `internal/transaction/boundaries_test.go`
  (2 new real-journal capture-window regressions). No other file
  touched — install/envprofile/transaction production code unchanged.
- Spec: curator-spec main `871d11b`,
  protocol/skillfile-sources.md (revision 1 semantics; §2 physical
  boundaries, §5 diagnostics) and protocol/repository-transport.md
  rev1. Proven changes keep exact spec §5 `source_output_overlap`;
  inspection failures keep the rework-authorized
  `boundary_identity_unreadable`, never overlap. Frozen v1 wire schemas
  untouched (journal `boundary_proof` shape unchanged; same values).
- Shell for all commands below: tool-managed `/bin/sh`; every exit
  below is the direct process exit (output redirected to a file, no
  pipe masking).

## F1 — one inspection per pin; the durable token is the single authoritative pin

`captureIdentity` (boundaries.go) pins one canonical path with a
single filesystem inspection whose result feeds the token derivation
and nothing else:

- unix `inspectPath` is one `os.Stat`; `tokenOfInspection` formats
  Dev:Ino from that buffer with zero I/O (unchanged semantics).
- Windows `inspectPath` opens the object ONCE (`CreateFile` with
  `FILE_FLAG_BACKUP_SEMANTICS`, links followed to match Stat) and
  reads volume serial + file index from THAT handle;
  `tokenOfInspection` formats the triple with zero I/O. There is no
  second pathname-based open anywhere in capture: the old
  `identityToken(_, canonical) → FileIdentity(canonical)` second read
  is deleted, with it the Info(A)+Token(B) split the verdict
  demonstrates (pin/swap/capture/restore-A/Prepare/swap-B/Recover).
- `PinnedIdentity` is now `{Token}` only: the durable token is the
  single authoritative pin used by both the same-process guard and
  the journal, per the rework-6 ruling. Same-process `Recheck` and
  restart `RecheckDurableOne` both verify through the one
  `verifyDurableIdentity` comparison (live `FileIdentity` vs the
  planning-time token); `pinIdentity` (SameFile self-compare eager
  load) and `recheckIdentity` (Stat+SameFile) are removed. A
  hand-built pin with no token fails closed as overlap-or-unreadable
  through the existing gates, never panics (old `SameFile(nil, …)`
  would).
- Error classes preserved: missing admitted input is still
  `source_member_missing`; absent ancestors still unpinned;
  non-absence inspection failures still
  `boundary_identity_unreadable`; proven change/vanish still
  `source_output_overlap` with the same "changed since planning" /
  "no longer exists" texts.
- `Durable` is untouched and still free of filesystem reads.

## Test seam

`SetCaptureHook` (boundaries.go) installs the test-only capture hook
that `captureIdentity` fires after its single inspection and before
the pure token derivation — the exact point where a second
pathname-based read would observe a swapped object. Production never
sets it; tests clear it before returning (plus `t.Cleanup`). A swap
injected here proves single-inspection capture on every platform: on
fixed code the pin still names the pre-swap object; on any code that
re-reads the path after the hook the pin names the replacement and
the tests below fail.

## Tests (committed, production entry points)

- Staging pin pair (hook swap inside `Snapshot`, then token-equality
  + `RecheckOne` refusal + restore control, rename+mkdir only so
  every lane runs them):
  `TestSnapshotCapturePinsAncestorBeforeHookSwap` (destination
  ancestors: pin == pre-swap token, recheck refuses with
  overlap+identity gate, restore passes),
  `TestSnapshotCapturePinsAdmittedBeforeHookSwap` (admitted inputs:
  same shape, admitted-identity gate).
- Transaction real-journal restart pair (committed twins of the rev8
  reviewer probe `TestReviewWindowsCaptureCoherence`, same sequence
  through real Engine.Prepare + fresh Engine.Recover: hook swap
  inside Snapshot, control refusal, Durable, restore-A, pre-journal
  control, Prepare, swap-B, same-process control, Recover refuses
  with `source_output_overlap`, live stays old, journal gone):
  `TestRecoveryCaptureWindowAncestorRefuses`,
  `TestRecoveryCaptureWindowAdmittedRefuses` (additionally asserts
  the admitted-identity gate text in the recovery error).
- Controls where nothing changed (unchanged + legacy): existing
  `TestRecoveryWithUnchangedBoundaryPublishes`,
  `TestRecoveryLegacyJournalSkipsGuard` — both pass unchanged and
  both pass under the mutant (see below), proving the new tests are
  narrowing, not broad.

## Validation (this producer, final candidate)

| Command | Exit | Evidence |
|---|---|---|
| `go test -count=1 ./internal/staging/` | 0 | ok 0.426s, incl. 2 new capture tests |
| `go test -count=1 -p 1 ./internal/transaction/ -run 'TestCommit.*Boundary\|TestCommitRollbackNever\|TestRecovery'` | 0 | ok 8.629s, 12/12 incl. 2 new restart regressions |
| `go test -count=1 -p 1 ./internal/snapshot/ ./internal/privatedir/ ./internal/adapters/` | 0 | ok 2.383s/0.376s/0.423s |
| `go test -count=1 -p 1 ./internal/install/ -run 'Test(StageProjectTargets\|StageGlobalTargets\|RunCommit.*(Boundar\|Recheck\|RealStaging)\|PerWriteGuard\|ProjectInstallRefusesAdapter\|GlobalInstallRefusesAdapter)'` | 0 | ok 13.802s |
| `go test -count=1 -p 1 ./internal/envprofile/ -run 'Test(DraftPath\|LegacyPath\|AdmitPath)'` | 0 | ok 380.446s (host slow-start; see note) |
| `go test -count=1 ./internal/install/ -run TestRemovedSkillCleanedUp` | 0 | ok 11.398s (Windows-failing control, local lane) |
| `go vet` on snapshot/staging/privatedir/adapters/envprofile/install/transaction | 0 | clean, empty output |
| `gofmt -l` on same 7 packages | 0 | empty |
| `git diff --check` | 0 | clean |
| `GOOS=windows go build ./internal/staging/ ./internal/transaction/ ./internal/install/` | 0 | cross-compiles (new single-handle Windows capture) |
| `GOOS=windows go vet ./internal/staging/` | 0 | clean |

Full landing suite not run manually; handoff owns the single
remote-gate execution. No installs, daemon restarts, tags, LOGBOOK
edits (prohibited; findings here), runtime-home changes, live
credential export, or ax calls.

Host note: fresh test binaries on this host intermittently take
minutes to start at 0% CPU (the rev8 verdict addendum documents the
same); one parallel evidence re-run hit the 10m `go test` timeout in
all three packages without executing a single test, on a tree whose
shasum-verified identical bytes pass the same suites in seconds (see
rows above). No suite in the table above ever failed on test
assertions; timings are the passing runs.

## Measured negative evidence (this run)

`internal/staging/identity_unix.go` + `boundaries.go` snapshotted
with `shasum -a 256` before mutation; restored bytes verified OK
after (`shasum -c`, both OK, zero MUTANT markers remain).
Post-restore suites green (staging exit 0, transaction exit 0).

| # | Narrowing mutation | Killer (exit) | Siblings |
|---|---|---|---|
| M1 (capture window) | `tokenOfInspection` re-stats the path (`os.Stat(canonical)`, discarding the inspection) — the exact reintroduced second open | staging capture pair FAIL (both `TestSnapshotCapturePins*` FAIL, package FAIL); transaction `TestRecovery` mask direct exit 1 with exactly `TestRecoveryCaptureWindowAncestorRefuses` + `TestRecoveryCaptureWindowAdmittedRefuses` FAIL (`control: physical guard did not detect replacement` — the pin captured the post-hook identity, the modeled defect) | 6/6 transaction siblings PASS incl. unchanged + legacy controls; `TestDurableKeepsPlannedIdentityAfterParentReplacement` PASS |

1/1 killed locally, both halves (ancestors + admitted) at both
levels (staging pin + real-journal restart). The rev8 reviewer probe
(`TestReviewWindowsCaptureCoherence`) is subsumed by the committed
twins above (same swap/restore/Prepare/Recover schedule, production
seam instead of the overlay hook, committed names without the Review
prefix per the twin convention).

## Coverage map (rev9 delta → entry points)

- Planning capture: `Plan.Snapshot→captureIdentity→inspectPath/
  tokenOfInspection` (+`snapshotAncestors`) — every admitted input
  and destination ancestor carries a single-inspection token;
  uncapturable fails closed before journaling.
- Same-process guard: `Recheck`/`RecheckOne→verifyDurableIdentity`
  — reached via install `attachBoundaries→journalPlan→Prepare`
  pre-journal check (`commit.go:658`) and the per-write engine
  `BoundaryCheck`, both exercised by the existing install/transaction
  suites above with zero behavior change on green paths.
- Restart: `Snapshot.Durable` (unchanged, still zero I/O) →
  `Engine.Recover→checkBoundary→staging.RecheckDurableOne→
  verifyDurableIdentity→FileIdentity` — the new twins pin the full
  restart path through the real journal with the capture-window
  schedule.
- Windows production path (`inspectPath` single `CreateFile`):
  cross-compiles and vets clean; native Windows execution is covered
  by the remote gate (this host is darwin). The committed tests model
  the Windows two-read algorithm deterministically through the seam
  on every lane.

## Bounds (stated, not deferred in-scope work)

- All rev8 bounds carry over (admitted-set definition,
  `ensureDefault`/git-clone funnels, NFD blind spot, inode-reuse
  parity — durable Dev:Ino/FileIndex shares the in-memory reuse
  window — mid-path-link+case analyzed flip; journal forward-compat
  single-binary upgrade; manager-home trust). The rev8 "swap between
  the two adjacent planning syscalls" bound is CLOSED by this
  revision: there is no longer a second planning read to land
  between, on either platform.
- `Within` still compares live-vs-live with `os.Lstat`+`os.SameFile`
  (planning-time alias detection, no stored pin, hence no capture
  window of this class); inherent TOCTOU between its two live reads
  is unchanged and out of scope.
- The capture hook is test-only global state, not safe for
  concurrent use; no test in the affected packages runs parallel,
  and production never sets it.
- 16k7xy seam (binding ruling: not a finding): the draft
  package-snapshot store and its acquisition entry point belong to
  TASK-260910-16k7xy (capture-and-store-local-package-snapshots).
  Exact call site: `PrepareLocalAcquisition` /
  `LocalAcquisition` in `internal/snapshot/boundaries.go` (this
  leaf validates, enumerates, and reserves staging; 16k7xy owns
  byte capture, hashing, race detection, and store publication).
- Checklist item 12 findings live here (LOGBOOK edits
  prohibited).
