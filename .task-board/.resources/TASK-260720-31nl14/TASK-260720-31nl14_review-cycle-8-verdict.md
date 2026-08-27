# TASK-260720-31nl14 review cycle 8 verdict

## Verdict

Changes requested. Route to `to-dev`.

## Finding

1. P1 - crash after a staging chunk sync but before its progress journal write is not recoverable.

`copyStagingFile` writes a chunk and calls `destination.Sync()` at `internal/transaction/staging.go:128-137`; only afterward does it update `StagingBytes` and `StagingPrefixDigest` and durably replace the journal at lines 138-143. A process or power loss in that interval leaves durable transaction-produced bytes beyond the last durable count. `validateActiveStagingEntry` requires the file size to equal the old count at lines 327-332, so `Recover` reports `ErrImplementationCorruption`, preserves the partial sidecar and journal, and cannot complete deterministic preparation rollback. The cycle-7 longer-prefix regression at `subprocess_test.go:223-225` deliberately proves that this exact state shape is rejected; it cannot distinguish concurrent extension from the engines own unjournaled next synced chunk. The exposed `PointDuringStagingCopy` occurs only after `saveJournal`, so current fault tests skip the vulnerable interval.

This violates interrupted-work recovery, restart determinism, and the durability-before-transition intent of the AC. It is ordinary implementation rework, not an external or human-only blocker.

## Required rework

- Introduce durable write-ahead ownership for each in-flight chunk before filesystem mutation, or use an equivalent staging mechanism with no ambiguous sync-before-journal state. Recovery must recognize every state the engine can leave at an arbitrary crash while still preserving bytes outside the durably authorized transition.
- Add a fault boundary after chunk file sync and before the progress journal becomes durable. Add subprocess regressions for first and later chunks showing restart recovery returns success, preserves the live preimage, and durably removes owned staging plus journal.
- Retain corruption coverage for shorter, longer, replaced, or changed bytes that are outside the exact durable write-ahead state, plus all earlier preparation, cleanup, rollback, namespace, and platform regressions.

## Independent validation

Passed on native Darwin arm64: focused tests; package race and vet; repository `make check`; repository full race; native build and version smoke; Linux amd64 and Windows amd64 full compile graphs via `go test -exec=true`; golangci-lint v2.4.0 with 0 issues; gofmt and diff checks; 74.8 percent focused statement coverage. No product code was modified or staged. Native Windows runtime remains unavailable; cross-compilation is not runtime evidence, and the inherited TASK-260720-1zl1cj qualification gate remains unchanged.