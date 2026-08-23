# TASK-260822-27bvo4 — symlink launcher canonicalization

Manager executable identity when curator is started through a package-manager
link (shim symlink on PATH, symlinked install directory, Windows directory
junction over the install directory).

Branch: `task/TASK-260822-27bvo4-symlink-launcher`
Worktree: `.temp/TASK-260822-27bvo4/worktree` (base `origin/main` @ `6a9b201`)

## Finding

`internal/godriver/identity.go` canonicalized the running executable with
`filepath.EvalSymlinks`, while every other canonicalization in the driver
(`selectToolchain`, `verifySelectedRoot`, the test helper `mustPhysical`) uses
the driver's own rule, `physicalPath`.

On unix the two are identical — `physicalPath` *is*
`filepath.EvalSymlinks(filepath.Clean(path))` — so a symlinked launch already
resolved and the AC's "documents with evidence that the current path already
canonicalizes" branch holds for darwin and linux.

On Windows they are not the same rule, and this is a real fail-open→fail-closed
inversion: a manager installed under a directory the host publishes as a
**junction** was refused at the identity boundary before any substitution check
ran. Since Go 1.23 `os.Lstat` reports a junction as `ModeIrregular` rather than
`ModeSymlink`, `walkSymlinks` refuses to descend through it, and
`EvalSymlinks("<junction>\curator.exe")` fails outright with
"The system cannot find the path specified." That is the operator's launch
shape being treated as a fault, which is exactly what `profiles/manager.md`
orders against: resolution comes first, rejection targets substitution.
`internal/godriver/platform_windows.go` already documents this junction shape
as non-hypothetical (the GitHub Actions Go tool cache is published that way).

## Change

`internal/godriver/identity.go` — `resolveExecutableIdentity` now canonicalizes
with `physicalPath(absolute)` instead of `filepath.EvalSymlinks`. Behavior on
unix is bit-identical; on Windows every link kind is resolved in one
`GetFinalPathNameByHandle` call. Comment records the ordering rule and why
resolving makes the downstream checks stricter rather than looser: they are
applied to the physical file, so retargeting the link afterwards cannot move the
identity the manager and its worker agree on.

Nothing else changed. `readExecutableIdentity` still runs the full substitution
battery on the resolved path — `os.Lstat` regular-file check, size bounds,
`artifactHasMultipleLinks` (hard link / reparse point), `os.SameFile` across the
open, and full-content hashing with a trailing-byte check.

## Tests

`internal/godriver/identity_test.go` (new, cross-platform)

- `TestExecutableIdentityResolvesALauncherLink` — four launch shapes (shim
  symlink on PATH, symlink chain, symlinked install directory, unclean relative
  spelling) all resolve to the one installed file's identity; `matches()` and
  `Verify()` green for each.
- `TestExecutableIdentityRejectsSubstitutionThroughALauncherLink` — fail-closed
  cases through a link: dangling link, link to a directory, link to an empty
  file, link to a hard-linked file — all `worker_identity_invalid`. Then: a
  recorded identity stays anchored on the physical file when the shim is
  retargeted (`Verify()` green, `matches()` against the substituted file
  rejected), and mutating the physical file's bytes is still caught by
  `Verify()`.
- `TestExecutableIdentityResolvesARealSymlinkedProcessLaunch` — starts *this
  binary* through a shim link in a new test-only hidden mode and reads back what
  the started process resolved for itself, so `os.Executable`'s behavior for the
  platform's launch shape is observed rather than assumed.
- `TestBuildAcceptsAManagerStartedThroughALauncherLink` — full identity
  handshake: `managerExecutable` points at a shim link, the manager resolves and
  hashes itself, re-executes the canonical file as the worker, and the worker
  proves the same identity back over the protocol. Build succeeds.

`internal/godriver/identity_windows_test.go` (new, `//go:build windows`)

- `TestExecutableIdentityResolvesALauncherReachedThroughADirectoryJunction` —
  the regression case; logs the `EvalSymlinks` outcome as evidence.
