# TASK-260720-1nlmvv — cycle 6 tester evidence

## Verdict

The focused cycle-6 boundary is **ready for review**.

The production-owned `quarantine-rollback-sync` fault point runs after the
compensating rename restores the live cache slot and before the parent
directory is synced. A failure at that second durability boundary preserves the
recoverable live bytes and becomes a key-bearing `StateChangedError`.
Durably compensated failures remain ordinary errors.

No product code or test code was edited by this tester. The preserved task
worktree remains unstaged and uncommitted.

## Static audit

- `internal/buildcache/publish.go:110-151` defines
  `faultQuarantineRollbackSync`, `unsyncedRollbackError`, and `stateChangeFor`.
  The conversion is intentionally narrow: only the unsynced compensating
  rollback becomes `StateChangedError`; all other errors pass through.
- `internal/buildcache/publish.go:493-556` performs the live-to-quarantine
  rename, attempts the first parent sync, restores the entry to the live slot
  on failure, then invokes `faultQuarantineRollbackSync` before the second
  parent sync. A second sync failure returns an empty moved path with the live
  bytes preserved and durable state typed as uncertain.
- `Publish` records every non-empty moved path before acting on an error and
  re-keys an unsynced rollback through `stateChangeFor`. It returns no adoptable
  `PublicationResult` on this failure.
- Public `Quarantine` also applies `stateChangeFor`. `withdrawEntry` preserves
  the non-empty moved-path-plus-error recovery contract. `restoreDisplaced`
  retains wrapped error identity, while `Revert` marks every failed reversal as
  key-bearing changed state.
- `internal/install/commit.go:731-755` derives retained cache state from
  `buildcache.StateChanged` on publication failure.
  `internal/install/commit.go:832-840` carries failed reversals through the
  build boundary. `Result.failCommit` routes that boundary through
  `Result.failBuild`, which applies the shared bounded/path-redacted diagnostic
  renderer before populating `Result.Errors`.
- The focused tests assert exact fingerprints/artifact bytes, one live slot,
  no private staging residue, no adoptable publication result, and the
  changed-state versus durably compensated distinction.

No additional mapping test was needed beyond the producer-added direct
`Quarantine` discrimination test. It was selected as the one permitted optional
mapping test because the primary cycle-6 regression exercises `Publish`, not
the direct public `Quarantine` caller.

## Standalone focused gates

Every command ran directly as a standalone process, sequentially. The recorded
exit code is the process's real exit code.

| Command | Exit | Result |
| --- | ---: | --- |
| `go test ./internal/buildcache -run '^TestAQuarantineRollbackThatCannotBeMadeDurableReportsChangedState$' -count=1` | 0 | green, package time 0.575s |
| `go test ./internal/buildcache -run '^TestAFailedPublicationRestoresAPredecessorMovedBeforeQuarantineError$' -count=1` | 0 | green, package time 0.291s |
| `go test ./internal/buildcache -run '^TestAFailedReversalReturnsTheWinnerMovedBeforeWithdrawError$' -count=1` | 0 | green, package time 0.280s |
| `go test ./internal/buildcache -run '^TestAQuarantineThatCannotBeMadeDurablePutsTheEntryBack$' -count=1` | 0 | green, package time 0.257s |
| `go test ./internal/buildcache -run '^TestADurabilityFaultInsideAQuarantineIsCompensatedByItsCaller$' -count=1` | 0 | green, package time 0.288s |
| `go test ./internal/buildcache -run '^TestAQuarantineThatCannotPutTheEntryBackHandsItsCallerTheRecord$' -count=1` | 0 | green, package time 0.294s |
| `go test ./internal/install -run '^TestAPublicationThatChangedTheCacheIsReportedAsRetained$' -count=1` | 0 | green, package time 0.605s |
| `go test ./internal/install -run '^TestAReversalThatDidNotCompleteIsReportedAsRetained$' -count=1` | 0 | green, package time 0.443s |
| `go test ./internal/buildcache -run '^TestAQuarantineReportsChangedStateOnlyWhenItsRollbackIsNotDurable$' -count=1` | 0 | green, package time 0.304s |
| `gofmt -l internal/buildcache/publish.go internal/buildcache/compensation_test.go` | 0 | empty output |
| `git diff --check` | 0 | no whitespace errors in tracked delta; task buildcache files remain untracked against base, so `gofmt -l` is their formatting evidence |
| `go build ./internal/buildcache ./internal/install ./cmd/curator` | 0 | affected packages build |
| `go vet ./internal/buildcache ./internal/install ./cmd/curator` | 0 | affected packages vet clean |

## Scope and evidence exclusions

- The operator-cancelled cycle-6 `go test ./... -count=1` attempt is incomplete
  and is not evidence.
- No package-wide `cmd/curator`, package-wide `internal/install`,
  repository-wide, race, lint, performance, coverage, cache-clearing,
  timeout-increase, or parallel test command was run in this tester cycle, per
  the focused completion directive.
- Coverage was not rerun in cycle 6 because the focused directive explicitly
  prohibited it. Preserved cycle-3 spawn evidence records these exact
  standalone green commands:
  - `go test ./internal/buildcache -run '^(TestRevertRestoresExactlyWhatAPublicationDisplaced|TestRevertFailsClosed)$' -count=1 -coverprofile=.temp/TASK-260720-1nlmvv/buildcache-focused.cover` — exit 0, package coverage 40.6%.
  - `go tool cover -func=.temp/TASK-260720-1nlmvv/buildcache-focused.cover` —
    exit 0, affected `buildcache.(*Store).Revert` statement coverage 85.0%.
  Cycle 6 adds only the secondary rollback-sync branch; the focused tests above
  exercise both its successful second sync and its typed second-sync failure
  through `Publish`, direct `Quarantine`, and `Revert`. This preserved measured
  evidence plus exhaustive new-branch coverage satisfies the approximate
  affected-code target without violating the no-new-coverage directive.
- Previously accepted diagnostics/currentness/repair behavior and prior
  install/upgrade E2E evidence were preserved and not re-executed.

## Lifecycle evidence

The first cycle-6 `task-board handoff TASK-260720-1nlmvv --role tester`
attempt exited 1 because checklist item 19 lacked linked coverage evidence. No
check was manufactured and no prohibited gate was run. The attached cycle-3
tester artifact and spawn log were audited instead, yielding the exact green
coverage commands and 85.0% affected-function result above.
