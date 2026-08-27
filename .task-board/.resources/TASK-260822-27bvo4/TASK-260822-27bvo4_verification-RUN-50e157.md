# TASK-260822-27bvo4 — independent verification and handoff (RUN-260822-50e157)

Third spawn on this task. The two earlier runs (`RUN-260822-1e48d3`, `RUN-260822-507162`)
converged on the same fix and `RUN-260822-507162` merged them; that run then ended on
`exit=124` (timeout) before running the role handoff, so the item stayed in `development`
with the work uncommitted. This run adds no product code. It re-verifies the merged tree
from scratch, adds the darwin mutation check the earlier runs did not have, and hands off.

Tree under verification: `.temp/TASK-260822-27bvo4/RUN-260822-507162`,
branch `task/TASK-260822-27bvo4-symlink-launcher-507162`, base `6a9b201`.
Uncommitted per standing orders.

## Change under verification (unchanged by this run)

`internal/godriver/identity.go:56` — `resolveExecutableIdentity` canonicalizes with
`physicalPath(absolute)` instead of `filepath.EvalSymlinks(filepath.Clean(absolute))`.
Everything below it — regular-file check, size bound, link-count/reparse check,
`os.SameFile` re-check across the open, SHA-256 over exactly `Size()` bytes — is untouched
and now applies to the physical file.

Tests: `internal/godriver/identity_test.go` (new, 5 tests),
`internal/godriver/identity_windows_test.go` (new, 2 junction tests),
`main_test.go` (test-only `identityProbeMode`), `worker_test.go`
(`startRawWorkerFrom` + `workerScenario.launchPath`).

Full diff attached as `TASK-260822-27bvo4_symlink-launcher.patch`; re-diffed at the end of
this run and byte-identical to the working tree, so the artifact is faithful rather than asserted.

## Gates run by this run (real exit codes, each a standalone process, no pipes)

| Command | Exit |
|---|---|
| `gofmt -l ./cmd ./internal` | 0 (no output) |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `go test ./internal/godriver/ -count=1` | 0 (61.3s) |
| `go test ./internal/godriver/ -count=1 -run '<identity set>' -v` | 0, 6 PASS, 0 SKIP |
| `go test ./internal/godriver/ -count=1` (re-run after mutate/restore) | 0 (108.6s) |
| `golangci-lint run` | 0, `0 issues.` |
| `GOOS=windows go vet ./internal/godriver/` | 0 |
| `GOOS=linux go vet ./internal/godriver/` | 0 |

The six identity tests all RAN on darwin/arm64 — none skipped:

    --- PASS: TestExecutableIdentityRejectsSubstitutionAndMutation (0.11s)
    --- PASS: TestExecutableIdentityResolvesALauncherLink (0.25s)
    --- PASS: TestExecutableIdentityRejectsSubstitutionThroughALauncherLink (0.49s)
    --- PASS: TestExecutableIdentityResolvesARealSymlinkedProcessLaunch (0.03s)
    --- PASS: TestBuildAcceptsAManagerStartedThroughALauncherLink (1.93s)
    --- PASS: TestWorkerProvesTheInstalledIdentityWhenLaunchedThroughALink (0.90s)

## NEW: darwin mutation check (this run's addition)

Earlier runs proved red-to-green only on Windows, where the `EvalSymlinks` -> `physicalPath`
difference lives. On unix the two functions are the same code (`platform_unix.go:16`), so that
mutation is a no-op here and says nothing about whether the darwin tests are load-bearing.
The mutation that *is* meaningful on darwin is removing canonicalization altogether:

    -	canonical, err := physicalPath(absolute)
    -	if err != nil { ... }
    +	canonical := filepath.Clean(absolute)

Result — `go test ./internal/godriver/ -run '<identity set>' -v` exit **1**, all six red:

    --- FAIL: TestExecutableIdentityRejectsSubstitutionAndMutation (0.04s)
    --- FAIL: TestExecutableIdentityResolvesALauncherLink (0.04s)
    --- FAIL: TestExecutableIdentityRejectsSubstitutionThroughALauncherLink (0.13s)
    --- FAIL: TestExecutableIdentityResolvesARealSymlinkedProcessLaunch (0.03s)
    --- FAIL: TestBuildAcceptsAManagerStartedThroughALauncherLink (1.77s)
    --- FAIL: TestWorkerProvesTheInstalledIdentityWhenLaunchedThroughALink (0.50s)

Log: `TASK-260822-27bvo4_darwin-mutant-no-canonicalization.log`.

