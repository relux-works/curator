# TASK-260822-27bvo4 — symlink launcher canonicalization

Branch: `task/TASK-260822-27bvo4-symlink-launcher-507162` (from `origin/main` @ `6a9b201`)
Worktree: `.temp/TASK-260822-27bvo4/RUN-260822-507162`

## Question

Does manager executable identity hold when the binary is started through a
package-manager link (a shim on PATH, a junction over the install directory)?
Per the manager profile the resolution comes first and the rejection second:
substitution is the fault, the operator's launch shape is not.

## What the code does

`internal/godriver/build.go:166` resolves the manager once, from the process
itself:

```go
executable, err := resolveExecutableIdentity(managerExecutable())
```

`managerExecutable()` (build.go:266) is `os.Executable()`, nothing else — no
manifest value, no PATH lookup. `resolveExecutableIdentity`
(`internal/godriver/identity.go`) makes the path absolute, canonicalizes it,
and only then applies the substitution controls in `readExecutableIdentity`:
regular-file check, `ModeSymlink` check, size bound, `artifactHasMultipleLinks`
(hard link / reparse point), `os.SameFile` between the `Lstat` and the open
handle, and a full SHA-256 over exactly `Size` bytes.

The canonical path is what everything downstream is anchored on:

- `workerclient.go:106` execs `plan.Executable.Path`, not the launch path;
- `workerclient.go:144` sends that path as the expected identity;
- `workerserver.go:88` has the worker resolve *itself* the same way and
  `workerserver.go:92` matches it against what the parent sent;
- `workerclient.go:179` and `build.go:255` re-verify against the same path.

So the shape is correct in principle. Two things needed proving.

## Finding 1 — os.Executable really does report the link (evidence)

`os.Executable()` does not canonicalize on macOS or Windows:

- darwin (`$GOROOT/src/os/executable_darwin.go`): returns the runtime's
  `executablePath` verbatim, absolutized against the initial cwd. No symlink
  resolution.
- windows: `GetModuleFileName`, i.e. the launch path.
- linux (`executable_procfs.go`): `readlink /proc/self/exe`, already physical.

Measured, not assumed — `TestALaunchedProcessResolvesItsInstalledIdentity`
starts the test binary through a link in a test-only hidden mode and reads back
what the started process was told:

```
darwin:  os.Executable reports the launch path
         ".../TestALaunchedProcess.../001/curator-launcher.test";
         resolution recovered the installed
         "/private/var/folders/.../go-build.../b001/godriver.test"
windows: os.Executable reports the launch path
         "C:\Users\admin\AppData\Local\Temp\TestALaunchedProcess...\001\curator-launcher.exe";
         resolution recovered the installed
         "C:\Users\admin\godriver_win_test.exe"
```

Without the canonicalization step the very next check (`Lstat` +
`Mode().IsRegular()`) sees a symlink and rejects the manager's own installed
executable. Confirmed by mutation: replacing the canonicalization with a plain
absolute path turns all three new positive tests red, with

```
go-v1 build_execution_worker_identity_invalid: the manager executable is not a canonical regular file
```

So on macOS/Windows the canonicalization is load-bearing, not decorative.

## Finding 2 — the Windows junction shape was actually broken (fixed)

`resolveExecutableIdentity` used `filepath.EvalSymlinks`. That is not the rule
the rest of this driver resolves by: `physicalPath` is (see the long comment on
`internal/godriver/platform_windows.go`, and
`TestSelectResolvesAGoRootReachedThroughADirectoryJunction`). On Windows a
directory junction is reported by `os.Lstat` as `ModeIrregular` since Go 1.23,
`EvalSymlinks` follows only `ModeSymlink`, and `walkSymlinks` refuses to descend
through a component whose mode is not a directory — so a path *through* a
junction fails outright.

That is exactly an install-directory junction, which is how the GitHub Actions
tool cache publishes a directory and a common way a machine-wide install points
at a versioned directory.

Measured on a real Windows host (go1.25.5 windows/amd64), before the fix:

