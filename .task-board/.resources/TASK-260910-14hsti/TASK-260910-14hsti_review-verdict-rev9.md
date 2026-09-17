# TASK-260910-14hsti — revision 9 independent review

Verdict: ACCEPTED. No remaining blocking finding in the reviewed scope.

## Exact candidate and scope

CR-TASK-260910-14hsti-9; base 12f1287ee0fb538f9ca004dd53b870e824e5baf2;
candidate tree d211e9bf59555176ba44c19cbd551c7d49bcf4c8. All 28 changed
worktree files were compared byte-for-byte with the candidate blobs before
checks and SHA-256 checked again after attacks: 28/28 unchanged. No code edits,
commits, runtime-home changes, or full local suite. Attacks used Go overlays in
/tmp only. The accompanying review-probes archive carries overlays, logs, and
the candidate byte manifest.

Read skillfile-sources revision 1 physical-boundary contract and repository-
transport revision 1, producer results rev9, and rev9 hosted validation log.
Scope includes install/envprofile/transaction wiring explicitly authorized by
rework-3 and subsequent rulings. No frozen protocol schemas, parser/collections
reimplementation, transport-v2 behavior, or unrelated feature additions.

## Finding closure

Rev8 F1 is closed: Windows inspectPath opens once with directory-capable backup
semantics, obtains volume/file index from that handle, and returns the captured
triple. tokenOfInspection does no I/O. Both same-process and durable recovery
checks use that same authoritative token; Durable only serializes captured
values. Unix similarly derives from one Stat buffer. Following the canonical
ancestor/input for this capture is intentional; mirror entry checks retain their
separate Lstat behavior. Inspection errors propagate boundary_identity_unreadable;
proven identity changes propagate source_output_overlap.

Capture-window coverage drives Snapshot and real Engine.Prepare/fresh Recover,
for both destination ancestors and admitted inputs, with restore-A-before-Prepare
and B-before-Recover. Unchanged and legacy controls pass. Earlier recovery proof
persistence and checks immediately before backup/install renames remain present.
Protected missing proof is not silently reclassified as a legacy journal.

## Acceptance traceability

Six task acceptance groups inspected and exercised (6/6 at the named entry
points; this is not a claim to cover every semantic vector):

1. Physical paths, symlinks, case: snapshot validation tests
   TestValidateSymlinkManagedAndEscape / TestValidateCaseAliasUsesFilesystemIdentity;
   Plan.Snapshot/Recheck tests and real-journal recovery identity tests. The
   case-sensitive platform skip remains declared; native platform results below.
2. Authored versus output trees: PrepareLocalAcquisition via
   TestPrepareLocalAcquisitionAdmitsSafeSubdirectory / Refusals /
   ValidatesBeforeStaging; authored agents accepted, managed outputs refused.
3. Operator root_inputs: TestPrepareLocalAcquisitionAdmitsRootInputs / Refusals,
   with detailed ValidateRootInputs negative tests for coverage, overlap, links,
   missing and invalid entries. Operator policy loading is the separate policy leaf.
4. Before traversal and at publication: draft envprofile.Install path and overlay
   tests drive stateForPath -> PrepareLocalAcquisition; real install stagers fill
   Group.Admitted and scopeTargets.boundaries. TestPerWriteGuardRefusesSwappedParentAndRollsBack
   uses the real journal, and transaction recovery regressions cover restart.
5. Unmanaged protection: TestUnmanagedTakeoverRefusedBesideAdmittedCheck drives
   adapters.StageProject and verifies foreign bytes remain. Production project/global
   install refusal tests drive real staging and destination-over-snapshot refusal.
6. Broad path '.' with safe selection: TestPrepareLocalAcquisitionAdmitsSafeSubdirectory
   explicitly selects agents/skills/review from the project alias root. Draft profile
   positive and legacy-off controls remain passing.

## Independent execution (zsh; direct process results)

- `go test -count=1 -p 1 ./internal/staging ./internal/transaction -run 'Test(SnapshotCapture|Durable|Recovery|Commit.*Boundary|Recheck|PlanRecheck)'`: exit 0; staging 0.368s, transaction 14.761s.
- `go test -count=1 -p 1 ./internal/snapshot ./internal/privatedir ./internal/adapters`: exit 0; 1.445s / 0.846s / 1.483s.
- `go test -count=1 -p 1 ./internal/install ./internal/envprofile -run 'Test(Stage(Project|Global)Targets|RunCommit.*(Boundar|Recheck|RealStaging)|PerWriteGuard|ProjectInstallRefusesAdapter|GlobalInstallRefusesAdapter|DraftPath|LegacyPathInstall|AdmitPathSource)'`: exit 0; install 44.287s, envprofile 30.193s.
- `go vet -p 1` on staging, transaction, snapshot, privatedir, adapters, install, envprofile followed by `gofmt -l` on those directories: combined command exit 0, no diagnostics or formatting output.
- `git diff --check`: exit 0.

## Independent narrowing attacks

2/2 killed, 0 survivors; original repository bytes never modified.

M1: reintroduce a second pathname inspection after the capture hook, retaining all
other gates. `go test -overlay /tmp/14hsti-review9/capture.json -count=1 -p 1
./internal/transaction -run 'TestRecoveryCaptureWindow(Ancestor|Admitted)Refuses'`
exits 1. BOTH tests fail at the explicit control: physical guard did not detect
replacement. This proves capture coherence, not merely the existence of a guard.

M2: narrow durable recovery enforcement to proofs with nonempty Admitted, allowing
ancestor-only proofs to skip. `go test -overlay /tmp/14hsti-review9/recovery.json
-count=1 -p 1 ./internal/transaction -run
'TestRecovery(MustRecheckPhysicalBoundary|CaptureWindowAncestorRefuses|WithUnchangedBoundaryPublishes|LegacyJournalSkipsGuard)'`
exits 1. Both changed-parent tests fail because recovery publishes new bytes;
unchanged and legacy controls do not fail. This tests the real restart publisher.

## Hosted evidence accepted, not rerun

TASK-260910-14hsti_change-request_rev9-validation.log records exit 0 for
https://github.com/relux-works/curator/actions/runs/35138057448 . Gate commit
6be43c75afeb53f16d0a47447780f25c2e9a656f has exactly candidate tree
d211e9bf59555176ba44c19cbd551c7d49bcf4c8 (verified locally with git show).
Linux/macOS/Windows test lanes, Linux/macOS race, lint, naming, interop and all
three platform gate self-tests succeeded. Rose-air and candidate-suite lanes
were skipped: no native rose-air claim. This is log-level hosted evidence,
not an independent Windows execution or per-test artifact audit.

## Bounds

The draft package-snapshot store/capture belongs to TASK-260910-16k7xy, whose
handoff seam is PrepareLocalAcquisition / LocalAcquisition in
internal/snapshot/boundaries.go. This leaf validates/enumerates/reserves staging;
16k7xy owns frozen byte capture, race detection, hashing and store publication.
The profile store retains its own capture traversal. Additional effective input
requirements must be supplied by the owning pipeline; this review does not claim
full end-to-end draft Skillfile installation. Existing missing-path Unicode
normalization and filesystem identifier reuse limitations remain stated bounds.
No full local module suite, native Windows replay, or release qualification was
performed. LOGBOOK.md edits are prohibited by the campaign; this task-scoped
verdict records the review evidence instead.

Run goal queried: not goal-bound. All checklist entries were already checked.
Route via accept_cr revision=9 only; acceptance is not integration or done.
