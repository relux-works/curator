# TASK-260720-1nlmvv — cycle-2 continuation gate evidence

Continuation of `RUN-260728-5fd800` after `RUN-260729-39026a` was killed mid-gate.
No product code was changed in this continuation: the tree is byte-identical to the
one cycle 2 handed over.

Worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1nlmvv/worktree`
(base `origin/main` 17804ce plus the accepted `TASK-260720-1ljev5` tree; nothing
staged, committed, or published).
Host: darwin/arm64, Go 1.25.5, golangci-lint v2.4.0 (pinned).

## What this continuation had to close

The cycle-2 handoff left three race gates unreported. The first race sweep of that
run hit `ENOSPC`, the run then removed the shared Go build cache while concurrent
reviewers were active, restarted `internal/install` under `-race`, and was killed
before the result existed — `race.log` held only its header and no exit code.

## Freshness of the already-reported gates

No `.go` file under the worktree is newer than the cycle-2 `go test ./...` log:

```
find . -name '*.go' -newer gate-logs-cycle-2/go-test-all.log -not -path './.git/*'   ->  (no output)
```

Source last changed 04:02; `static.log` 04:09, `windows-native.log` 04:10,
`go-test-all.log` 04:20. Every static and functional gate therefore postdates the
last edit and stands. Removing a Go **build cache** does not invalidate a test run
that already completed — it only removes a compilation speedup — so no completed
gate was re-run on that account. Only the three gates that never produced a result
were run here.

## Race gates — this continuation

Run by `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1nlmvv/run-race-gates.sh`,
sequentially, each as a standalone process with its real exit code recorded. No
`tee`, no pipe chain, no shared-cache mutation. Raw per-gate logs and the summary
are in `gate-logs-cycle-2/`.

| Gate | Wall | Exit |
|---|---|---|
| `go test -race -timeout 30m -count=1 ./internal/godriver` | 46.5s | **0** |
| `go test -race -timeout 60m -count=1 ./internal/install` | 764.3s | **0** |
| `go test -race -timeout 60m -count=1 ./cmd/curator` | 651.9s | **0** |

Disk headroom was watched throughout (`df -g /System/Volumes/Data`):
15Gi at start, 14Gi at the end. No `ENOSPC`. No cache was cleared.

## Gates carried forward from cycle 2 (unchanged tree)

| Gate | Exit |
|---|---|
| `gofmt -l .` | 0 (no output) |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `golangci-lint run ./...` (pinned v2.4.0) | 0, **0 issues** |
| `go test ./... -count=1 -timeout 60m` (40 packages) | 0 |
| `GOOS=windows GOARCH=amd64 go build ./...` | 0 |
| `GOOS=linux GOARCH=amd64 go build ./...` | 0 |
| `GOOS=linux GOARCH=amd64 go vet ./...` | 0 |
| native Windows `install.test.exe` redaction suite | 0 |
| native Windows `curator.test.exe` classification suite | 0 |

### Expected-red gate, reported as failing

`GOOS=windows GOARCH=amd64 go vet ./...` exits **1**:

```
vet: internal/runtimestore/targets_windows_test.go:97:14: undefined: decodeHelperOutput
```

This is a real failure and is reported as one. It is pre-existing and
sibling-owned (`internal/runtimestore`, `TASK-260720-29hi1h`): the identical
command on the untouched accepted base `.temp/TASK-260720-1ljev5/worktree` exits
**1** with the identical message. Excluding that one package, this tree's Windows
vet exits 0.

### Not run

Native Linux execution. No Linux host with a Go toolchain is reachable
(`ssh lev` has no `go` on `PATH` and no approved toolchain identity). The change
is platform-neutral presentation and planning code; `GOOS=linux` build and vet are
clean, and the only platform-specific behaviour it touches — absolute-path
redaction in Unix, Windows-drive, and UNC form — passes on macOS and on native
Windows.

## Environment note

A concurrent agent was running `./cmd/curator` tests in
`.temp/TASK-260729-2kaopg/worktree` for the whole window. It shares the host Go
build cache but not the worktree; no gate above touched that tree, and no cache
was cleared while it ran.