FINDING worth keeping: the mutant reddens even the pre-existing
`TestExecutableIdentityRejectsSubstitutionAndMutation`, which uses no link at all. Reason is
in the failure text — on macOS `t.TempDir()` hands back `/var/folders/...`, and `/var` is
itself a symlink to `/private/var`:

    identity = {Path:/var/folders/.../opt/curator ...},
    want the installed file {Path:/private/var/folders/.../curator ...}
    the manager executable is not a canonical regular file

So on macOS canonicalization is not only about the operator's launch shape. Without it the
manager cannot identify itself through an ordinary absolute path whose *parent* is a link,
which is the default shape of `/var`, `/tmp`, and any `/opt`-style install published behind
one. That is the same class of defect as the Windows junction, reachable without any package
manager involved.

The tree was restored from a byte backup after each mutation attempt and re-verified: `git diff`
returns the identical 3-file/70-insertion diff, and the attached patch re-diffs identical.
Product code was never left mutated at the end of any step.

## AC mapping

- "Test proves symlinked launch resolves" — `TestExecutableIdentityResolvesALauncherLink`
  (four launch shapes resolve to one identity), `TestExecutableIdentityResolvesARealSymlinkedProcessLaunch`
  (a real re-exec through a shim link, reading back what the started process resolved for itself),
  `TestBuildAcceptsAManagerStartedThroughALauncherLink` (full Build handshake with
  `managerExecutable` pointed at a link), `TestWorkerProvesTheInstalledIdentityWhenLaunchedThroughALink`
  (the worker process itself started from a link). All PASS, none skipped.
- "substitution controls still fail closed" — `TestExecutableIdentityRejectsSubstitutionThroughALauncherLink`
  and `TestExecutableIdentityRejectsSubstitutionAndMutation` PASS: dangling link, directory,
  empty file, link loop, hard link, retargeted shim, mutated bytes, and the physical file swapped
  for a link after resolution all stay refused with `build_execution_worker_identity_invalid`.
  Windows-only equivalents behind a junction are in `identity_windows_test.go`.
- "go test green" — `go test ./internal/godriver/ -count=1` exit 0, twice.

## Not run, and why

- `go test ./...` (whole module). NOT RUN by this run. The host data volume was at 100%
  (1.0 GiB free) for most of this run while two concurrent spawns held full-suite runs, and
  a concurrent process was purging the shared `GOCACHE` mid-link. Three mutation attempts
  died on that, not on any test: `link: mapping output file failed: no space left on device`,
  then twice `could not import io (open <GOCACHE>/dd/dd72...-d: no such file or directory)`.
  The mutation check only completed once it was moved onto a private per-run `GOCACHE`
  (337 MB, deleted afterwards). Same host condition `RUN-260822-507162` recorded.
- Windows execution of `identity_windows_test.go`. NOT RUN by this run — no Windows host
  reached from here. It was run by `RUN-260822-507162` on a real native checkout
  (`C:\Users\admin\TASK-260822-27bvo4-507162`, go1.25.5 windows/amd64): merged tree
  `go test ./internal/godriver/ -count=1` exit 0 (93.7s), and the same tree with `identity.go`
  reverted to `filepath.EvalSymlinks` exits 1 on the junction test. Evidence attached as
  `TASK-260822-27bvo4_windows-identity-tests.log` and
  `TASK-260822-27bvo4_windows-prefix-baseline-red.log`. This run compiled the Windows lane
  instead: `GOOS=windows go vet ./internal/godriver/` exit 0.
- Conformance gates (`make ci-test`, `candidate-*`). NOT RUN — `CURATOR_CONFORMANCE_ROOT`
  is unset on this host and the suites self-skip.
- `make ledger-check`. NOT RUN — it is red at HEAD on rows naming packages absent from the
  module, recorded by `TASK-260822-5wfdfx`; unrelated to this change.

## Where the work is

Uncommitted in `.temp/TASK-260822-27bvo4/RUN-260822-507162` on branch
`task/TASK-260822-27bvo4-symlink-launcher-507162` (base `6a9b201`).
The superseded first-lane worktree `.temp/TASK-260822-27bvo4/worktree`
(branch `task/TASK-260822-27bvo4-symlink-launcher`) is a strict subset and was left untouched.

NOTE FOR THE REVIEWER on repo geometry: `internal/godriver` does not exist on the main
checkout's current branch (`handoff/cocoaskills-parity-20260731` at `74fe162`) — that lineage
predates the build-driver packages. It exists on `main` (`0d9fe6e`), which still carries the
unfixed `filepath.EvalSymlinks` at `internal/godriver/identity.go`, so the fix has not landed
anywhere yet. `6a9b201` is one commit behind `main` and that commit is board metadata only,
so the branch applies to `main` cleanly.
