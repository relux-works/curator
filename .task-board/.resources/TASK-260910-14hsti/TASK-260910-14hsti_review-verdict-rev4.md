# TASK-260910-14hsti revision 4 — CHANGES_REQUESTED

Candidate: d9db72b6cddbf847fa94f2181937cf91f8951f8d; base: 12f1287ee0fb538f9ca004dd53b870e824e5baf2. All 12 candidate files matched the review workspace byte-for-byte before verification. Reviewer made no product-code changes; adversarial variants use Go overlays in ignored .temp only.

## Findings

1. HIGH — Rework item 1 remains incomplete: the new boundaries are not fed by real acquisition/planning. `PrepareLocalAcquisition` (internal/snapshot/boundaries.go:73) has no non-test caller; `scopeTargets.boundaries` (internal/install/commit.go:517) is never populated outside tests, and production adapter groups (internal/install/install.go:752) omit Admitted. Consequently the publication guard at commit.go:657 is disabled on every actual caller. Tests in commit_boundaries_test.go inject scopeTargets via a stub stageTargets callback, and acquisition tests directly call the newly added helper. These prove conditional seams, not production installation enforcement. Comments explicitly defer planning to a future integration leaf; that does not fulfill this task's production-path/rework acceptance. Wire the actual draft acquisition and planner to these records and test the public production path without injecting its boundary evidence. Preserve the legacy path.

2. HIGH — Publication protection is a one-time pre-journal check, not a check immediately before each publication write as skillfile-sources section 2 requires. `runCommit` calls Recheck once before journalPlan; journalPlan then creates directories and records fresh preimages, and the transaction engine commits targets separately. Neither transaction.Plan nor engine.commitTarget receives this boundary snapshot. A boundary changed after that single check is not compared with its planned identity before subsequent writes. Carry enforcement to the actual write boundary (including rollback of unpublished state); add a real-journal fault-injection regression that swaps a parent after planning/checking and before a later write. The existing stubJournal tests cannot establish this property.

## Independent verification

Shell: zsh. No whole-module or hosted-gate rerun.
- `go test -count=1 ./internal/snapshot ./internal/staging ./internal/privatedir ./internal/adapters`: exit 0, all four packages pass.
- `go test -count=1 ./internal/install -run 'TestRunCommit(Recheck|WithoutBoundaries)'`: exit 0.
- `go vet ./internal/snapshot ./internal/staging ./internal/privatedir ./internal/adapters ./internal/install`: independent standalone rerun exit 0, no diagnostics.
- `gofmt -l` on changed Go files and relevant test files: empty, exit 0.

## Coverage bounds

The diff is within the named packages plus internal/install as explicitly required by rework; frozen wire schemas are untouched. New APIs are labelled draft/revision 1. Acquisition, capture, locking, and schema-2 integration are not established by this leaf. Two of two inspected production data-feeding connections are missing (acquisition invocation and publisher boundary population).

Named seam tests: TestPrepareLocalAcquisitionAdmitsSafeSubdirectory; TestPrepareLocalAcquisitionAdmitsRootInputs; TestPrepareLocalAcquisitionRefusals; TestPrepareLocalAcquisitionValidatesBeforeStaging; TestValidateCaseAliasUsesFilesystemIdentity; TestStageProjectRefusesMirrorOverAdmittedInput; TestStageProjectAdmitsDisjointAdmittedInputs; TestUnmanagedTakeoverRefusedBesideAdmittedCheck; TestRunCommitRecheckRefusesDestinationInsideAdmitted; TestRunCommitRecheckRefusesSwappedParent; TestRunCommitRecheckPassesUnchangedBoundaries. The unmanaged protection test reaches existing StageProject behavior. The other new conditional checks still need actual application data flow.

Same-spelling identity pinning is implemented and baseline tests pass; Windows eager SameFile pinning is consistent with the local Go Windows implementation. Cross-platform success is not inferred from local execution.

Run goal queried: no active goal (run is not goal-bound). Verdict is ordinary implementation rework, not an external blocker.

## Hosted evidence read, not replayed

Read producer results for revisions 1, 3 and 4 and TASK-260910-14hsti_change-request_rev4-validation.log. Hosted run https://github.com/relux-works/curator/actions/runs/35090680604 reports success, exit 0: Test ubuntu/macos/windows, Race ubuntu/macos, lint, naming, interop, gate self-tests. Rose-air and Candidate suite lanes were skipped. Gate commit 069ad0f6ef57e00e67af2e08b8ee202af60bc19d resolves locally to candidate tree d9db72b6cddbf847fa94f2181937cf91f8951f8d. This is accepted exact-tree platform test evidence, not independent local cross-platform execution. Rev3/4 producer notes explicitly acknowledge the two deferred behaviors in the findings; those bounds conflict with acceptance and the requested rework.

## Independent narrowing attempts and execution limitation

Two Go overlays were prepared without editing product bytes:
1. First snapshot output refusal narrowed from `else if pruned` to `else if pruned && directory == "."`. Command: `go test -count=1 -overlay .temp/review-14hsti-rev4/managed.json ./internal/snapshot -run 'TestValidateManaged|TestPrepareLocalAcquisition'`.
2. Physical identity refusal narrowed to non-directories (`!os.SameFile(recorded, current) && !current.IsDir()`). Command: `go test -count=1 -overlay .temp/review-14hsti-rev4/identity.json ./internal/staging ./internal/install -run 'Test.*(Recheck|Swapped|Replaced)'`.

Both compiled but the new test processes remained at 0 CPU time with no output for over five minutes on this host. Terminated these owned processes; both commands returned exit 143. Results are UNKNOWN, not kills or survivors: 0/2 completed attacks, 2/2 unverified. Producer mutant results were read but are not substituted for independent executions. Baseline narrow tests above did complete successfully. All 12 candidate paths still match exact tree bytes; overlays never changed them, so no source restoration was needed. The new default task-board executable also stalled; reads and evidence attachment use the already-installed task-board-main-6cb09a23-curatorlike with --no-update-check, without installing or modifying host state.

## Required next cycle

Supply real acquisition/planning data flow and per-write physical-boundary enforcement, drive them through actual application and journal entry points, then repeat narrow tests and independent mutant checks. Keep the eager Windows identity pin and same-spelling replacement regressions. Route to to-dev; no acceptance or commit acknowledgement. These are implementation rework findings, not human-only decisions. This outcome records the findings in lieu of a LOGBOOK.md edit, which campaign instructions prohibit.
