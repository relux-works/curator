# BUG-260731-33v6zz — curator-windows-lane-unowned-packages

Branch `task/BUG-260731-33v6zz-windows-lane`, rebased onto `main` after PR 10
and PR 11 merged. PR https://github.com/relux-works/curator/pull/13 → `main`.
Every commit is signed. Worktree: `.temp/BUG-260731-33v6zz/worktree`.

## The problem

14 failing top-level cases on `Test (windows-latest)` across five packages that
appeared in no board item: `internal/buildcache` 2, `internal/managerlock` 2,
`internal/globalbins` 1, `internal/staging` 1, `cmd/curator` 8. All pre-existing,
all masked until PR 10 let `go vet` through on the Windows runner.

They were not one bug. They were six, stacked: each fix exposed the next, and
the `cmd/curator` set turned out to be a chain in which **no compiled command
could ever be built or published on Windows at all**.

`.github/ci/platform-cases.tsv` and `.github/ci/skip-classes.tsv` are **byte
unchanged** — `git diff main -- .github/` is empty. No case was deleted, skipped
or platform-excluded to make a gate pass.

## Root causes, in the order the runner surfaced them

### 1. `internal/buildcache` — create-then-protect race (2 cases)

`ensureProtectedBase` created each cache directory with `os.Mkdir` and applied
its protected DACL afterwards. In that window the directory is observable
carrying its parent's inherited entries, so a concurrent publisher that opened
it refused the whole publication as untrusted provenance — a correct verdict
about state the manager had just created itself. Unix has no such window because
`mkdir` takes the mode at creation.

Fixed by creating the directory with the owner-only protected descriptor already
attached (`CreateDirectory` + `SECURITY_ATTRIBUTES` carrying `SE_DACL_PROTECTED`).
`TestWindowsProtectedBaseIsNeverObservableUnprotected` races 16 publishers at one
base and requires all 16 to succeed.

### 2. `internal/managerlock`, `internal/staging`, `internal/globalbins` (4 cases)

Unix-shaped expectations, not Windows defects. An identity compared against the
raw on-disk spelling when Windows folds case-insensitive components by design;
POSIX path literals that are neither absolute nor clean under `filepath` on
Windows; a runtime profile pinned to `"unix"` on a host with no POSIX execute
bit. Each now states the same property through the host's own rules and still
runs on every platform.

### 3. `internal/identity` — a Windows path read as a network remote (1 case)

`C:\...` left `Parse` through the scp fallback, which saw the drive colon, found
no host, and rejected an ordinary local checkout as a malformed network remote.
The carve-out already existed for `C:/...` via the single-letter-host rule; it
now covers both spellings.

**This hunk is byte-identical to the one in BUG-260731-27h1yc** (PR 12), which
shares the root cause for its `internal/install` cases and carries the unit
coverage. Taken identically so the two branches merge without conflict; no
change was made to `identity_test.go` here, so there is no other overlap.

### 4. `internal/godriver` — the host GOROOT is a directory junction (7 cases)

All seven remaining `cmd/curator` cases failed on their setup install with
`go_toolchain_missing: trusted GOROOT is not a real directory`, and passed on
macOS. `cmd/curator` is the only place the real `build.Default.GOROOT` path runs
— `internal/install` uses `fakeToolchain`, `internal/godriver` injects a
synthetic `hostFacts` — so nothing else could see it.

A diagnostic case reported the host facts from the runner itself:

```
build.Default.GOROOT="C:\hostedtoolcache\windows\go\1.25.5\x64"
  lstatDir=false mode=?rw-rw-rw- size=0 statDir=true
  attributes=0x00000410 directory=true reparsePoint=true reparseTag=0xa0000003
EvalSymlinks(GOROOT)="C:\hostedtoolcache\windows\go\1.25.5\x64" err=<nil>
```

`0x410` is `FILE_ATTRIBUTE_DIRECTORY|FILE_ATTRIBUTE_REPARSE_POINT`; `0xa0000003`
is `IO_REPARSE_TAG_MOUNT_POINT`. The GitHub Actions tool cache publishes Go
behind a **directory junction**. Since Go 1.23 (`os/types_windows.go`,
`fileStat.mode()`) a junction is a reparse-tag name surrogate, so the `ModeDir`
branch is skipped, and its tag is not `IO_REPARSE_TAG_SYMLINK`, so it falls to
`default: m |= ModeIrregular`. `filepath.EvalSymlinks` follows only
`ModeSymlink`, so it returned the junction unresolved and `selectToolchain`
refused an ordinary directory. `EvalSymlinks` is worse than incomplete for the
launcher: `walkSymlinks` refuses to descend through a non-directory mode, so
`<junction>/bin/go.exe` fails outright with `ENOTDIR`.

Fixed with `physicalPath` — `EvalSymlinks` on unix, `GetFinalPathNameByHandle`
on Windows — applied to the root, the launcher and `absolutePhysical`, so all
three agree about what a path is. **This is stricter than before, not looser**:
the returned GOROOT is the physical directory, so the fingerprint and the
`verifySelectedRoot` `os.SameFile` anchor bind to the target rather than to the
link, and retargeting the junction cannot move the trusted toolchain under a
live session.

### 5. `internal/buildcache` — no manager home was ever protected

