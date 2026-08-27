# TASK-260720-31nl14 review cycle 4 verdict

## Verdict

Changes requested. Route to `to-dev`.

The cycle-3 Darwin Unicode-normalization and resumable partial-directory cleanup fixes are present, and their regressions pass. One P1 safety defect remains in the new owned-tomb recovery path.

## Finding

1. **P1 — restart cleanup deletes unknown concurrent bytes added to a transaction-owned partial tomb.** `finishRecordedRemoval` treats a journal `removal_path` as sufficient ownership proof and calls `removeDurablyWithOwnership(..., true, ...)` (`engine.go:737-747`). When a tomb already exists and `ownedTomb` is true, `removeDurablyWithOwnership` skips the expected-digest check (`journal.go:421-446`) and recursively deletes every current entry (`journal.go:472-505`). This permits unknown bytes created after the cleanup crash to be deleted on recovery, contrary to the task requirement to refuse to overwrite unknown concurrent state.

   The task-scoped exported-API probe faults at `PointDuringCleanupRemoval`, writes `foreign-concurrent-bytes` into the partial backup tomb, then resumes through a new engine. It reports:

   ```text
   recovery_error=<nil>
   foreign_exists_after_recovery=false
   foreign_stat_error=... no such file or directory
   ```

   Probe: `.temp/TASK-260720-31nl14/worktree/.temp/review-cycle-4/owned-tomb-foreign-probe.go`.

## Required rework

- Retain durable provenance granular enough to distinguish the expected partial remainder from entries added or replaced after the tomb rename. A durable per-entry manifest/progress model is one viable approach: validate each remaining recorded entry before deletion, accept already-durably-removed recorded entries, and reject/preserve unrecorded or mismatched entries.
- Add direct and restart regressions for foreign bytes added to an owned tomb after `PointAfterCleanupRename` and `PointDuringCleanupRemoval`. Recovery must return `ErrImplementationCorruption`, preserve the foreign bytes, desired live target, journal, and usable recovery state.
- Cover both commit cleanup and rollback cleanup. Existing recovery of an unmodified full or partial owned tomb must continue to complete.

This is ordinary implementation rework, not a stop-the-line or human-only decision.

## Independent validation

Passed on native Darwin/arm64:

- `go test ./internal/transaction -count=1`
- `go test -race ./internal/transaction -count=1`
- `go vet ./internal/transaction`
- `go test -cover ./internal/transaction -count=1` — 78.2% statement coverage
- `make check`
- `go test -race ./... -count=1`
- Linux/amd64 complete compile graph via `go test -exec=true ./...`
- Windows/amd64 complete compile graph via `go test -exec=true ./...`
- `golangci-lint v2.4.0 run ./internal/transaction/...` — 0 issues
- `gofmt`, `git diff --check`, and staged-file checks

No product code was modified or staged during review. Native Windows runtime remains unavailable on this Darwin host; Windows compilation is not runtime evidence, and the inherited `TASK-260720-1zl1cj` qualification gate remains preserved.
