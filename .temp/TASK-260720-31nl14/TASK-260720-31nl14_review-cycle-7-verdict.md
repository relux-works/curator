# TASK-260720-31nl14 review cycle 7 verdict

## Verdict

Changes requested. Route to `to-dev`.

The cycle-6 implementation now recovers ordinary crash-created partial file and directory staging, and its existing replacement/addition corruption cases pass. One P1 concurrent-state preservation gap remains: the durable journal records only the active manifest entry, not the exact durable byte progress within that file.

## Finding

1. P1 — preparing recovery deletes concurrently truncated partial staging when the replacement remains a valid source prefix.

   `copyStagingFile` syncs every written chunk and then exposes `PointDuringStagingCopy` (`internal/transaction/staging.go:115-130`), but it does not durably advance a byte offset or prefix digest in the journal. On restart, `validateActiveStagingEntry` accepts any file whose bytes are any prefix of the recorded source (`staging.go:313-335`). `discardPreparing` then captures that current prefix as removal ownership and durably deletes it (`staging.go:163-190`).

   A subprocess probe crashed immediately after the first 32 KiB chunk had successfully synced, durably truncated the sidecar to a different 16 KiB source prefix, and invoked recovery. Five repeated runs produced the same result; a final run after explicitly syncing the truncation reported:

   ```text
   partial_before=32768 truncated_to=16384 recover_error=<nil> staged_exists=false journal_exists=false live="old" live_read_error=<nil>
   ```

   The 16 KiB state cannot be produced by the recorded execution after the successful 32 KiB sync boundary; it is unknown concurrent state. Recovery nevertheless returns success and deletes both the changed bytes and the journal. The probe is at `.temp/TASK-260720-31nl14/worktree/.temp/reviewer_cycle7_probe/main.go` and runs with `go run ./.temp/reviewer_cycle7_probe` from the task worktree.

This violates the acceptance criterion to refuse overwriting unknown concurrent state and the reviewer DoD requiring reverse/restart recovery to preserve it. It is ordinary implementation rework, not an external or human-only blocker.

## Required rework

- Make the durable preparation record identify exact acknowledged file-copy progress (for example, a durable byte count plus prefix digest after each synced chunk), or use an equivalent staging design that cannot confuse a later valid-prefix mutation with crash-created state.
- Recovery may remove only the exact last durably recorded partial state. A shorter, longer, replaced, or otherwise different prefix must return `ErrImplementationCorruption` while preserving the staged bytes and journal.
- Add a subprocess regression that crashes after at least one durable chunk, mutates the partial file to a different source-valid prefix, syncs that mutation, and proves recovery preserves it. Include shorter-prefix/truncation coverage and retain the current foreign-replacement and directory-addition cases.
- Preserve the cycle-5 parent-directory durability fix, cycle-6 ordinary partial-file/directory recovery, all earlier cleanup/namespace/rollback behavior, and platform portability.

## Independent validation

Passed on native Darwin/arm64 from the review source state:

- `go test ./internal/transaction -count=1`
- `go test -race ./internal/transaction -count=1`
- `go vet ./internal/transaction`
- repository-wide `make check` using the established task-local module overlay
- repository-wide `go test -race ./... -count=1` with the same overlay
- Linux/amd64 and Windows/amd64 complete compile graphs with `go test -exec=true ./...`
- `golangci-lint` v2.4.0 scoped to `internal/transaction`: 0 issues
- `gofmt -l internal/transaction`, `git diff --check`, and staged-file checks: clean

No product code was modified or staged during review. The only local review helper is under ignored task-scoped `.temp` state.

Native Windows runtime remains unavailable on this Darwin host. Cross-compilation is not runtime evidence, and the inherited `TASK-260720-1zl1cj` native Windows qualification gate remains unchanged.