With the toolchain selectable, the cases reached a real build and then failed on
`prepare protected cache root: ... DACL is not protected from inheritance`,
single-threaded, on the first publication.

The manager home is created by `os.MkdirAll`, and a Windows directory made that
way inherits its parent's ACEs. Nothing in Curator ever attached a protected
descriptor to it. The boundary requires an owner-only, de-inherited DACL, so
**no publication on Windows could ever succeed against a real manager home** —
`internal/buildcache`'s own suite passed only because its fixture calls
`protectTestHome` first. Unix has no equivalent gap: the `mkdir(0o700)`
`managerlock` already uses on the same home satisfies `validateUnixDir`, and an
owner-only protected DACL is the Windows spelling of that `0o700`.

`ensureProtectedBase` now establishes the boundary on an owned ancestor instead
of only asserting it, and only for that one defect. Rewriting a DACL needs
`WRITE_DAC`, which only the owner holds; the owner check runs first and is
unchanged, so a directory owned by another principal is refused rather than
seized, and one granting another principal explicit access is refused too.

### 6. `internal/buildcache` — a protected container that hands nothing down

The hardening fixed three of the seven and moved the rest to
`open ...\global\.curator-txn-...: Access is denied` and the same on a manager
lock — the owner locked out of its own new files.

The owner-only ACE used `NO_INHERITANCE`, right for a cache artifact and wrong
for a container. The manager home is the parent of everything else Curator
writes — manager locks, transaction files, the global root, the registry cache —
and all of those are created by ordinary `os` calls that rely on inheritance. A
protected DACL that propagates nothing gives each new child an **empty** DACL,
unopenable even by its owner.

A directory's entry now propagates to sub-containers and objects; a file's stays
non-inheritable; `validateWindowsDACL` accepts propagation flags while still
rejecting `INHERITED_ACE`, and an inherit-only entry still contributes no
effective rights, so the existing inherit-only refusal is unchanged.
`TestWindowsProtectedHomeStillServesOrdinaryManagerState` publishes into an
ordinary home and then does what the manager does next: `MkdirAll` a nested
state directory, write a lock file, read it back, reopen it for writing.

## Verification actually run

Every command was run directly, unpiped; the real exit code is reported.

| command | where | exit |
|---|---|---|
| `gofmt -l cmd internal` | worktree, macOS | 0, no output |
| `go vet ./...` | worktree, macOS | 0 |
| `GOOS=windows go vet ./...` | worktree | 0 |
| `GOOS=linux go vet ./...` / `GOOS=darwin go vet ./...` | worktree | 0 |
| `go test ./...` | worktree, macOS | **0** (full suite green) |
| `golangci-lint v2.12.2 run ./...` | worktree, host GOOS | **0**, 0 issues |
| `GOOS=windows golangci-lint v2.12.2 run ./internal/buildcache/... ./internal/godriver/...` | worktree | 1 — **byte-identical findings to the base commit** (pre-existing errcheck + gosec). CI's Lint job runs on ubuntu only and never lints windows-tagged files, so this was checked by hand against the same command on the unmodified tree. |

Native CI evidence, `Test (windows-latest)`, from raw `go test -json`:

| package | base (run 30620739038) | round 2 | round 5 | round 6 (run 30627771028) |
|---|---|---|---|---|
| `internal/buildcache` | 2 | **0** | **0** | see PR checks |
| `internal/managerlock` | 2 | **0** | **0** | " |
| `internal/globalbins` | 1 | **0** | **0** | " |
| `internal/staging` | 1 | **0** | **0** | " |
| `cmd/curator` | 8 | 7 | 4 | " |

`Test (macos-latest)`, `Test (ubuntu-latest)`, `Lint` and `Race (ubuntu-latest)`
are all green on this branch since the rebase onto merged `main` (run
30626599614).

## Ownership

`internal/godriver` was excluded from this bug's scope because BUG-260731-lepevi
was working in it concurrently. **That bug's PR 11 has merged**, this branch is
rebased onto it, and PR 11 touched `controls_other.go` rather than `session.go`,
so nothing here collides with it. The change restores the boundary's existing
intent — resolve links, then require a real directory — which a Go 1.23 stdlib
change had broken on Windows; it does not relax the trust posture. Flagged
explicitly in the PR body for review.

## Still red, and not in this bug's scope

- `internal/godriver TestFingerprintReportsUnreadableDirectoryIdentically` — was
  failing at the base commit, was mapped to BUG-260731-lepevi in the ownership
  map, and survived that PR. It is unowned again and needs a board item.
- `internal/install` (58), `internal/install/atomicity` (8),
  `internal/buildsource` (2) — BUG-260731-27h1yc, PR 12, open.
- `internal/transaction` (5) — was mapped to BUG-260731-lepevi and survived PR
  11 on Windows; that bug's AC covered the Linux lane, so these are unowned too.

## Notes for the next session

- No local Windows host exists here, so every Windows verdict comes from native
  `windows-latest` CI, parsed from raw `go test -json` rather than a job summary.
- Six CI rounds were needed because each fix uncovered the next layer; the
  diagnostic case `TestHostGoToolchainIsSelectableOnAnInventoryPlatform` is what
  made rounds 4-6 possible and is worth keeping.
