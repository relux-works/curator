# BUG-260731-33v6zz — Curator Windows lane, unowned packages

Outcome evidence for the five packages this bug owns: `cmd/curator`,
`internal/buildcache`, `internal/globalbins`, `internal/managerlock`,
`internal/staging`.

- PR: https://github.com/relux-works/curator/pull/13 (open, targets `main`, signed)
- Branch: `task/BUG-260731-33v6zz-windows-lane`
- Head: `c6cb2dff29844aa1554b41f7f52f598130567b48`
- Native evidence: CI run `30630564858`, artifact `test-evidence-windows-latest`
  (id `8793545541`), parsed from raw `test/go-test.json`, not from a job summary.

## Acceptance criteria — result

**Met.** `Test (windows-latest)` reports **zero failures** for all five owned
packages, ledgers are byte-unchanged, and macOS/Linux behavior is preserved.

| package | base (run 30620739038) | head c6cb2df | package verdict |
|---|---|---|---|
| `cmd/curator` | 8 | **0** | pass (63 top-level pass, 0 skip) |
| `internal/buildcache` | 2 | **0** | pass (51 pass, 4 pre-existing root-unset skips) |
| `internal/managerlock` | 2 | **0** | pass (17 pass, 0 skip) |
| `internal/globalbins` | 1 | **0** | pass (5 pass, 2 pre-existing platform-control skips) |
| `internal/staging` | 1 | **0** | pass (4 pass, 0 skip) |
| **total owned** | **14** | **0** | |

All 14 originally-named cases were verified individually as `pass` in the raw
JSON — none is `skip` and none is absent.

### No skips, no ledger relaxation

- `git diff origin/main -- .github/` is **empty**: `platform-cases.tsv` and
  `skip-classes.tsv` are byte unchanged.
- `ledger-consistency: ok` (56 rows across linux/darwin/windows).
- The platform-case gate reports `ok` for **every** required `cmd/curator` case.
- The 6 skips observed in owned packages are all pre-existing and classified
  (`root-unset`, `platform-control`); none was added by this branch.
- The two `TestCompiledProjectStatusRepairRollbackRecovery` subtest skips are
  **pre-existing on `main`** (`cmd/curator/status_test.go:1098`, byte-identical
  message) and classified `allowed-host-capability`.
- This branch adds 4 `t.Skip` calls, all host-capability guards inside **new**
  Windows tests. None fired on the runner — all three new tests ran and passed.

### No regressions

Set-diff of failing cases, base vs head: **zero** cases that passed at base now
fail. 17 cases went red→green (14 owned, 1 `runtimestore` from merged PR 10,
2 `internal/install` dry-run cases incidentally fixed by the identity change).

## Root causes — six stacked defects

Each fix uncovered the next. Two were product defects that made Curator
**unusable on Windows**, not test-only issues.

1. **`internal/buildcache` create-then-protect TOCTOU.** `ensureProtectedBase`
   created a directory then applied its DACL; a concurrent publisher observing
   the gap correctly refused inherited access. Now created atomically via
   `windows.CreateDirectory` with an owner-only `SE_DACL_PROTECTED` descriptor —
   what `mkdir`'s mode argument already gives Unix. 16 racers assert it.

2. **`internal/identity` rejected `C:\...`.** `scpRE` forbids a backslash in the
   path part, so a drive-letter path fell through to the fallback, which saw the
   colon, found no host, and rejected an ordinary local checkout as a malformed
   network source. A single letter is never a hostname, so both spellings are
   local on every platform. *(Hunk byte-identical to BUG-260731-27h1yc/PR 12 so
   both PRs merge cleanly; the identity unit coverage lives in that PR.)*

3. **`go-v1` refused the host's own GOROOT.** The Actions tool cache publishes Go
   behind a **directory junction** (`attributes=0x410`, `reparseTag=0xa0000003`
   = `IO_REPARSE_TAG_MOUNT_POINT`). Since Go 1.23 a junction is a reparse-tag
   name surrogate, so `os.Lstat` reports `ModeIrregular`, and
   `filepath.EvalSymlinks` follows only `ModeSymlink` — it left the junction
   unresolved, so `selectToolchain` found something neither directory nor link
   and refused it. `EvalSymlinks` also fails `ENOTDIR` for a junction mid-path,
   breaking `<junction>/bin/go.exe`. Now resolved through `physicalPath`
   (`EvalSymlinks` on unix, `GetFinalPathNameByHandle` on Windows). **Stricter,
   not looser**: the fingerprint and the `verifySelectedRoot` `os.SameFile`
   anchor now bind to the physical target, so retargeting the junction cannot
   move the trusted toolchain under a live session.

4. **No manager home was ever protected.** The home is made by `os.MkdirAll`; a
   Windows directory made that way inherits its parent's ACEs; the boundary
   requires an owner-only de-inherited DACL. So **no publication on Windows
   could ever succeed** against a real manager home. `internal/buildcache`'s own
   suite passed only because its fixture calls `protectTestHome` first.
   `ensureProtectedBase` now *establishes* the boundary on an owned, merely
   inheriting ancestor. **This is not a relaxation**: rewriting a DACL requires
   `WRITE_DAC`, which only the owner holds; the owner check runs first and is
   unchanged, so another principal's directory is refused, never seized; reparse
   points are refused; and full validation is re-run rather than assumed.

