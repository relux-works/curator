# TASK-260720-1zl1cj review rework cycle 3

## Review finding resolved

The cycle-2 defect was an unstable first-use manager-home identity on case-insensitive Windows filesystems. Managers constructed from case-variant spellings while the home was absent could retain different identity strings, hash different home-lock filenames, and split process-local ordering state.

The canonicalization seam is now platform-specific:

- Unix preserves the canonical existing prefix and missing component spelling.
- Windows queries `FileCaseSensitiveInfo` on the longest existing directory prefix.
- Ordinary case-insensitive Windows paths normalize the complete canonical identity, including nonexistent suffix components, to one uppercase form.
- Windows directories with per-directory case sensitivity enabled preserve exact spelling, so distinct case-sensitive paths are not conflated.
- Windows/filesystem versions that do not expose per-directory case-sensitivity information retain traditional case-insensitive normalization.

The Windows-only subprocess regression constructs two case-variant managers before home creation and proves one `Home`, one lock path, and one process-order state. It then proves build-key/home nesting is rejected, a case-variant child blocks while the first home lock is held, the child acquires after release, and identity stays stable after first-use state creation. The earlier symlinked-existing-prefix regression remains intact.

## Scoped files

- `internal/managerlock/identity.go`
- `internal/managerlock/identity_unix.go`
- `internal/managerlock/identity_windows.go`
- `internal/managerlock/identity_windows_test.go`

No transaction journal, target swap, installer orchestration, recovery, or GC ownership was added in this rework.

## Validation evidence

- `go test -race -cover ./internal/managerlock -count=1 -v` passed on Darwin/arm64 with 82.4% statement coverage, including native Unix subprocess contention and abnormal-exit cases.
- `go vet ./internal/managerlock`, `gofmt`, and scoped diff checks passed.
- `make check` passed: repository-wide `go vet`, tests, and formatting.
- `go test -race ./... -count=1` passed uncached repository-wide.
- `make build` produced the native Curator CLI successfully.
- Linux/amd64 and Windows/amd64 managerlock test binaries compiled successfully; the Windows build includes the new Windows-only subprocess regression.
- Native Windows runtime execution was not available on this Darwin host and was not claimed.
- `golangci-lint` was not installed; the task-defined lint equivalents (`go vet` and `gofmt`) passed.
- `git diff --check` passed and the task worktree has no staged changes.

## Provenance

The isolated task worktree remains based at exact `origin/main` commit `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`, with the previously imported reviewer-accepted product diff preserved. This cycle changed only the managerlock identity seam and its Windows regression named above.
