# TASK-260910-14hsti revision 6 — CHANGES_REQUESTED

Reviewed candidate tree `0029945602c8638fbcd253f62fbeaa255f8036ab`, base `12f1287ee0fb538f9ca004dd53b870e824e5baf2`. All 26 changed worktree paths matched the candidate bytes before checks and remained unchanged after all overlay attacks. No repository code, index, branch, runtime home or LOGBOOK.md was modified.

## Findings requiring rework

1. HIGH — Restart/recovery bypasses the publication boundary guard. `internal/transaction/engine.go:320` resumes prepared/committing journals by calling commit; `:482-485` treats the absent in-memory guard as permission to write. A fresh Engine always has an empty guards map. This violates skillfile-sources §2's recheck immediately before EACH publication write, including writes not attempted before restart. Preimage/digest verification cannot detect same-spelling parent identity replacement. The producer documents this as a bound, but the binding scope ruling exempts only the 16k7xy snapshot-store seam, not existing transaction recovery.

   Reproduction through real Engine.Prepare and a fresh Engine.Recover: prepare an old→new file replacement with a staging physical-boundary snapshot; rename parent→parent-old; mkdir parent; move original children (including sidecars) back so bytes and names remain identical. The retained original guard refuses `destination ancestor ... changed since planning`; the restarted engine's Recover returns nil and publishes `new`. Reviewer overlay test `TestReviewRecoveryMustRecheckPhysicalBoundary` exits 1 with that exact assertion. No journal stub involved.

   Required: make a boundary-protected journal distinguishable from a legacy unguarded journal and preserve/reconstruct a trustworthy boundary proof across restart, or fail closed without further publication when it cannot be restored, with safe rollback/recovery behavior. Do not silently reinterpret missing protected evidence as legacy. Add a committed restart regression exercising the real journal and same-spelling swap, plus unchanged-boundary and legacy controls.

2. MEDIUM — Identity/path-inspection failures remain mislabeled as proven overlap. `internal/staging/boundaries.go:405,426,432,441` (also Snapshot/admitted and adapter wrappers) returns `source_output_overlap` for arbitrary inspection errors. Revision-5 failure instructions and revision-6 review note explicitly require an identity-probe failure never to be reported as overlap. The Windows Lstat fix addresses the access-denied cause, but no distinct unreadable identity classification exists.

   Reproduction: Snapshot a target; replace its parent with a self-referential symlink; call Plan.Recheck. It returns `source_output_overlap: ... EvalSymlinks: too many links` although no overlap was established. Reviewer overlay `TestReviewIdentityReadErrorIsNotOverlap` exits 1. Required: preserve a distinct fail-closed diagnostic for inability to inspect/prove identity (for example the explicitly authorized boundary_identity_unreadable), propagate it through planning and publication, and test inspection failures separately from proven overlap. Keep the Windows directory/mirror regression green.

## Independent verification

Shell: zsh. Commands below run independently by this reviewer, without pipelines; each exit is the command result.

| Command | Exit |
|---|---:|
| go test -count=1 ./internal/snapshot ./internal/staging ./internal/privatedir ./internal/adapters | 0 |
| go test -count=1 -p 1 ./internal/install -run 'Test(StageProjectTargets|StageGlobalTargets|RunCommit.*(Boundar|Recheck|RealStaging)|PerWriteGuard|ProjectInstallRefusesAdapter|GlobalInstallRefusesAdapter)' | 0 |
| go test -count=1 -p 1 ./internal/envprofile -run 'Test(DraftPath|LegacyPath|AdmitPath)' | 0 |
| go test -count=1 -p 1 ./internal/transaction -run 'TestCommit.*Boundary|TestCommitRollbackNever' | 0 |
| go vet ./internal/snapshot ./internal/staging ./internal/privatedir ./internal/adapters ./internal/envprofile ./internal/install ./internal/transaction | 0 |
| gofmt -l on those seven package directories | 0, empty output |
| git diff --check | 0 |
| go test -count=1 -p 1 -overlay recovery.json ./internal/transaction -run TestReviewRecoveryMustRecheckPhysicalBoundary -v | 1, reproduced F1 |
| go test -count=1 -p 1 -overlay diagnostic.json ./internal/staging -run TestReviewIdentityReadErrorIsNotOverlap -v | 1, reproduced F2 |