5. **The hardening then denied the owner its own files.** A non-inheritable ACE
   on a container hands children nothing, so an ordinary file created inside
   afterwards got an *empty* DACL and was unopenable by its own owner —
   surfacing as `Access is denied` on a manager lock, far from the cache. A
   container ACE now propagates, a file ACE does not, and `validateWindowsDACL`
   accepts propagation flags while still rejecting `INHERITED_ACE`. An
   inherit-only ACE contributes no effective rights, so an inherit-only DACL is
   still a refusal.

6. **Unix-shaped expectations, corrected rather than weakened.** An identity
   compared against the on-disk spelling where Windows folds case-insensitive
   components by design; POSIX path literals that are neither absolute nor clean
   under `filepath`; a runtime profile pinned to `"unix"` on a host with no POSIX
   execute bit; a cache snapshot restore that replaced mode bits but not the
   Windows boundary; `chmod 0o777` used to make a boundary unprovable, which on
   Windows only toggles the read-only attribute; and an artifact name pinned to
   `bin/build-tool` where the target reports `bin/build-tool.exe`.

## Regression coverage added

| test | package | ran on runner |
|---|---|---|
| `TestWindowsProtectedBaseIsNeverObservableUnprotected` (16 racers) | `buildcache` | pass |
| `TestWindowsCreateProtectedDirectoryRefusesAnExistingName` | `buildcache` | pass |
| `TestWindowsPublicationHardensAnInheritingManagerHome` | `buildcache` | pass |
| `TestWindowsPublicationRefusesAHomeOwnedByAnotherPrincipal` (negative) | `buildcache` | pass |
| `TestWindowsProtectedHomeStillServesOrdinaryManagerState` | `buildcache` | pass |
| `TestSelectResolvesAGoRootReachedThroughADirectoryJunction` | `godriver` | pass |
| `TestHostGoToolchainIsSelectableOnAnInventoryPlatform` | `cmd/curator` | pass |

The negative case is the important one: it proves the DACL hardening did not
weaken the boundary it repairs.

## Verification actually run

| command | host | real exit |
|---|---|---|
| CI `Test (windows-latest)` @ `c6cb2df` | native windows-latest | 1 — **red only outside this scope** (see below) |
| CI `Test (macos-latest)` | CI | 0 |
| CI `Test (ubuntu-latest)` | CI | 0 |
| CI `Race (ubuntu-latest)` / `Race (macos-latest)` | CI | 0 |
| CI `Lint` (golangci-lint v2.12.2) | CI | 0 |
| CI `Gate self-test` (all 3 OS) | CI | 0 |
| `go build ./...` | local darwin/arm64 | 0 |
| `go vet ./...` | local darwin/arm64 | 0 |
| `GOOS=windows go build ./...` | local cross | 0 |
| `GOOS=windows go vet ./...` | local cross | 0 |
| `go test ./internal/{buildcache,globalbins,managerlock,staging}/...` | local darwin | 0 |
| `go test -timeout 40m ./cmd/curator/...` | local darwin | 0 (1219s) |
| `bash .github/ci/no-broad-suppression.sh` | local | 0 |

**One command needed a retry, reported honestly.** The first local `go test` of
`cmd/curator` hit the **default 600s package timeout** (real exit 1). That is not
a defect: macOS CI passes the same package in 475s, CI itself uses `-timeout 30m`
with an in-file comment calling it "headroom for a slower hosted" runner, and
this host was at load average ~19 with 11 competing build processes. Rerun with
`-timeout 40m`: **exit 0 in 1219s**.

**Not run locally, and why:** `golangci-lint` is not installed on this host, so
lint was not reproduced locally. Lint rests on the pinned `v2.12.2` CI job, which
is green on this exact head SHA — a stronger signal anyway, since CI lints on
ubuntu.

## Ownership — flagged for review

**`internal/godriver` was touched, and was nominally out of scope.** It was
excluded because BUG-260731-lepevi was working there concurrently. That bug's
**PR 11 has merged**, this branch is rebased onto it, and PR 11 touched
`controls_other.go` rather than `session.go`, so nothing collides. The change was
unavoidable: 7 of the 8 `cmd/curator` failures were the GOROOT junction defect,
and only `cmd/curator` exercises the real `build.Default.GOROOT` path. Raising it
explicitly rather than leaving it for review to discover.

## Still red on windows-latest — NOT this bug's scope

`Test (windows-latest)` remains red at 74 cases, none in an owned package:

| package | cases | owner |
|---|---|---|
| `internal/install` | 58 | BUG-260731-27h1yc (PR 12, `reviewing`) |
| `internal/install/atomicity` | 8 | BUG-260731-27h1yc |
| `internal/buildsource` | 2 | BUG-260731-27h1yc |
| `internal/transaction` | 5 | **unowned** |
| `internal/godriver TestFingerprintReportsUnreadableDirectoryIdentically` | 1 | **unowned** |

**Coverage gap needing a new board item:** the last 6 were mapped to
BUG-260731-lepevi, but that bug's AC covered only the **Linux** lane. It is now
`done` and these survived it, so they are unowned again. **`Test
(windows-latest)` will still be red after both PR 12 and PR 13 merge** until
someone owns those 6. That blocks the story's green-CI goal and is the single
most important thing to route next.
