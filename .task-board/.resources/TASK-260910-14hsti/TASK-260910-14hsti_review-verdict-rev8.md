# TASK-260910-14hsti revision 8 — CHANGES_REQUESTED

Candidate 9b0cbb14133454bfa7ff27289eb0f2ae5150db1d; base 12f1287ee0fb538f9ca004dd53b870e824e5baf2. Reviewer RUN-260916-d93ed4. Goal queried: not goal-bound. All 28 candidate files compared byte-for-byte against the exact tree before and after overlay preparation; 28/28 match. No repository source or test files edited.

## F1 HIGH — Windows still captures inconsistent in-memory and durable identities

Locations: internal/staging/identity_windows.go:19–20; internal/staging/boundaries.go:276–283 and :331–340; production consumer internal/install/commit.go:658 and attachBoundaries, transaction Engine.Recover/checkBoundary.

Revision 8 fixes conversion-time reinspection: Durable now copies stored tokens. Unix derives the token from the SAME FileInfo and closes the reported rev7 hole. Windows identityToken explicitly discards that FileInfo and calls FileIdentity(canonical), which opens the pathname again. Thus PinnedIdentity can still hold Info(A) and Token(B). No comparison establishes that the second handle names the object pinned by the first.

Counterexample is the prior accepted regression moved inside Snapshot: pin parent A; swap A for B between pinIdentity and identityToken; capture token B; restore A before the pre-journal Recheck; Prepare a protected journal; restore B before fresh Engine.Recover. The pre-journal check compares Info(A) to restored A and passes. Recovery compares the durable token B to B and publishes. The results_rev8 assertion that the pre-journal Recheck necessarily prevents an inconsistent pin from reaching the journal is false for this swap-and-restore sequence. Both admitted inputs and ancestor capture use this pattern.

This is a Windows-specific code finding, not a native Windows execution claim. The supplied deterministic overlay models Windows's second-path-read algorithm on Unix and injects the swap at the syscall boundary; its binary did not start on this host, so this is a code-derived counterexample, not an observed runtime reproduction; native Windows execution of that schedule remains unverified. Go's local os/types_windows.go confirms SameFile loads/caches the first identity in loadFileId; it does not establish equality with the later independent CreateFile in identityToken.

Required: capture one authoritative planning identity and use it consistently for same-process and durable verification. On Windows derive both proof forms from the same opened object, or use the captured durable identity as the authoritative pin for both paths. Do not rely on temporal adjacency or a later pathname-only/pre-journal check. Add deterministic capture-window coverage for destination ancestors and admitted inputs, including restore-A-before-Prepare / B-before-Recover and unchanged/legacy controls. Durable must remain free of filesystem reads. Ordinary implementation rework, no human/platform decision required.

## Independent validation

Shell zsh, direct process exits. No whole-module suite or manual hosted gate rerun.

- `go test -count=1 -p 1 ./internal/staging ./internal/transaction -run 'Boundary|Boundar|Recovery|Durable|Snapshot|Identity|Recheck'`: exit 0, staging 0.760s; transaction 54.779s. New real-journal restart regression passes on this macOS host.
- `go test -count=1 -p 1 ./internal/snapshot ./internal/privatedir ./internal/adapters`: exit 0; 3.122s / 0.555s / 3.476s.
- Focused install/envprofile command: install reported PASS 14.024s; remaining result recorded in execution addendum below.
- `go vet ./internal/staging ./internal/snapshot ./internal/privatedir ./internal/adapters ./internal/transaction ./internal/install ./internal/envprofile`: independently observed exit 0.
- `gofmt -l` on all seven packages: empty output. `git diff --check`: exit 0.

M1 (narrowing): retain planned admitted tokens but re-stat destination ancestors in Durable. `go test -count=1 -p 1 -overlay .temp/review-14hsti-rev8/mutant-durable.json ./internal/transaction -run 'TestRecovery(UsesOriginalSnapshotIdentity|WithUnchangedBoundaryPublishes|LegacyJournalSkipsGuard)' -v` exits 1. TestRecoveryUsesOriginalSnapshotIdentity fails with `recovery published across replaced parent: live="new"`; unchanged and legacy controls PASS. This independently confirms the new regression kills the original rev7 defect. Additional attacks/results appear below. All mutations are Go overlays; producer bytes remain unchanged.

Read producer results_rev8 and change-request_rev8-validation.log. Attached hosted run 35125599003, gate commit 00b12a7ad0c20d68c2ceb4be2b64c49f6c775fb3, reports exit 0: Ubuntu/macOS tests and race, Windows tests, lint, naming, interop and gate self-tests successful. This is accepted as attached platform evidence, not rerun locally. Rose-air and Candidate suite skipped; no ARM64 claim. Existing Windows mirror/directory tests stay green in that evidence, but do not inject the newly identified capture interval.

## Acceptance trace and bounds

Reviewed skillfile-sources.md §2 and the binding rework rulings; only five files differ from rev7, all in staging identity capture/tests and the transaction regression. Full delta stays within boundary packages and explicitly authorized install/envprofile/transaction callers. Frozen schemas unchanged; no transport-v2 implementation. Legacy path acquisition remains behind the existing DraftSourcesV1 switch.