```
identity_launcher_windows_test.go:39: evidence: filepath.EvalSymlinks(
    "C:\...\current\curator") = The system cannot find the path specified.,
    which is the rejection this shape used to produce
identity_launcher_windows_test.go:44: a manager launched through an
    install-directory junction was rejected:
    go-v1 build_execution_worker_identity_invalid: cannot canonicalize the manager executable
--- FAIL: TestExecutableIdentityResolvesAnInstallDirectoryJunction
```

Fix: `identity.go` now resolves with `physicalPath(absolute)` instead of
`filepath.EvalSymlinks(filepath.Clean(absolute))`. On unix `physicalPath` *is*
`EvalSymlinks(Clean(...))`, so this is a no-op there; on Windows it is
`GetFinalPathNameByHandle`, which resolves every link kind of every component in
one call. After the fix the same test passes on the same host.

## Substitution controls still fail closed

`TestExecutableIdentityRejectsSubstitutionThroughALauncherLink` re-proves each
control with a link in front of it: dangling link, link to a directory, link
loop, hard-linked target behind the link, physical file swapped for a link after
resolution, mutated bytes behind the link. All reject with
`CodeWorkerIdentityInvalid`.

The property that makes accepting a link safe is pinned too: the identity is
recorded against the *physical* file, so retargeting the launcher link (unix) or
the install-directory junction (Windows) afterwards does not move it —
`Verify()` still names and still validates the originally installed file.

## Changes

- `internal/godriver/identity.go` — `physicalPath` instead of
  `filepath.EvalSymlinks`; documents why resolution precedes rejection and why
  resolving makes the checks stricter rather than more lenient.
- `internal/godriver/main_test.go` — test-only `launchShapeMode` hidden mode
  reporting what a started process is told about itself.
- `internal/godriver/worker_test.go` — `startRawWorkerFrom` / `workerScenario.launchPath`,
  so a real worker can be launched from a link.
- `internal/godriver/identity_launcher_test.go` (new) — 4 tests.
- `internal/godriver/identity_launcher_windows_test.go` (new) — junction test.

## Gates

| command | exit |
|---|---|
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `go test ./internal/godriver/ -count=1` | 0 |
| `golangci-lint run` | 0 |
| `GOOS=windows GOARCH=amd64 go vet ./internal/godriver/` | 0 |
| `GOOS=windows GOARCH=amd64 golangci-lint run ./internal/godriver/...` | 1 — see below |
| Windows host, targeted `-test.run` over the 4 launcher tests + junction test | PASS |

The Windows cross-lint reports 4 findings, all pre-existing in files this change
does not touch (`platform_windows.go` x1 errcheck, `process_alive_windows_test.go`
x1 errcheck, `controls_windows.go` x2 gosec G103). The changed and added files
are clean; `git diff origin/main --name-only` plus the two new files is the whole
change set.

`go vet ./...` needs `git submodule update --init` in a fresh worktree —
`agents/skills/skill-go-testing-tools` is a submodule and `internal/ui` has a
`replace` onto it.

## Environment anomaly

The first `go test ./...` run failed across `cmd/curator`, `internal/godriver`,
`internal/install`, `internal/install/atomicity` and `internal/transaction` with
`no space left on device`. The data volume is at 99% (10 GiB free) with ~6.6 GiB
of stale per-task worktrees under this repo's `.temp/` and a 32 GiB Go build
cache. Not related to this change — the same package passes standalone — but the
full suite is not reliably runnable on this host at default parallelism.

## Collision

Two spawn runs were queued for this task: `RUN-260822-507162` (this one) and
`RUN-260822-1e48d3`. The other run was live-editing
`.temp/TASK-260822-27bvo4/worktree` (branch
`task/TASK-260822-27bvo4-symlink-launcher`) while this run started, converging on
the same `physicalPath` fix. This run moved to a separate run-scoped worktree and
branch rather than fight over the files. Reviewer should expect two candidate
branches for one task.
