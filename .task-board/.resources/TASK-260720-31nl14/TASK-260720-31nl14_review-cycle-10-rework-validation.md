# TASK-260720-31nl14 review cycle 10 rework validation

## Resolution

Both native-Windows P1 findings from review cycle 10 are resolved within `internal/transaction`.

- Regular-file durability now dispatches through a platform primitive. Windows opens the existing regular file with `GENERIC_WRITE`, shared read/write/delete access, and `FILE_FLAG_OPEN_REPARSE_POINT`, then requires `FlushFileBuffers` and handle close to succeed. It does not suppress `ERROR_ACCESS_DENIED`. Unix retains its existing read-handle `File.Sync` behavior.
- The journal cleanup digest now uses the mode the platform can persist for the canonical writable journal. Unix remains `0600`; Windows uses Go's persisted writable regular-file mode `0666`, because Windows permission reporting maps a cleared read-only attribute to `0666`. Changed journal bytes still produce `ErrImplementationCorruption` and remain in place.
- Windows-only regressions directly exercise regular-file/tree syncing, canonical journal removal, actual persisted journal mode, and preservation of durably changed journal bytes.

The implementation follows the Windows API contract that `FlushFileBuffers` requires a handle with `GENERIC_WRITE`: <https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-flushfilebuffers>.

## Native Windows runtime evidence

- Runner: Windows 10.0.19045 AMD64 through SSH alias `win`.
- Toolchain used to cross-compile: Go 1.25.5 on Darwin arm64.
- Coverage-enabled Windows amd64 test binary size: 6,646,784 bytes.
- Local and remote binary SHA-256: `ee4844bbbc2ce64ac26783c8afc4202b221ae2f35126969388fb515ea841e3f0`.
- The two focused Windows regressions passed natively.
- The complete `internal/transaction` suite passed natively with no skips and 74.4 percent statement coverage. This includes boundary fault injection, exact reverse rollback, cleanup/restart recovery, preparation crash recovery, subprocess helpers, journal removal, and Windows case-alias handling.
- Full Windows log SHA-256: `ec0b5efe81c8d8dcd102e3b4e88c2fcc55323d9dad1a8520adc9291794877784`.
- Windows coverage profile SHA-256: `5594370b7f7a92e99cc0e09b6bad3f475242a7da11d544f690d3c55905c16f02`.
- Remote test binary, log, coverage profile, and earlier smoke binary were removed; follow-up checks reported every path absent.

## Verification

Passed:

- `go test ./internal/transaction -count=1`
- `go test -race ./internal/transaction -count=1`
- `go vet ./internal/transaction`
- focused Darwin statement coverage: 75.0 percent
- native Windows focused regressions and complete coverage-enabled package suite
- repository-wide `make check` with the established task-local module overlay
- repository-wide `go test -race ./... -count=1` with the same overlay
- scoped golangci-lint v2.4.0: 0 issues
- native Darwin `make build` and `curator --version` runtime smoke (`curator dev`)
- complete Linux amd64 and Windows amd64 compile graphs using `go test -exec=true ./...`
- `gofmt -l internal/transaction`, `git diff --check`, and staged-file checks: clean

## Provenance and scope

- Worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-31nl14/worktree`
- Exact detached base: `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`
- Rework files: `internal/transaction/files.go`, `durability_unix.go`, `durability_windows.go`, `journal.go`, and `validation_windows_test.go`.
- No install consumer, CLI, GC, build execution, or manager-lock product code changed.
- No file was staged or committed.

The accepted manager-lock dependency has since received two Windows-only portable test-expectation edits. Its product implementation in this exact-base transaction worktree is unchanged; this task did not import those ownership-external test-only edits.
