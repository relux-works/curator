# TASK-260927-1wc76r review verdict — CR rev1: ACCEPTED

Reviewer: claude-opus-5-5 (low). Candidate tree 9734f6765d2b8dc242357c59c130b6b27d1efbf2 == worktree (temp-index write-tree), base 0ffe2e1d.
`git status`: only product/test/docs/.github/ci paths; no CHANGELOG/LOGBOOK edit.

## Spec (curator-spec v1.0.0-rc.13, tag 23435129)
- environments §9.2 `env unmanage [--restore-backups] [--env] [--target]`: remove marker-recorded surfaces, copy newest §8.3 generation back before marker delete, never touch unrecorded/credential files, clear scope's current profile → unmanage.go Unmanage/planUnmanageHomes/applyUnmanagePlan, credentialPath skip, unmanageCurrentRemovals.
- §8.3 backup paragraph + §8.4.1 backup-record row: inventory unknown → `environment_backup_record_unreadable`, restore stops before mutating; absent → nothing to restore → newestBackupFiles (stateread seam) runs in planning, before any write; preflight precedes apply.
- --target refused as deferred (secondary fixed-home writes not implemented) — acceptable bound.

## Vectors (rc.13)
Only two rc.13 vectors involve unmanage/restore: backup-record-unreadable-restore-stops, backup-record-absent-restore-nothing (environments-write-nofollow's only related case is `takeover`). Both driven through `cli.run(env unmanage --restore-backups)`; case-count row `environments-read-failure/restore 2` added.
Gap rows: at base 0ffe2e1d no gap/xfail row referenced these vectors (git grep); before = 0, after = 0 — nothing to remove (they were simply unconsumed).

## Independent runs (zsh, pipefail, CURATOR_CONFORMANCE_ROOT = git archive of rc.13 tag)
`go test ./cmd/curator -run TestEnvUnmanage -count=1 -v` → ok (both subtests + 4 tests pass). `go vet ./cmd/curator ./internal/envprofile` exit 0; gofmt clean. internal/envprofile full package not rerun (change there is one const); hosted gate is arbiter.

## Mutants (disposable rsync clone)
- M1 readDir error on inventory → treated as absent: KILLED (TestEnvUnmanageBackupRecordVectors).
- M2 restore without observed record (absent lstat view falls through to physical re-read): KILLED (TestEnvUnmanageDoesNotRestoreWithoutAnObservedRecord).
- M3 partial write before refusing (surface removed on inventory error): KILLED.
- M1b Lstat *error* on inventory root → treated as absent: SURVIVES. Residual: no test injects a BackupLstat error (the rc.13 vector injects at listing). Non-blocking; follow-up row suggested.

## Other residuals (non-blocking)
- applyUnmanagePlan is not journaled; a mid-apply I/O failure can leave partial state. §9.2/§8.4.1 require stop-before-mutating only for read failures, which holds (planning+preflight precede writes).
- Restored files written 0644 (original mode not recorded in backups).
