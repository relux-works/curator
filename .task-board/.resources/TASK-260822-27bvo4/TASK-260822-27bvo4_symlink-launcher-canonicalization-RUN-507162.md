# TASK-260822-27bvo4 — symlink launcher canonicalization (RUN-260822-507162)

Second, independent run of this task. Two spawn runs were queued for one task and
both went live; this note is the converged deliverable. See "Two runs" at the end.

Branch: `task/TASK-260822-27bvo4-symlink-launcher-507162` (base `origin/main` @ `6a9b201`)
Worktree: `.temp/TASK-260822-27bvo4/RUN-260822-507162`

## Question

Does manager executable identity hold when the binary is started through a
package-manager link — a shim on PATH, a symlinked install directory, a Windows
directory junction over the install directory? The manager profile orders that
resolution comes first and rejection second: substitution of the installed file
is the fault, the operator's launch shape is not.

## Where the code stands

`internal/godriver/build.go:166` resolves the manager once, from the process
itself — `resolveExecutableIdentity(managerExecutable())`, where
`managerExecutable()` (build.go:266) is `os.Executable()` and nothing else: no
manifest value, no PATH lookup. `resolveExecutableIdentity`
(`internal/godriver/identity.go`) makes the path absolute, canonicalizes it, and
only then applies the substitution battery in `readExecutableIdentity`:
regular-file check, `ModeSymlink` check, size bounds, `artifactHasMultipleLinks`
(hard link / reparse point), `os.SameFile` across the open, and SHA-256 over
exactly `Size` bytes with a trailing-byte check.

The canonical path is what everything downstream anchors on: `workerclient.go:106`
execs `plan.Executable.Path` rather than the launch path; `:144` sends it as the
expected identity; `workerserver.go:88` has the worker resolve *itself* the same
way and `:92` matches; `workerclient.go:179` and `build.go:255` re-verify.

So the ordering is right in principle. Two things needed proving.

## Finding 1 — os.Executable really does report the link (measured)

Not assumed — read out of a started process:

- darwin (`$GOROOT/src/os/executable_darwin.go`): returns the runtime's
  `executablePath` verbatim, absolutized against the initial cwd. No resolution.
- windows: `GetModuleFileName`, i.e. the launch path.
- linux (`executable_procfs.go`): `readlink /proc/self/exe`, already physical.

`TestExecutableIdentityResolvesARealSymlinkedProcessLaunch` starts this binary
through a shim link in a test-only hidden mode and reads back what the started
process was told, then records it either way:

```
darwin:  evidence: on darwin os.Executable reports the launch path
         ".../TestExecutableIdentityResolvesARealSymlinkedProcessLaunch.../001/curator"
         rather than the physical ".../go-build.../b001/godriver.test",
         so resolution is load-bearing here
windows: evidence: on windows os.Executable reports the launch path
         "C:\Users\admin\AppData\Local\Temp\...\001\curator.exe"
         rather than the physical "C:\Users\admin\godriver_win_merged.exe",
         so resolution is load-bearing here
```

Without the canonicalization step the very next check (`Lstat` +
`Mode().IsRegular()`) sees a symlink and rejects the manager's own installed
executable. Confirmed by mutation on darwin: replace the canonicalization with a
plain absolute path and the positive launch-shape tests go red with

```
go-v1 build_execution_worker_identity_invalid: the manager executable is not a canonical regular file
```

## Finding 2 — the Windows junction shape was actually broken (fixed)

`resolveExecutableIdentity` canonicalized with `filepath.EvalSymlinks`. That is
not the rule the rest of the driver resolves by — `physicalPath` is
(`selectToolchain`, `verifySelectedRoot`, the `mustPhysical` test helper, and the
long comment on `platform_windows.go:17-37`). On Windows a directory junction is
reported by `os.Lstat` as `ModeIrregular` since Go 1.23, `EvalSymlinks` follows
only `ModeSymlink`, and `walkSymlinks` refuses to descend through a component
whose mode is not a directory — so a path *through* a junction fails outright.

