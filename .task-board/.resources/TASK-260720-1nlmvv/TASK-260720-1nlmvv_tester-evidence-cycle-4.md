# TASK-260720-1nlmvv — tester evidence, cycle 4

## Verdict

CHANGES REQUESTED. Route to `development`.

Cycle 4 does not close P1 over failures that occur after `quarantinePath`
renames a live cache entry but before its parent-directory sync succeeds.
The advertised focused gates are green, but two new production-boundary
regressions are expected-red and leave the logical cache slot absent.

No product code was changed. Two test-only regressions were added to
`internal/buildcache/compensation_test.go`.

## Blocking evidence

`quarantinePath` renames `entryPath` to a quarantine path and then calls
`syncDirectory(parent)`. If that sync fails, it returns an empty path plus an
error even though the rename already changed live state
(`internal/buildcache/publish.go:461-467`).

In `Publish`, `displaced` is assigned only after `quarantinePath` returns
success (`publish.go:266-272`). Its deferred compensation returns immediately
when `displaced == "" && !selected` (`publish.go:188-198`). A post-rename
quarantine-sync error therefore returns an empty `PublicationResult`, an
ordinary error, and leaves the previous live slot missing.

`restoreDisplaced` has the mirror hole: it calls the same `quarantinePath` to
withdraw the published winner and returns immediately on its error
(`publish.go:378-385`). A post-rename sync failure leaves the launcher-visible
slot missing. `StateChangedError` reports that state changed, but contains no
recovery path or record with which `runCommit` could restore the winner.

The existing `faultQuarantine` and `faultWithdraw` cases fire before
`quarantinePath`, so they do not exercise these interior post-mutation exits.
The new tests deterministically use the production quarantine helper to make
the rename, then return the boundary error that a failed sync produces.

## Expected-red regression gates

- `go test ./internal/buildcache -run '^TestAFailedPublicationRestoresAPredecessorMovedBeforeQuarantineError$' -count=1`
  — exit **1**. Failure: `live verdict = miss, want the corrupt predecessor restored`.
- `go test ./internal/buildcache -run '^TestAFailedReversalReturnsTheWinnerMovedBeforeWithdrawError$' -count=1`
  — exit **1**. Failure: `live verdict = miss, want the published winner returned`.

These are reported as failing gates, not passes. The expected-failure rationale
is that cycle 4 has no mutation/recovery record for an error returned after the
quarantine rename.

## Green focused gates

- Cycle-4 build-cache publication/reversal tests:
  `go test ./internal/buildcache -run '^(TestAFailedPublicationRestoresTheCacheItDisplaced|TestAPublicationThatCannotRestoreReportsAChangedCache|TestAFailedReversalIsFailClosedAndReportsAChangedCache|TestRevertRestoresExactlyWhatAPublicationDisplaced|TestRevertFailsClosed)$' -count=1`
  — exit **0**, 1.011s.
- Install compensation, live-journal, retention, and redaction tests:
  focused `go test ./internal/install ... -count=1`
  — exit **0**, 6.246s.
- Real install and upgrade exact-cache/install-state E2E:
  `go test ./cmd/curator -run '^TestInstallAndUpgradeRestoreTheCacheWhenTheCommitFails$' -count=1`
  — exit **0**, 94.736s.
- `gofmt -l` over affected cycle-4 files — exit **0**, empty output.
- `git diff --check` — exit **0**.
- `go build ./cmd/curator` — exit **0**.
- `go vet ./internal/buildcache ./internal/install ./cmd/curator` — exit **0**.

No repository-wide, race, performance, shared-cache-clearing, staging, commit,
publication, pin, or host-install operation was run.

## Required rework

Make the quarantine move observable before the parent sync can fail. Either:

1. have `quarantinePath` compensate the rename internally on every later error;
   or
2. return a complete mutation/recovery record on post-rename error and make
   `Publish`/`restoreDisplaced` immediately own and reverse it.

Add a deterministic fault point specifically after the quarantine rename and
before/during its parent sync. The two expected-red regressions must then pass
without leaving an empty live slot, for both publication and reversal paths.
