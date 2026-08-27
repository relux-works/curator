# TASK-260720-1ljev5 — cycle-3 gate manifest

Every command was run standalone, not through a pipe; each exit code below is
the real status of that process and is also stored beside its log in a matching
`.exit` file.

Tree: `.temp/TASK-260720-1ljev5/worktree` (uncommitted, unstaged).
Host: darwin/arm64, Go 1.25.5, golangci-lint 2.4.0.
Native Windows: host `win`, Windows 10 19045, non-elevated `schtasks /rl LIMITED`.

## macOS, primary

| Log | Command | Exit |
|---|---|---|
| `gofmt.log` | `gofmt -l .` | 0 (empty output) |
| `diffcheck.log` | `git diff --check` | 0 |
| `build.log` | `go build ./...` | 0 |
| `vet.log` | `go vet ./...` | 0 |
| `test-all.log` | `go test ./... -count=1` (40 packages) | 0 |
| `race-scoped.log` | `go test -race ./internal/scopes ./internal/buildcache ./cmd/curator` | 0 |
| `race-install.log` | `go test -race ./internal/install -run 'TestPostCommitMaintenance\|TestMaintenanceFailureAfterCommitIsAWarning\|TestConcurrent'` | 0 |
| `lint.log` | `golangci-lint run ./...` | 0 (0 issues) |

## Cross-platform

| Log | Command | Exit |
|---|---|---|
| `crossbuild-windows.log` | `GOOS=windows GOARCH=amd64 go build ./...` | 0 |
| `crossbuild-linux.log` | `GOOS=linux GOARCH=amd64 go build ./...` | 0 |
| `crossvet-linux.log` | `GOOS=linux GOARCH=amd64 go vet ./...` | 0 |
| `crossvet-windows.log` | `GOOS=windows GOARCH=amd64 go vet ./...` | **1 — expected red** |
| `crossvet-windows-scoped.log` | same, excluding `internal/runtimestore` | 0 |
| `crossvet-windows-base.log` | same full command on the accepted base tree | **1 — identical message** |

The expected red is `internal/runtimestore/targets_windows_test.go:97:14:
undefined: decodeHelperOutput`: a pre-existing, sibling-owned defect
(`TASK-260720-29hi1h`) that the untouched accepted base reproduces identically.
It is reported as failing, not as passing.

## Negative controls (expected red by construction)

| Log | Control | Exit |
|---|---|---|
| `negative-control-registry.log` | struct decoding restored in `parseConsumers` → the registry regressions | **1** |
| `negative-control-classification.log` | `store.inspectEntry(entryPath, …)` restored in the sweep → the classification-swap test | **1** |

`negative-control-classification.log` fails with *"an entry was retired on
someone else's proof"*, which is the cycle-2 finding reproduced deterministically.

## Native Windows

| Log | Command | Exit |
|---|---|---|
| `win/win-scopes.log` | `scopes.test.exe -test.count=1 -test.v` | 0 |
| `win/win-curator-gc.log` | `curator.test.exe -test.run TestGC` | 0 |
| `win/win-buildcache.log` | `buildcache.test.exe -test.count=1 -test.v` | **1 — inherited failures only** |
| `win/win-mine-race.log` | `buildcache.test.exe -test.count=5 -test.run TestAtomicPublication`, this tree | **1 — inherited** |
| `win/win-base-race.log` | same, accepted-base binary | **1 — inherited** |

Every sweep test passes natively, including
`TestSweepClassificationSurvivesACacheRootExchangedMidPass` and
`TestSweepRemovalSurvivesACacheRootExchangedMidPass` (the OS refused both
exchanges), and every registry regression in `internal/scopes`.

Inherited `internal/buildcache` failures, not task-owned:

- `TestAtomicPublicationIdenticalRace` and `TestAtomicPublicationConflictingRace`
  — the `ensureProtectedBase` DACL-inheritance race owned by
  `TASK-260720-3pwg2w`. Five like-for-like runs: this tree 3/2 and 0/5, accepted
  base 1/4 and 0/5, identical failure message on both.
- `TestWindowsProtectedStateMatrix/artifact_has_inherit-only_owner_allow` —
  fails on the accepted base as well.

Skips, honestly reported and none of them task-owned:
`TestWindowsProtectedStateMatrix/artifact_reparse_point` (needs a **file**
reparse point, i.e. `SeCreateSymbolicLinkPrivilege`),
`TestCandidateCacheOutcomeVocabulary` (needs `CURATOR_CONFORMANCE_ROOT`), and
the two `unreadable directory` scopes subtests (`chmod 0000` has no POSIX
meaning on Windows).
