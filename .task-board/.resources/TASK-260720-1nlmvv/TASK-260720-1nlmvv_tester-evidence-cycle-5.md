# TASK-260720-1nlmvv — cycle 5 tester evidence

## Verdict

**CHANGES REQUESTED.**

Cycle 5 restores the live pathname for the two cycle-4 post-quarantine-rename
failures, and all five requested regressions pass. It does not close the
secondary durability boundary: after the first cache-root sync fails,
`quarantinePath` renames the quarantined entry back to the live slot and returns
without syncing the parent directory again.

The visible bytes are restored, but the compensating rename is not proven
durable. A second sync failure therefore cannot be injected or reported as
`StateChangedError`, and the install layer can still receive an ordinary error
and leave `BuildCacheRetained` false while durable cache state is uncertain.
This conflicts with the cycle-5 requirement to report secondary
compensation/sync failure truthfully.

## Test deliverable

Added
`TestAQuarantineRollbackThatCannotBeMadeDurableReportsChangedState` to
`internal/buildcache/compensation_test.go`.

The regression:

- injects the existing post-rename quarantine sync failure;
- requires a production-owned `quarantine-rollback-sync` boundary after the
  compensating rename;
- requires a failure at that boundary to surface as changed durable state;
- preserves and verifies the exact corrupt predecessor bytes;
- requires one live cache entry and no private staging residue.

Current result (expected red, real exit code):

```text
go test ./internal/buildcache -run '^TestAQuarantineRollbackThatCannotBeMadeDurableReportsChangedState$' -count=1
exit 1
--- FAIL: TestAQuarantineRollbackThatCannotBeMadeDurableReportsChangedState
    compensation_test.go:370: the compensating rename was not followed by a faultable durability sync
```

## Focused validation

Every command below ran directly as a standalone process.

| Command | Exit | Result |
| --- | ---: | --- |
| `go test ./internal/buildcache -run '^TestAFailedPublicationRestoresAPredecessorMovedBeforeQuarantineError$' -count=1` | 0 | cycle-4 publication regression green |
| `go test ./internal/buildcache -run '^TestAFailedReversalReturnsTheWinnerMovedBeforeWithdrawError$' -count=1` | 0 | cycle-4 reversal regression green |
| `go test ./internal/buildcache -run '^TestAQuarantineThatCannotBeMadeDurablePutsTheEntryBack$' -count=1` | 0 | first new quarantine test green |
| `go test ./internal/buildcache -run '^TestADurabilityFaultInsideAQuarantineIsCompensatedByItsCaller$' -count=1` | 0 | second new quarantine test green |
| `go test ./internal/buildcache -run '^TestAQuarantineThatCannotPutTheEntryBackHandsItsCallerTheRecord$' -count=1` | 0 | third new quarantine test green |
| `go test ./internal/install -run '^TestAPublicationThatChangedTheCacheIsReportedAsRetained$' -count=1` | 0 | typed publication change maps to retained state |
| `go test ./internal/install -run '^TestAReversalThatDidNotCompleteIsReportedAsRetained$' -count=1` | 0 | typed reversal failure maps to retained state |
| `gofmt -l internal/buildcache/publish.go internal/buildcache/compensation_test.go` | 0 | empty output |
| `git diff --check` | 0 | no tracked whitespace errors; the task's buildcache files are untracked against the base, so gofmt is the formatting evidence for them |
| `go build ./internal/buildcache ./internal/install ./cmd/curator` | 0 | affected packages build |
| `go vet ./internal/buildcache ./internal/install ./cmd/curator` | 0 | affected packages vet clean |

The named CLI install/upgrade E2E and broader suites were not rerun after the
new deterministic regression proved the mutation boundary incomplete. This
follows the focused-only directive and avoids treating unrelated green work as
evidence that the red durability contract is satisfied.

## Required rework

After moving `quarantinePath`'s entry back to the live slot, sync the parent
directory through a deterministic production-owned fault boundary. If that
secondary sync fails, preserve all recoverable bytes and propagate a typed
changed-state result so `runCommit` sets `BuildCacheRetained` and bounded,
redacted diagnostics describe the actual state. The new regression must turn
green without weakening its exact-byte and residue assertions.