- `TestExecutableIdentityStillRejectsSubstitutionBehindAJunction` — junction
  retargeted at another install does not move a recorded identity and the file
  it now reaches cannot pass as the same manager; a junction whose target is
  gone is rejected; a hard-linked executable behind a junction is rejected.

`internal/godriver/main_test.go` — added `identityProbeMode`, a test-only second
hidden mode alongside the existing `WorkerMode`, plus `runIdentityProbe`. Test
code only; the shipped binary's mode surface is unchanged.

## Evidence: os.Executable per platform

The real-launch probe prints what `os.Executable` reported for a launch through
a shim link:

- darwin/arm64: `.../001/curator` — the **link** path, not the target. Darwin
  reports the path the process was started with, so canonicalization is what
  makes the manager and the worker agree.
- windows/amd64: `...\001\curator.exe` — likewise the link path.
- linux would report the already-resolved `/proc/self/exe`; not run here (no
  local linux runner — CI ubuntu-latest covers it).

## Red → green on Windows

Ran on the native Windows host (`ssh win`, go1.25.5 windows/amd64), same
sources, only `identity.go`'s canonicalization line differing.

Pre-fix (`filepath.EvalSymlinks`):

```
=== RUN   TestExecutableIdentityResolvesALauncherReachedThroughADirectoryJunction
    identity_windows_test.go:47: EvalSymlinks("...\current\curator.exe") = "", The system cannot find the path specified.
    identity_windows_test.go:52: a manager launched through a directory junction was refused: go-v1 build_execution_worker_identity_invalid: cannot canonicalize the manager executable
--- FAIL
GO_TEST_EXIT=1
```

Post-fix (`physicalPath`): PASS, together with the substitution-behind-a-junction
case.

## Gate commands and real exit codes

Run standalone, no pipes.

| Command | Where | Exit |
|---|---|---|
| `go test ./internal/godriver/ -run 'TestExecutableIdentity\|TestBuildAcceptsAManagerStartedThroughALauncherLink' -v -count=1` | darwin/arm64 | 0 |
| `go test ./internal/godriver/ -count=1` | darwin/arm64 | 0 (29.2s) |
| `go build ./...` | darwin/arm64 | 0 |
| `gofmt -l cmd internal` | darwin/arm64 | 0, no output |
| `go vet ./...` | darwin/arm64 | 0 |
| `GOOS=windows GOARCH=amd64 go vet ./internal/godriver/` | cross | 0 |
| `golangci-lint run` | darwin/arm64 | 0, "0 issues." |
| `go test ./internal/godriver/ -run TestExecutableIdentity -v -count=1` | windows/amd64 (`ssh win`) | 0 |
| `go test ./internal/godriver/ -count=1` | windows/amd64 (`ssh win`) | 0 (248.8s) |
| `go test ./internal/godriver/ -run ...Junction -v -count=1` (pre-fix baseline) | windows/amd64 (`ssh win`) | 1 — **expected red**, this is the regression the change fixes |

`GOOS=windows GOARCH=amd64 golangci-lint run` exits 1 with 10 findings
(errcheck 5, gosec 4, revive 1) in `internal/buildcache/protection_windows.go`,
`internal/buildrepo/protection_windows.go`, `internal/godriver/controls_windows.go`,
`internal/godriver/platform_windows.go`, `internal/managerlock/identity_windows.go`
and its test. None of those files are touched by this change; the native
`golangci-lint run` gate this repo uses is clean.

Note: `go vet ./...` first failed on `internal/ui/ui_test.go` because the fresh
worktree had not initialized the `agents/skills/skill-go-testing-tools`
submodule that `go.mod` replaces `tuitestkit` with. Fixed with
`git submodule update --init --recursive` (exit 0); unrelated to this change.

## Not changed

`internal/managerlock/identity.go` also uses `filepath.EvalSymlinks`, but it
resolves the *lock path*, not the manager executable, and has its own
Windows-specific handling in `identity_windows.go`. Out of scope for this task —
worth a separate look if the story wants that boundary reviewed too.