Hosted evidence accepted only for its recorded scope: TASK-260910-14hsti_change-request_rev6-validation.log, run 35111601714, https://github.com/relux-works/curator/actions/runs/35111601714. Remote gate exit 0; Test Ubuntu/macOS/Windows, Race Ubuntu/macOS, lint and conformance/naming gates successful. Local git resolves hosted commit 0be2e073bc8c95e19cada8f65eca12509fb79d52 to the exact reviewed candidate tree. Rose-air and Candidate suite lanes skipped; no ARM64 execution claim. Full suite was not rerun locally. Log reports test_case_coverage=unknown; aggregate CI success is not proof of the two uncovered regressions.

## Narrowing mutants

Temporary Go -overlay replacements only; original bytes never changed. 2/2 mutants killed, 0 survivors:

- M1: narrow the backup-site refusal to target index 0, keeping the install-site check intact. `go test -count=1 -p 1 -overlay mutant-backup.json ./internal/transaction -run TestCommitBoundaryRefusalAtBackupRollsBackPublishedTargets` exits 1: refused target reached backup; backedUp=[0 1].
- M2: narrow stateForPath draft admission to name == default. `go test -count=1 -p 1 -overlay mutant-acquisition.json ./internal/envprofile -run 'TestDraftPath.*RefusesStoreOverlap'` exits 1: both root install and overlay install wrongly succeed.

Attached TASK-260910-14hsti_review-probes-rev6.tar.gz contains replacement files and overlay maps for reproducibility. Maps contain this host's paths; remap Replace keys/values when replaying elsewhere. hashes.json pins the original 26 candidate files. Overlay test additions are reviewer probes, not shipped regressions.

## Acceptance evidence and bounds

- Physical symlinks/case: TestValidateSymlinkManagedAndEscape, TestValidateCaseAliasUsesFilesystemIdentity, staging identity/case suites; production Project/Global via TestProjectInstallRefusesAdapterDestinationOverSnapshot and TestGlobalInstallRefusesAdapterDestinationOverSnapshot. Windows Lstat fix is present; hosted Windows is green. This host alone does not prove Windows behavior.
- Authored versus managed output and safe path-dot subdirectory: TestPrepareLocalAcquisitionAdmitsSafeSubdirectory, TestPrepareLocalAcquisitionRefusals, TestValidateManagedMissingRefusedAsOverlapBeforeTraversal. Adapter roots and production groups carry admitted snapshots.
- root_inputs: TestPrepareLocalAcquisitionAdmitsRootInputs and refusal checks, plus TestValidateRootInputsRefusals at helper level. The full root-input configuration matrix is not proven through an integrated draft install in this leaf. The future capture/store consumer belongs to TASK-260910-16k7xy; its existing seam is PrepareLocalAcquisition / LocalAcquisition in internal/snapshot/boundaries.go. This scoped seam is NOT a finding.
- Existing local-path acquisition: draft-enabled Install and overlay tests reach stateForPath → admitPathSource → PrepareLocalAcquisition. TestLegacyPathInstallIgnoresStoreOverlap pins switch-off behavior.
- Publication: TestPerWriteGuardRefusesSwappedParentAndRollsBack drives real staging + journal; engine tests cover backup and install writes with rollback. Recovery path is NOT covered by those existing tests and fails the reviewer probe.
- Unmanaged paths: TestUnmanagedTakeoverRefusedBesideAdmittedCheck and existing adapter tests pass. No schema/parser reimplementation or frozen-v1 wire-schema edit in the candidate; envprofile/install/transaction additions are authorized production wiring.

Overall acceptance is not established: normal-path tests pass, but 2/2 new targeted adversarial probes expose unmet requirements. This is ordinary implementation rework, not an external or human-only blocker. Route to to-dev; independent review required after fixes. Findings are recorded here and in board notes because campaign rules prohibit LOGBOOK.md edits. Reviewer goal query reports this run is not goal-bound.