That is exactly an install-directory junction: how the GitHub Actions tool cache
publishes a directory, and a common way a machine-wide install points at a
versioned directory. The manager rejected its own installed executable before any
substitution check ran — a fail-open-shaped input turned into a hard fail-closed
on the operator's working install.

**Fix**: `identity.go:56` canonicalizes with `physicalPath(absolute)`. On unix
`physicalPath` *is* `filepath.EvalSymlinks(filepath.Clean(path))`
(`platform_unix.go:16`), so unix behaviour is bit-identical; on Windows it is
`GetFinalPathNameByHandle`, which resolves every link kind of every component in
one call. Resolving more makes the boundary stricter, not looser: the checks then
run against the physical file, so retargeting the link afterwards cannot move the
identity.

## Substitution controls still fail closed

`TestExecutableIdentityRejectsSubstitutionThroughALauncherLink` re-proves each
control with a link in front of it: dangling link, link to a directory, link to
an empty file, link loop, link to a hard-linked file. All reject with
`CodeWorkerIdentityInvalid`. Then, out of the table: a recorded identity stays
pinned to the physical file when the shim is retargeted (`Verify()` green,
`matches()` against the substituted file rejected), mutating the physical file's
bytes is still caught, and swapping the physical file itself for a link is
rejected by the re-proof.

`TestExecutableIdentityStillRejectsSubstitutionBehindAJunction` (Windows) does
the same for the junction shape: retargeted junction, dangling junction,
hard-linked executable behind a junction.

## Change set

| File | Change |
|---|---|
| `internal/godriver/identity.go` | `physicalPath` instead of `filepath.EvalSymlinks`; comment records the ordering rule and why resolving makes the checks stricter |
| `internal/godriver/main_test.go` | test-only `identityProbeMode` hidden mode + `runIdentityProbe` reporting what a started process is told about itself |
| `internal/godriver/worker_test.go` | `startRawWorkerFrom` + `workerScenario.launchPath`, so a real worker can be started from a link |
| `internal/godriver/identity_test.go` (new) | 5 tests, cross-platform |
| `internal/godriver/identity_windows_test.go` (new) | 2 junction tests, `//go:build windows` |

Tests in `identity_test.go`:

- `TestExecutableIdentityResolvesALauncherLink` — 4 launch shapes (shim symlink
  on PATH, symlink chain, symlinked install directory, unclean relative
  spelling) all resolve to the one installed identity; `matches()` and
  `Verify()` green for each.
- `TestExecutableIdentityRejectsSubstitutionThroughALauncherLink` — the
  fail-closed battery above.
- `TestExecutableIdentityResolvesARealSymlinkedProcessLaunch` — the platform
  evidence.
- `TestBuildAcceptsAManagerStartedThroughALauncherLink` — full `Build` with
  `managerExecutable` pointed at a link: the manager resolves and hashes itself,
  re-executes the canonical file as the worker, worker proves the same identity.
- `TestWorkerProvesTheInstalledIdentityWhenLaunchedThroughALink` — the half the
  previous one cannot reach. There the manager resolves a link but still execs
  the physical file, so the worker's own resolution never sees a link. Here the
  worker process itself is started from one and still has to prove the installed
  identity over the protocol (`workerserver.go:88`).

## Gates — real exit codes, each command run standalone, no pipes

darwin/arm64 (this host):

| Command | Exit |
|---|---|
| `go build ./...` | 0 |
| `gofmt -l cmd internal` | 0, no output |
| `go vet ./...` | 0 |
| `go test ./internal/godriver/ -count=1` | 0 (48.8s) |
| `golangci-lint run` | 0, "0 issues." |
| `GOOS=windows GOARCH=amd64 go vet ./internal/godriver/` | 0 |

windows/amd64, native host over `ssh win` (go1.25.5 windows/amd64), sources
synced into `C:\Users\admin\TASK-260822-27bvo4-507162`:

