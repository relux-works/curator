# TASK-260910-14hsti revision 7 — CHANGES_REQUESTED

Candidate: bc77cd0b123d6909845d05161b0aa067594fa5d8. Base: 12f1287ee0fb538f9ca004dd53b870e824e5baf2. Independent reviewer run RUN-260916-1bbf4f. Goal query: not goal-bound (checked at start and before verdict). All 28 candidate files compared byte-for-byte with git tree before testing and after overlay attacks: 28/28 match. No production or committed test file edited. Tests below used zsh, direct commands, observed process exit codes; no background process left running.

## F1 HIGH — Durable conversion replaces the original boundary identity

Location: internal/staging/boundaries.go:527–552, especially FileIdentity calls at 533 and 545; production consumer internal/install/commit.go attachBoundaries and transaction Engine.Recover/checkBoundary.

Snapshot.Durable iterates the keys of the stored FileInfo maps but ignores the stored identities. It stats each pathname again and records whatever object occupies that spelling now. The durable proof therefore need not represent the same planning boundary as the in-memory guard. Recovery faithfully verifies the wrong identity.

Independent regression TestReviewRecoveryUsesOriginalSnapshotIdentity, supplied as a Go overlay over transaction/boundaries_test.go, fails through real Engine.Prepare and a fresh Engine.Recover:

