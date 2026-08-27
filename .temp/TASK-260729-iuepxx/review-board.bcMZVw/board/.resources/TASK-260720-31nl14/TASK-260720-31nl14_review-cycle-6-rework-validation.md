# TASK-260720-31nl14 review cycle 6 rework validation

## Resolution

The `PhasePreparing` crash-recovery gap is closed within `internal/transaction`.

- The canonical journal now records the absolute staged source, a deterministic entry manifest, and a durable per-target preparation cursor before creating or extending each staged filesystem entry.
- Completed entries are checked against their recorded kind, mode, and digest. Only the durably active regular file may be partial, and its current bytes must be an exact prefix of the unchanged recorded source.
- Preparation uses a new `during_staging_copy` fault boundary after a durable partial-file write.
- Restart recovery validates the complete partial tree before recording its exact removal digest and per-entry removal manifest. Unrecorded additions, changed metadata, replacement content, and unsafe entries return `ErrImplementationCorruption` without removing current bytes.
- Partial-sidecar removal records the discarded staging state in the same journal transition that clears removal ownership, preserving deterministic recovery across another crash.
- `PhasePrepared` atomically clears preparation-only source and progress metadata after every staged target and containing parent are durable.

No install consumer, CLI, GC, or build execution code was refactored.

## Regression coverage

- Subprocess death during a durable partial regular-file staging write recovers without mutating the live preimage.
- Subprocess death during a nested-directory file write recovers the partial tree and removes the journal.
- Concurrent replacement of the partial regular file produces implementation-corruption and preserves the replacement plus journal.
- Concurrent addition inside the partial directory produces implementation-corruption and preserves the added bytes plus journal.
- The prior preparing-recovery test now records canonical completed progress instead of bypassing the preparation protocol.
- The new crash/concurrent-state cases pass 20 consecutive iterations.

## Provenance and scope

- Worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-31nl14/worktree`
- Exact base: `origin/main` commit `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`
- Imported manager-lock candidate provenance remains unchanged.
- Review-cycle-6 files: `internal/transaction/types.go`, `engine.go`, `journal.go`, new `staging.go`, `subprocess_test.go`, and `validation_test.go`.
- No file was staged or committed.

## Verification

Passed on native Darwin/arm64 from the rework source state:

- `go test ./internal/transaction -count=1`
- `go test -race ./internal/transaction -count=1`
- `go vet ./internal/transaction`
- new subprocess and concurrent-state regressions repeated 20 times
- focused statement coverage: 74.6 percent
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.4.0 run ./internal/transaction/...`: 0 issues
- repository-wide `make check` with the established task-local module overlay
- repository-wide `go test -race ./... -count=1` with the same overlay
- native `make build` and `curator --version` runtime smoke (`curator dev`)
- complete Linux/amd64 and Windows/amd64 compile graphs before the final checked-close cleanup; final Linux/Windows transaction package compile checks after it
- `gofmt -l internal/transaction`, `git diff --check`, and staged-file checks: clean

The exact-base worktree intentionally leaves the tracked testing-tool submodule unmaterialized. Repository-wide commands use `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-31nl14/go.test.mod`, whose only semantic difference is the established absolute replacement to the canonical checkout of that same tracked module.

Native Windows runtime execution is unavailable on this Darwin host. Cross-compilation is not runtime evidence, and the inherited `TASK-260720-1zl1cj` native Windows qualification gate remains preserved.

`task-board validate` reports 14 inherited board-wide issues unrelated to this task: 12 legacy `EPIC-260712-*` broken dependency links and two orphan resources (`TASK-260713-7a9c1e/review.md` and `TASK-260713-c7a18d/research.md`). No board file was edited directly.