| Command | Exit |
|---|---|
| `go test ./internal/godriver/ -count=1` (with the fix) | 0 (93.7s) |
| `go test ./internal/godriver/ -run ...Junction -v -count=1` (pre-fix `identity.go`) | 1 — **expected red**, this is the defect the change fixes |
| `go test ./internal/godriver/ -count=1` (fix restored) | 0 (93.7s) |

Pre-fix red, verbatim:

```
=== RUN   TestExecutableIdentityResolvesALauncherReachedThroughADirectoryJunction
    identity_windows_test.go:47: EvalSymlinks("C:\...\002\current\curator.exe") = "", The system cannot find the path specified.
    identity_windows_test.go:52: a manager launched through a directory junction was refused:
        go-v1 build_execution_worker_identity_invalid: cannot canonicalize the manager executable
--- FAIL
GO_TEST_EXIT=1
```

`GOOS=windows GOARCH=amd64 golangci-lint run ./internal/godriver/...` exits 1
with 4 findings — `platform_windows.go` (errcheck),
`process_alive_windows_test.go` (errcheck), `controls_windows.go` x2 (gosec
G103). None of those files are touched by this change (`git diff origin/main
--name-only` plus the two new files is the whole change set), and the native
`golangci-lint run` gate the repo actually uses is clean.

`go vet ./...` needs `git submodule update --init --recursive` in a fresh
worktree: `agents/skills/skill-go-testing-tools` is a submodule and
`internal/ui` `replace`s `tuitestkit` onto it. Unrelated to this change.

Cross-compiled test binaries are not a substitute for the native run — a
`go test -c` for windows carries the build host's toolchain path, so the fixture
tests that compile a stub `go` fail with `fork/exec \Users\iv\.goenv\...\go.exe:
The system cannot find the path specified`. Only the pure-identity tests are
meaningful that way; the numbers above are from a real checkout on the host.

## Environment anomaly

The first `go test ./...` failed across `cmd/curator`, `internal/godriver`,
`internal/install`, `internal/install/atomicity` and `internal/transaction` with
`no space left on device`. The data volume sits at 99% (≈10 GiB free) with
~6.6 GiB of stale per-task worktrees under this repo's `.temp/` and a 32 GiB Go
build cache. Unrelated to this change — the same package passes standalone — but
the full suite is not reliably runnable at default parallelism on this host.

## Two runs

`RUN-260822-507162` (this one) and `RUN-260822-1e48d3` were both queued for this
task and both went live, the peer editing `.temp/TASK-260822-27bvo4/worktree`
while this run had just created it. This run moved to a separate run-scoped
worktree rather than fight over the files, per the `2035` logbook entry.

The peer finished, attached its artifacts, checked the board items — and then
ended without running the handoff, leaving the task in `development` with its
work uncommitted in `.temp/TASK-260822-27bvo4/worktree`.

The two implementations converged independently on the same one-line
`physicalPath` fix and the same probe-mode technique, which is itself worth
something: two runs that never saw each other's code reached the same diagnosis.
This deliverable is the union — the peer's test files and probe mode adopted
wholesale (better naming, wider junction coverage), with three cases grafted in
from this run that the peer's set lacked:

1. `launcher link loop` in the rejection table;
2. physical file swapped for a link after resolution → `Verify()` rejects;
3. `TestWorkerProvesTheInstalledIdentityWhenLaunchedThroughALink`, the
   worker-side resolution boundary.

The peer's artifacts on the board (`..._symlink-launcher-canonicalization.md`,
`..._windows-identity-tests.log`, `..._windows-prefix-baseline-red.log`) describe
the same fix and remain accurate for it; they predate the three grafted cases and
this run's Windows numbers.

## Not changed

`internal/managerlock/identity.go` also uses `filepath.EvalSymlinks`, but for the
lock path rather than the manager executable, and it has its own Windows handling
in `identity_windows.go`. Out of scope here — worth a separate look if the story
wants that boundary reviewed too. (Carried over from the peer's note; verified
independently.)