1. Snapshot pins original parent A.
2. Rename A away, put replacement B at the same spelling, move identical children into B. The original guard refuses.
3. Call Snapshot.Durable: it succeeds and records B, without comparing against A.
4. Restore A and its children. The pre-journal in-memory check passes (explicit assertion, matching the install path's pre-journal check). Prepare a real protected journal with both guard and durable proof.
5. Put B back at the spelling and move all children, including journal sidecars, into it. The same-process guard again refuses.
6. Fresh Engine.Recover succeeds and publishes live="new". Expected refusal and live="old". Test exits 1 with `recovery published across replaced parent: live="new", want refusal and old bytes`.

This is not journal tampering, inode reuse, or the deferred package-capture seam. The proof is created by the production conversion method and persisted by the real journal. It is a same-spelling identity substitution in the in-scope publication/recovery path. The explicit A-restoration control means the existing pre-journal Recheck does not eliminate this sequence.

Required rework: preserve the original identity in both proof forms. Capture a serializable identity coherently with the planning FileInfo (or serialize that captured identity), rather than recapturing a later pathname without establishing equality to the original. If equality cannot be established, refuse conversion. Apply this to destination ancestors AND admitted inputs. Add committed regressions for conversion after replacement and real-journal restart, plus unchanged and legacy controls. Keep the existing per-write checks and error-class separation.

## What passed independently

| Command | Exit | Result |
|---|---:|---|
| `go test -count=1 -p 1 ./internal/staging ./internal/snapshot ./internal/privatedir ./internal/adapters` | 0 | All four packages pass (0.432s, 1.195s, 0.353s, 0.502s) |
| `go test -count=1 -p 1 ./internal/transaction -run 'Test(Commit.*Boundary\|CommitRollbackNever\|Recovery.*Boundary\|RecoveryLegacy)'` | 0 | Selected publication and recovery tests pass, 4.337s |
| `go test -count=1 -p 1 ./internal/install ./internal/envprofile -run 'Test(RunCommit\|Stage.*Targets\|PerWriteGuard\|ProjectInstallRefusesAdapter\|GlobalInstallRefusesAdapter\|DraftPath\|LegacyPath\|AdmitPath)'` | 0 | Install 12.261s; envprofile 6.869s |
| `go vet ./internal/staging ./internal/snapshot ./internal/privatedir ./internal/adapters ./internal/transaction ./internal/install ./internal/envprofile` | 0 | Clean |
| `gofmt -l internal/staging internal/snapshot internal/privatedir internal/adapters internal/transaction internal/install internal/envprofile` | 0 | Empty output |
| `git diff --check` | 0 | Clean |
| `go test -count=1 -p 1 -overlay .temp/review-14hsti-rev7/probe.json ./internal/transaction -run TestReviewRecoveryUsesOriginalSnapshotIdentity -v` | 1 | F1 reproduced, real journal publishes across changed identity |

## Independent narrowing attacks

Go overlays keep repository bytes untouched. Both narrowings compiled and were killed: **2/2 killed, 0 survivors**. Candidate byte comparison after the attacks was 28/28 identical; no restoration of producer files was necessary.

- M1: in Engine.checkBoundary, keep durable verification for entry targets but skip byte targets. Command: `go test -count=1 -p 1 -overlay .temp/review-14hsti-rev7/mutant-recovery.json ./internal/transaction -run 'TestRecovery(MustRecheckPhysicalBoundary|WithUnchangedBoundaryPublishes|LegacyJournalSkipsGuard)' -v`. Exit 1. TestRecoveryMustRecheckPhysicalBoundary fails (live="new"); unchanged and legacy siblings pass.
- M2: retain boundary_identity_unreadable elsewhere but classify byte-target Canonicalize failures as source_output_overlap in recheckTarget. Command: `go test -count=1 -p 1 -overlay .temp/review-14hsti-rev7/mutant-diagnostic.json ./internal/staging -run 'TestPlanRecheck(IdentityReadErrorIsNotOverlap|RefusesSameSpellingParentReplacement)' -v`. Exit 1. IdentityReadErrorIsNotOverlap fails on the symlink loop; the proven same-spelling-overlap sibling passes.

F2 from rev6 is resolved for the reviewed overlap-versus-inspection-error branches; the committed negative tests and independent narrowing attack enforce that distinction. Original recovery, unchanged-boundary and legacy controls also pass, but their Snapshot→Durable interval is never attacked, explaining why F1 survived the suite.

## Scope, acceptance trace and bounds

Reviewed physical-boundary contract: protocol/skillfile-sources.md §2 and task rework rulings. Changes stay in boundary packages and their explicitly authorized install/envprofile/transaction callers; frozen schemas unchanged. Existing parser/collections are reused. Draft path admission is controlled by Policy.DraftSourcesV1; legacy path test passes.

Acceptance test anchors checked in the candidate:
- Physical paths/symlinks/case and authored-vs-managed selection: TestValidateCaseAliasUsesFilesystemIdentity, TestValidateSymlinkManagedAndEscape, TestValidateManagedMissingRefusedAsOverlapBeforeTraversal; current path caller controls TestDraftPathInstallRefusesStoreOverlap and TestDraftPathOverlayRefusesStoreOverlap.
- Operator root_inputs and broad root with selected safe subdirectory: TestPrepareLocalAcquisitionAdmitsRootInputs, TestPrepareLocalAcquisitionRefusals, TestValidateRootInputsRefusals, TestPrepareLocalAcquisitionAdmitsSafeSubdirectory.
- Pre-traversal refusal: TestPrepareLocalAcquisitionValidatesBeforeStaging; production entry PrepareLocalAcquisition, invoked by envprofile stateForPath via admitPathSource.
- Production destination separation and unmanaged protection: TestProjectInstallRefusesAdapterDestinationOverSnapshot, TestGlobalInstallRefusesAdapterDestinationOverSnapshot, TestUnmanagedTakeoverRefusedBesideAdmittedCheck, TestStageProjectRefusesMirrorOverAdmittedInput.
- Per-write physical recheck: TestPerWriteGuardRefusesSwappedParentAndRollsBack through real stageProjectTargets/runCommit/journal, plus transaction boundary tests. Recovery identity continuity remains incomplete per F1.

This is a focused boundary review, not a claim of all 73 semantic cases or every acceptance clause exhaustively proven. Helper/API tests of root_inputs are bounded by the 16k7xy seam: PrepareLocalAcquisition/LocalAcquisition in internal/snapshot/boundaries.go is the exact future package-capture input, and package byte capture/hash/store publication remain owned by TASK-260910-16k7xy. That scope boundary is not a finding.

Read producer results_rev7 and rev7-validation.log. The latter reports hosted run 35119005389 (gate commit f432fdfffa66b8f2d506a0c3c2898610af3e0e94), exit 0: Ubuntu/macOS tests and race, Windows tests, lint, naming, interop and gate self-tests successful. Accepted as attached hosted evidence, not rerun locally. Rose-air and Candidate suite are explicitly skipped; no ARM64 claim. Local Windows execution not performed; Windows FileIdentity uses BACKUP_SEMANTICS and shares deletion, and the prior Lstat mirror fix remains. Full landing suite was not rerun.

The probe/mutants and candidate SHA-256 manifest are attached as TASK-260910-14hsti_review-probes-rev7.tar.gz. Archive overlay JSON paths refer to this Story workspace; if extracted elsewhere, regenerate Replace paths to the extracted files and target worktree. Findings recorded here and in board notes; LOGBOOK.md remains untouched per campaign rule.

Verdict: CHANGES_REQUESTED; route to to-dev for F1 rework and another independent review. No human decision or external blocker. No accept_cr and no commit acknowledgement.

Lifecycle: explicit set_status(to-dev) succeeded. The brief's additional `task-board handoff ... --role reviewer` returned exit 1 because `role "reviewer" has no end_status and cannot use handoff`. The required explicit changes-requested branch is already persisted through to-dev plus this new task-scoped verdict; no default end status or acceptance was attempted.
