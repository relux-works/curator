# TASK-260922-1t2w1q revision 3 review

Verdict: CHANGES_REQUESTED. Route to `to-dev`. Ordinary implementation rework; no human decision required.

Candidate c459d79b7b0d922d8b0e645161683048196f3f5e; base 637e5b43608b76664ed36cedc8f484c389669a2e. All 11 changed files byte-match the candidate in the assigned workspace and disposable git-archive review copy. No candidate code modified or committed. Revision 1 to 3 changes are confined to the migration implementation, tests, repair hints and documentation (9 paths).

## Required fixes

1. **P1: recovery erases marker drift and accepts the stale plan.** `internal/envprofile/migrate.go:914` invokes recovery before the hash comparison at :930. Recovery at :1510-1515 unconditionally republishes journal Prior bytes, without checking that current bytes equal either the recorded prior or intended new state. `TestReviewerRecoveryMarkerDrift` drives PlanMigration -> ApplyMigration with the existing crash injection (one of two link operations applied), edits the codex marker to another valid profile.lock_sha256, then retries ApplyMigration with the original hash. Actual: edited marker overwritten; apply succeeds with nil error because recovery erased the drift. Validate the entire recovery inventory before any recovery write; refuse unknown marker/link states with the exact operator action, preserve them and the journal, and only reconcile states proven to belong to this transaction. Include unexpected symlink targets as well as marker edits in refusal coverage. Keep the ordinary interrupted-apply success row.

2. **P1: recovery cleanup deletes unrelated regular files.** `internal/envprofile/migrate.go:1259-1265` globs `.migrate-*.tmp` and blindly os.Remove's every match; it runs for every journal operation directory through reconcileJournalLinks (:1214-1216), including an operation not yet executed. `TestReviewerRecoveryPreservesRegularTemp` interrupts after the first of two operations, places a regular file containing an operator backup sentinel at `pi/.migrate-operator-backup.tmp`, then retries ApplyMigration. Actual: apply succeeds and the regular file is gone. Track exact temporary-link ownership in the journal and verify the expected shape/target before cleanup; do not delete arbitrary matching regular files, directories, or foreign symlinks. Add preservation/refusal rows through ApplyMigration recovery, plus a positive owned-temp cleanup row. A glob is not ownership evidence.

Both reproductions use temporary stores and the production ApplyMigration entry. These are new recovery-path failures, not regressions in the four straight-line revision-1 probes. Attached Go source reproduces them by copying it to internal/envprofile/reviewer_recovery_test.go in a disposable candidate and running `go test ./internal/envprofile -run TestReviewerRecovery -count=1 -v`.

## Independent verification

Shell zsh; bounded Go timeouts; -count=1. Exact candidate production/test files used; reviewer additions only in disposable copy.

- `go test ./internal/envprofile -run 'TestMigrate|TestCredential|Test.*Passthrough' -count=1 -timeout=180s -v`: PASS, 48.975s, exit 0.
- `go test ./cmd/curator -run 'TestEnvMigrate|TestEnvResolveRepairNeedsMigration' -count=1 -timeout=180s -v`: PASS, 53.793s, exit 0.
- `go test ./cmd/curator -run TestReviewerMigrationCLI -count=1 -timeout=120s -v`: original attached reviewer CLI probes, missing-plan and print-before-write, both PASS, exit 0.
- `go test ./internal/envprofile -run 'TestReviewer|TestMigrateFailedRelinkKeepsOldLink' -count=1 -timeout=120s -v`: exit 1, 8.592s. Original TestReviewerMarkerDrift PASS; TestMigrateFailedRelinkKeepsOldLink PASS (the producer port of original TestReviewerFailedRelinkRollback, whose private helper signatures changed). Existing F-C1 reviewer rows also PASS. New recovery probes fail 2/2 contract assertions as above. No compilation-error failure counted.
- Exact delta `git diff --check`: PASS.
- Independently queried GitHub run 35705009860: success, head 86bada526a5884225bef2ea0f629be2a0353acf6 resolves to exact candidate tree c459d79b7b0d922d8b0e645161683048196f3f5e. Hosted Ubuntu/macOS/Windows test jobs, Ubuntu/macOS race jobs, lint, naming, interop and gate self-tests succeeded. Rose-air and candidate-suite skipped. Reused this full-suite evidence; no redundant full landing suite run.

Read F-C1 results and accepted rev4 verdict, F-C2 results and rev1 verdict, and normative environments migration section. The ordinary path now requires a plan, prints it before link mutation, uses temporary symlink plus rename, journals intent, and includes raw marker digest in its hash. The syscall-failure, marker drift, print-failure, no-copy, normal recovery and no-silent-repair rows pass. Frozen marker schema unchanged. Production migration ReadFile sites read marker/journal, not credential payloads; the cleanup deletion defect remains independently disqualifying.

## Bounds

This rejection does not certify every AC. No independent mutation campaign was replayed: producer reports six mutant kills, but skip-journal and skip-lock kill evidence is not established here (skip-lock explicitly unrun in results). Recovery is exercised with the existing crash fault, not an actual killed OS process. No-copy holder scanning excludes manager state, so it does not prove absence of copies everywhere. Local checks ran on Darwin; hosted-platform claims are limited to exact-tree gate evidence. Inspect/plan only report pending recovery and do not restore it. No host credential data or control-root LOGBOOK was touched.

Run goal queried: none. Supporting logs, probes and task-scoped logbook attached before routing. Fresh producer revision and review required.
