# TASK-260720-31nl14 implementation results

## Review cycle 1 rework

- Directory digests now encode the root directory entry with the same permission metadata as descendant directories and verify that its identity and permissions remain stable while hashing.
- Rollback and cleanup regression tests change committed root-directory permissions and prove that implementation-corruption is returned without overwriting current bytes or deleting the retained backup/journal.
- Plan and journal validation now require deterministic sidecar names and independent target namespaces before the first journal write. It rejects ancestor/descendant targets, symlink aliases, hard-link aliases, self-collisions, and every cross-target live/staging/backup/rollback collision.
- Native-Windows validation treats filesystem components case-insensitively for collision safety and has a Windows-only case-alias regression test. The complete Windows test graph compiles; native Windows runtime execution remains unavailable and is not claimed.

## Provenance and scope

- Worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-31nl14/worktree`
- Exact base: `origin/main` commit `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`
- Imported prerequisite: complete product state from `TASK-260720-1zl1cj`, excluding `task-board.config.json` and blocked `TASK-260720-1zntv0`-only `internal/godriver`.
- Candidate/imported product manifest SHA-256: `4b34cdb44fed3cce5082f02ab3db7b385049f0ac49f547be40f4e5fe9161bb3a` on both trees after the required exclusions.
- Task-owned product scope: new `internal/transaction` package only. `install.Project`, `install.Global`, CLI orchestration, GC, and build execution were not refactored.
- No file was staged or committed. The generated native build binary was moved to Trash after validation.

## Implementation

- Canonical JSON journal under manager state records transaction/project identity, sorted class plus unsigned-byte identifier targets, expected generation or preimage digest, deterministic sibling staging/backup/rollback paths, desired and captured backup digests, referenced build keys, phase, and per-target state.
- Caller-held `AssertHeld` home-lock witness is mandatory for preparation, commit, recovery, and referenced-key discovery.
- State machine: `preparing -> prepared -> committing -> cleanup`, or `rolling_back -> rollback_cleanup`; per-target states are `pending -> backed_up -> committed` and `rolled_back`.
- Durable journal replacement uses synced temporary files and atomic replacement. POSIX uses file/directory `fsync`; Windows uses `FlushFileBuffers` where supported and `MoveFileEx(..., MOVEFILE_WRITE_THROUGH)` for namespace mutations.
- Native no-replace target renames use `renamex_np(RENAME_EXCL)` on Darwin, `renameat2(RENAME_NOREPLACE)` on Linux, and non-replacing `MoveFileEx` on Windows. Unknown concurrent bytes are never overwritten.
- Backup capture is re-digested after the atomic rename. Rollback atomically captures the current desired target, rechecks its digest, restores in exact reverse order, and recognizes every tested crash window, including a restored preimage whose state write was interrupted.
- Cleanup is itself a durable phase; backups remain until all desired consumers are synced and revalidated. Journal deletion is last.
- Recovery scans transaction IDs in unsigned byte order without a current-project filter and deterministically resumes preparation rollback, commit, rollback, or cleanup.
- `ReferencedBuildKeys` returns the sorted union from every live journal for GC retention.
- Directory desired digests bind root permissions, preventing rollback or cleanup from accepting a root-metadata mutation as the recorded desired state.
- Target namespace validation resolves existing symlink prefixes, compares existing file identity, detects ancestor relationships, and conservatively folds native-Windows component case before any journal is created.

## Tests added

- Exhaustive fault injection at every target/backup/install boundary across three ordered targets, with exact reverse rollback assertions and untouched-target checks.
- Desired-digest mismatch and broader unknown-state corruption tests proving current bytes are preserved.
- Subprocess crashes after backup and after desired installation, followed by new-manager home-lock recovery.
- Restart recovery for multiple transaction IDs, interrupted rollback, preparing, and cleanup phases.
- File, directory, creation, removal, expected-generation, deterministic ordering, canonical journal, referenced-key, lock-witness, unsafe-path, no-replace, tomb cleanup, and race coverage.
- Root-directory metadata mismatch coverage for both rollback and cleanup, plus file/directory alias, nested namespace, self-sidecar, cross-target sidecar, and native-Windows case-alias plan validation.

## Verification

The exact-base worktree intentionally kept the tracked `skill-go-testing-tools` submodule unmaterialized. Repository-wide commands used the task-local `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-31nl14/go.test.mod`, whose only semantic difference is an absolute replace to the already checked-out canonical submodule at the same tracked commit.

- `go test ./internal/transaction -count=1`: pass.
- `go test -race ./internal/transaction -count=1`: pass.
- `go test -cover ./internal/transaction -count=1`: pass, 77.7% statement coverage after review-cycle regression coverage.
- `go vet ./internal/transaction`: pass.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.4.0 run ./internal/transaction/...`: 0 issues.
- `make check`: pass across the complete imported product graph.
- `go test -race ./... -count=1`: pass across the complete imported product graph.
- `make build`: pass on native Darwin/arm64.
- `GOOS=linux GOARCH=amd64 go test -exec=true ./...`: complete Linux compile graph passes.
- `GOOS=windows GOARCH=amd64 go test -exec=true ./...`: complete Windows compile graph passes.
- `gofmt -l internal/transaction`, `git diff --check`, and staged-file check: clean.

Native Darwin subprocess/runtime evidence is present. Native Windows runtime execution is unavailable on this Darwin host; Windows compilation is not represented as runtime evidence. The reviewer-confirmed `TASK-260720-1zl1cj` native Windows runtime gate remains preserved.

## Evidence logs

- `make-check-final.log`
- `go-test-race-all-final.log`
- `make-build-final.log`
- `go-test-compile-linux-amd64-final.log`
- `go-test-compile-windows-amd64-final.log`
- `go-test-cover-transaction-final.log`
- `golangci-lint-transaction-04.log`
- `provenance-02.log`
- `make-check-rework.log`
- `go-test-race-all-rework.log`
- `make-build-rework.log`
- `go-test-compile-linux-amd64-rework.log`
- `go-test-compile-windows-amd64-rework.log`
- `go-test-cover-transaction-rework.log`
- `golangci-lint-transaction-rework.log`
- `task-board-validate-rework.log` (13 inherited board-wide issues; none belongs to this task)
