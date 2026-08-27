# TASK-260720-1ljev5 — gate exit codes

Every command below was run standalone, not through a pipe, and the exit code
recorded is the command's own status. Logs in this directory are the captured
output of those runs.

Worktree: `.temp/TASK-260720-1ljev5/worktree` (base `origin/main` 17804ce,
composed from the accepted `TASK-260720-2284br` tree at
`.temp/TASK-260720-1zntv0/worktree`).
Host: darwin/arm64, Go 1.25.5, `golangci-lint` 2.4.0 (`/Users/iv/go/1.25.5/bin`).

## Final gates (after every source and test edit)

| Command | Log | Exit |
|---|---|---|
| `gofmt -l .` | `gate-gofmt-final.log` | 0 (empty) |
| `git diff --check` | `gate-diffcheck-final.log` | 0 |
| `go build ./...` | `gate-build-final.log` | 0 |
| `go vet ./...` | `gate-vet-final.log` | 0 |
| `golangci-lint run ./...` | `gate-lint-final.log` | 0 — `0 issues.` |
| `go test ./... -count=1` | `gate-test-all-final.log` | see `gate-test-all-final.exit` |

## Earlier gates on identical product sources

Only test files and `README.md` changed between these and the final gates.

| Command | Log | Exit |
|---|---|---|
| `gofmt -l .` | `gate-gofmt.log` | 0 |
| `git diff --check` | `gate-diffcheck.log` | 0 |
| `go build ./...` | `gate-build.log` | 0 |
| `go vet ./...` | `gate-vet.log` | 0 |
| `go test ./... -count=1` (38 packages) | `gate-test-all.log` | 0 |
| `go test -race ./internal/scopes ./internal/buildcache ./cmd/curator` | `gate-race-scoped.log` | 0 |
| `go test -race ./internal/install -run 'TestPostCommitMaintenance\|TestMaintenanceFailureAfterCommitIsAWarning\|TestConcurrent'` | `gate-race-install.log` | 0 (62.4s) |
| `golangci-lint run ./...` | `gate-lint.log` | 0 — `0 issues.` |

## Cross-platform

| Command | Log | Exit |
|---|---|---|
| `GOOS=windows GOARCH=amd64 go build ./...` | `gate-crossbuild-windows.log` | 0 |
| `GOOS=linux GOARCH=amd64 go build ./...` | `gate-crossbuild-linux.log` | 0 |
| `GOOS=linux GOARCH=amd64 go vet ./...` | `gate-crossvet-linux.log` | 0 |
| `GOOS=windows GOARCH=amd64 go vet ./...` | `gate-crossvet-windows.log` | **1 — expected red, see below** |
| `GOOS=windows GOARCH=amd64 go vet $(go list ./... \| grep -v runtimestore)` | `gate-crossvet-windows-scoped.log` | 0 |

### The expected-red gate

`GOOS=windows go vet ./...` exits **1**:

```
vet: internal/runtimestore/targets_windows_test.go:97:14: undefined: decodeHelperOutput
```

Pre-existing and sibling-owned (`internal/runtimestore`, `TASK-260720-29hi1h`).
The identical command on the untouched accepted base
(`.temp/TASK-260720-1zntv0/worktree`) also exits **1** with the identical
message. This task does not touch that package; excluding it, the same gate
exits 0 over this tree.

## Native Windows (host `win`, cross-compiled binaries, non-elevated)

Windows has no Go toolchain, so test binaries were cross-compiled on macOS and
run natively through a temporary `schtasks /rl LIMITED` task (deleted after the
run). A **non-elevated** session is required: an elevated one makes Windows
assign `BUILTIN\Administrators` as owner of every created object, which
`buildcache.validateWindowsOwner` correctly rejects.

| Binary | Log | Exit |
|---|---|---|
| `internal/scopes` | `win-scopes.log` | 0 |
| `cmd/curator` (`-run TestGC`) | — | 0, all three serialization tests pass |
| `internal/buildcache` (this tree) | `win-buildcache.log` | 1 — see below |
| `internal/buildcache` (accepted base) | `win-base-buildcache.log` | 1 — same failures |
| `cmd/curator` (full) | `win-curator.log` | 1 — see below |

**Every new sweep test passes natively on Windows.** The three
`internal/buildcache` failures are `TestAtomicPublicationIdenticalRace`,
`TestAtomicPublicationConflictingRace`, and
`TestWindowsProtectedStateMatrix/artifact_has_inherit-only_owner_allow` — all
owned by `TASK-260720-3pwg2w`, and all reproduced by the accepted base binary
run in the same session. Five repeated `-test.run TestAtomicPublication` runs
exit 1 on base and on this tree alike.

The only `cmd/curator` failure is
`TestCLIEndToEndInstallStatusAndTamperCheck`: `git` is not installed on the host
(`where git` finds nothing). Pre-existing, untouched by this task.

`internal/managerlock` was also run and exits 1, with five subprocess tests
reporting `exec: "managerlock.test.exe": cannot run executable found relative to
current directory` — a standalone cross-compiled test binary cannot re-exec
itself. Harness artifact; package untouched by this task.

The three `internal/scopes` integration tests skip on Windows because building a
manager-protected home from outside `internal/buildcache` requires the
unexported `protectWindowsPath`. Those code paths do run natively on Windows
through the `internal/buildcache` sweep tests.
