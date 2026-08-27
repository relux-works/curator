# TASK-260720-31nl14 review cycle 7 rework validation

## Resolution

The remaining preparing-recovery concurrent-state gap is closed within `internal/transaction`.

- Every successfully synced staging-file chunk now advances a canonical durable byte count and SHA-256 prefix digest in the journal before `PointDuringStagingCopy` is exposed.
- Preparing recovery accepts only the exact acknowledged file size, prefix digest, source prefix, mode, and unchanged full source manifest entry.
- A shorter prefix, longer source-valid prefix, same-size changed payload, foreign replacement, or directory addition now returns `ErrImplementationCorruption` while preserving the staged bytes, live preimage, and journal.
- Preparation completion and discard paths clear prefix progress atomically with their existing durable journal transitions.
- Journal validation rejects missing, absent, malformed, negative, inactive, or uncreated prefix-progress combinations.

The reviewer's cycle-7 probe now reports implementation-corruption and preserves its synced 16 KiB truncation:

```text
partial_before=32768 truncated_to=16384 recover_error=recover transaction txn-prefix-truncation: transaction implementation-corruption: partial staging size changed from durable progress ... staged_exists=true journal_exists=true live="old" live_read_error=<nil>
```

No install consumer, CLI, GC, build execution, or manager-lock code was refactored in this cycle.

## Regression coverage

- Subprocess death after the first durable 32 KiB staging chunk proves the journal contains the exact acknowledged count and digest.
- Synced truncation to a 16 KiB source-valid prefix is preserved and rejected.
- Synced extension to a 48 KiB source-valid prefix is preserved and rejected.
- Synced same-size changed bytes are preserved and rejected.
- Existing ordinary partial-file/directory recovery, foreign replacement, and directory-addition cases remain green.
- The new and retained preparing crash/concurrent-state cases passed 20 consecutive iterations; the new exact-prefix cases separately passed 10 consecutive iterations after the same-size regression was added.

## Provenance and scope

- Worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-31nl14/worktree`
- Exact base: `origin/main` commit `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`
- Imported manager-lock candidate provenance remains unchanged.
- Cycle-7 files: `internal/transaction/types.go`, `staging.go`, `engine.go`, `journal.go`, `subprocess_test.go`, and `validation_test.go`.
- No file was staged or committed.

## Verification

Passed on native Darwin/arm64:

- `go test ./internal/transaction -count=1`
- `go test -race ./internal/transaction -count=1`
- focused preparation regressions with `-count=20`
- exact-prefix plus journal-validation regressions with `-count=10`
- `go vet ./internal/transaction`
- focused statement coverage: 74.8 percent
- reviewer's cycle-7 standalone truncation probe
- `golangci-lint` v2.4.0 scoped to `internal/transaction`: 0 issues
- repository-wide `make check` using the established task-local module overlay
- repository-wide `go test -race ./... -count=1` with the same overlay
- native `make build` and `curator --version` runtime smoke (`curator dev`)
- complete Linux/amd64 and Windows/amd64 compile graphs using `go test -exec=true ./...`
- `gofmt -l internal/transaction`, `git diff --check`, and staged-file checks: clean

The exact-base worktree intentionally leaves the tracked testing-tool submodule unmaterialized. Repository-wide commands use `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-31nl14/go.test.mod`, whose only semantic difference is the established absolute replacement to the canonical checkout of that same tracked module.

Native Windows runtime execution is unavailable on this Darwin host. Cross-compilation is not runtime evidence, and the inherited `TASK-260720-1zl1cj` native Windows qualification gate remains preserved.

`task-board validate` reports 14 inherited board-wide issues unrelated to this task: 12 legacy `EPIC-260712-*` broken dependency links and two orphan resources (`TASK-260713-7a9c1e/review.md` and `TASK-260713-c7a18d/research.md`). No board file was edited directly.