Test anchors in candidate:
- Physical paths, case, managed versus authored: TestValidateCaseAliasUsesFilesystemIdentity, TestValidateSymlinkManagedAndEscape, TestValidateManagedMissingRefusedAsOverlapBeforeTraversal; production path controls TestDraftPathInstallRefusesStoreOverlap and TestDraftPathOverlayRefusesStoreOverlap.
- root_inputs and broad alias with safe selection: TestPrepareLocalAcquisitionAdmitsRootInputs, TestPrepareLocalAcquisitionRefusals, TestValidateRootInputsRefusals, TestPrepareLocalAcquisitionAdmitsSafeSubdirectory.
- Before traversal: TestPrepareLocalAcquisitionValidatesBeforeStaging, PrepareLocalAcquisition called from envprofile stateForPath/admitPathSource.
- Adapter/public install separation and unmanaged protection: TestProjectInstallRefusesAdapterDestinationOverSnapshot, TestGlobalInstallRefusesAdapterDestinationOverSnapshot, TestUnmanagedTakeoverRefusedBesideAdmittedCheck, TestStageProjectRefusesMirrorOverAdmittedInput.
- Publication: TestPerWriteGuardRefusesSwappedParentAndRollsBack via real stageProjectTargets/runCommit/journal; restart quartet including TestRecoveryUsesOriginalSnapshotIdentity. Windows capture coherence remains incomplete per F1.

This is not a claim of all 73 semantic cases or exhaustive acceptance coverage. Package byte capture/hash/store publication belong to TASK-260910-16k7xy: exact seam PrepareLocalAcquisition/LocalAcquisition in internal/snapshot/boundaries.go. That bound is not a finding. Existing missing-path Unicode-normalization and inode-reuse bounds remain. No live runtime-home changes, installs, commits, or LOGBOOK.md edits; findings recorded here and in board notes per campaign rules.

Verdict: CHANGES_REQUESTED, route to to-dev. No accept_cr, commit_ack, or human approval requested.

## Execution addendum and honest bounds

The host stopped starting several fresh executables (0.00 CPU and no test RUN output after minutes). Reviewer terminated all its remaining command trees and drained them; each ended with exit 137, not a test result. No command is left running.

- Focused install/envprofile invocation ended 137 after install PASS; envprofile did not start. Command: `go test -count=1 -p 1 ./internal/install ./internal/envprofile -run 'Test(RunCommit|Stage.*Targets|PerWriteGuard|ProjectInstallRefusesAdapter|GlobalInstallRefusesAdapter|DraftPath|LegacyPath|AdmitPath)'`. Independent envprofile result UNKNOWN; attached hosted evidence covers it.
- M2, retain planned ancestor tokens but re-stat admitted inputs in Durable: `go test -count=1 -p 1 -overlay .temp/review-14hsti-rev8/mutant-admitted.json ./internal/staging -run 'TestDurableKeepsPlanned' -v`. Exit 137 before RUN output: UNKNOWN, neither killed nor survived.
- M3, retain recovery guard for entry targets but skip bytes: `go test -count=1 -p 1 -overlay .temp/review-14hsti-rev8/mutant-recovery.json ./internal/transaction -run 'TestRecovery(MustRecheckPhysicalBoundary|WithUnchangedBoundaryPublishes|LegacyJournalSkipsGuard)' -v`. Exit 137 before RUN output: UNKNOWN.
- Windows-algorithm capture probe: `go test -count=1 -p 1 -overlay .temp/review-14hsti-rev8/probe.json ./internal/transaction -run TestReviewWindowsCaptureCoherence -v`. Initial attempt's hook compared an uncanonicalized temporary path; corrected to EvalSymlinks(parent) before the second build. Both attempts never started. Corrected retry with `-ldflags=-linkmode=internal`, and a temporary ad-hoc-signed copy of the corrected test binary, also never started; all stopped with exit 137. No native Windows or emulated runtime pass/fail is claimed. F1 is based on the unchecked independent reads in actual Windows production code and the explicit swap/restore schedule, not these stalled executions.

Measured independent attacks: 1/3 killed, 0/3 survived, 2/3 unknown (host startup stalls). All overlays compiled. Candidate files still match 28/28; no source restoration necessary. Probe archive carries source overlays, JSON Replace mappings, and hash manifest, not binaries. JSON paths target this worktree; regenerate absolute paths if extracted elsewhere.

Current task-board wrapper also stalled; read-only goal/checklist/directives succeeded using the already-installed previous CLI `/Users/administrator/.local/bin/task-board-main-6cb09a23-curatorlike --no-update-check`. It reported no goal and no directives. This same compatible CLI is used for evidence and verdict writes; no installation or wrapper change was made. LOGBOOK.md edits remain prohibited; board notes hold the finding.

Lifecycle: both outcome resources attached before verdict routing. set_status(to-dev) succeeded (exit 0). The additionally requested reviewer handoff returned exit 1: `role "reviewer" has no end_status and cannot use handoff`. The explicit changes-requested status plus task-scoped verdict is persisted; the task is not left reviewing. No acceptance attempted.
