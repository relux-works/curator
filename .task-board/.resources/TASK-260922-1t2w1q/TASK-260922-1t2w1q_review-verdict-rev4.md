# TASK-260922-1t2w1q revision 4 review

Verdict: ACCEPTED for producer integration. No delivery-blocking finding in the revision-4 recovery rework. This does not mark delivery done.

## Exact candidate and scope

Candidate tree 2c2b4e0d0b06e347ced5b2e2f9c1a95f8635407b; base 637e5b43608b76664ed36cedc8f484c389669a2e. All 12 changed paths byte-match the assigned workspace and disposable git-archive review copy. Revision 3 (c459d79b7b0d922d8b0e645161683048196f3f5e) to revision 4 changes only migrate.go, migrate_test.go and the committed reviewer_recovery_test.go. Candidate code was not modified or committed; extra probes and mutations ran in an ignored disposable copy.

Independently queried GitHub run 35709733177: success, head 14528997ab83a02c3dd0f8141db21384e48294d1, whose tree equals the candidate exactly. Hosted Ubuntu/macOS/Windows test jobs, Ubuntu race, lint and gate self-tests succeeded. Rose-air and candidate-suite jobs were skipped, not passing. Build/lint/full-suite evidence reused from this exact-tree gate; no full landing-suite replay.

## Recovery findings

- recoverMigrationJournal calls validateRecoveryInventory before recovery writes. All journaled marker bytes must equal prior or intended bytes, all links must match recorded prior/intended state, and recorded temporary paths must be absent or the expected symlink. Unknown states refuse with operator action while retaining the journal. Marker and link checks repeat near mutation. The two prior reviewer rows pass, and ordinary interrupted recovery still converges before the plan-hash recheck.
- atomicRelinkJournaled records the exact temporary path and target before creating it; symlink plus rename replaces the directory entry atomically. removeOwnedTemp checks type and full target; there is no glob cleanup. Owned-temp positive coverage passes. Independent production ApplyMigration probes at the recorded temp path cover regular file, directory and foreign symlink: all three refuse and preserve the complete credential-scope snapshot and journal.
- Narrowed validation covering only Done operations is killed by the unexpected-target row (the pending operation must also be validated before any write). Reintroduced broad glob cleanup is killed by the foreign regular-temp preservation row. 2/2 selected mutants killed by behavioral assertions, no compilation-error kills. Production source restored byte-identically afterward.
- Production credential operations remain metadata/symlink operations. New ReadFile calls inspect manager markers/journal, not native credential contents. Focused no-copy, Pi mode/root, conflicts, drift, syscall rollback, required-plan, temporal print-before-write, and no-silent-repair tests pass. Frozen v1 schema unchanged. Earlier straight-line acceptance remains under the binding revision-4 note; docs and CHANGELOG unchanged by this recovery-only rework.

## Independent runs

Shell zsh; -count=1 and bounded timeouts. All tests use temporary stores.

1. `go test ./internal/envprofile ./cmd/curator -run 'Test(Migrate|ReviewerRecovery|EnvMigrate|EnvResolveRepairNeedsMigration)' -count=1 -timeout=180s`: exit 0; envprofile 12.034s, CLI 33.077s. Includes both prior reviewer recovery rows and committed replacements for the original straight-line probes.
2. In frozen disposable copy, `go test ./internal/envprofile -run 'Test(ReviewerRecordedTemp|ReviewerRecovery|MigrateRecovery|MigrateInterruptedApplyRecovers|CredentialLink|Dangling|SharedToIsolated|CodexIsolated)' -count=1 -timeout=180s -v`: exit 0, 5.692s. Complete log and three additional probe functions attached.
3. Selected mutants, each with -count=1 and timeout 90s: validate-only-done-links -> TestMigrateRecoveryRefusesUnexpectedTarget FAIL (exit 1); glob-cleanup -> TestReviewerRecoveryPreservesRegularTemp FAIL with unrelated regular file deleted (exit 1). Mutation script attached. Restored run in item 2 passes.
4. `git diff --check`: exit 0. Gate commit tree and all changed-file byte identities verified independently.

One initial extra-probe invocation used a nested-module package path from the outer module and failed setup; corrected by running from the disposable module directory (item 2). That setup failure is not counted as passing.

## Bounds

Review focuses on revision-4 recovery plus targeted regression reruns; it is not exhaustive clause mutation coverage. Existing crash rows simulate interruption through the production fault hook; owned-temp state is crafted, not an actual SIGKILL between syscalls. Concurrent operator edits between final checks and filesystem syscalls are not atomic-CAS protected. Windows/other hosted claims come from the verified gate, not local Darwin execution; rose-air remains unverified. Prior producer results explicitly bound the unrun skip-lock mutant; lock acquisition remains visible at ApplyMigration entry. No host credential changes or control-root LOGBOOK writes were made.

Read F-C1 results/review and F-C2 prior review/rework evidence. Run goal queried: not goal-bound; no directives. Attach this task-scoped verdict before accept_cr(revision=4); subsequent checkpoint/integration belongs to the tracked producer.
