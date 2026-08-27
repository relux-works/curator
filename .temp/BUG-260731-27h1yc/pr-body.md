Fixes BUG-260731-27h1yc.

`Curator / Test (windows-latest)` has never produced a test result. `go vet ./...`
aborted the job at an earlier step on every run, so five packages beyond
`internal/runtimestore` had never executed on a Windows runner at all. PR #10
repaired that step; the failures below are what it uncovered, and they are
pre-existing, not caused by that PR.

**Stacked on #10.** This branch is cut from `task/BUG-260731-11bpa4-windows-vet`
(`8a68692`), because without its `go vet` fix the Windows job still aborts before
the test gate and no Windows evidence can be produced for this change either.
Merge #10 first; the diff below `8a68692` is this task's alone.

## What was wrong

| package | cases | cause |
|---|---|---|
| `internal/install` | 60 | fixtures pin `Platform: "unix"` |
| `internal/install/atomicity` | 8 | same |
| `internal/identity` (via `internal/install` dry-run) | 2 | a local Windows path is refused as a network source |
| `internal/buildsource` | 2 | two vectors do not exist on NTFS |

### `Platform: "unix"` on a Windows host

The install fixtures pinned the shim platform. That is not a portable default —
it is a shape Windows cannot build. A unix script runtime is validated for a
POSIX execute bit no file on Windows carries (`runtimestore/scripts.go:172`), and
a natively compiled artifact is refused against a unix shim
(`runtimestore/targets.go:96`). Every baseline install failed before asserting
anything.

The platform is now derived from the host, shim assertions name the launcher the
host writes (`.cmd` on Windows), and the user-bin forwarding-shim cases prepare
their PATH on every runner rather than exempting Windows from a class they exist
to cover.

### A local Windows checkout was not installable

`identity.Parse` sent `C:\repos\skill` to the scp-style fallback, which forbids a
backslash in the path part; the drive colon was then read as a host separator and
the source came back `malformed or unsupported network source`. `C:/repos/skill`
was already local by the single-letter-host rule — the backslash spelling now is
too. This is a product defect, not a test expectation, and `Canonical` hid it by
discarding the error.

### Two build-source vectors that do not exist on NTFS

`bad:name` opens an alternate data stream, so the entry that landed was the
perfectly portable name `bad`. The Windows vector is a trailing dot, which
`PortableComponent` refuses precisely because Win32 normalisation strips it,
created through the extended-length prefix that bypasses that normalisation.

A frozen root cannot be renamed away on Windows at all: `os.OpenRoot` takes no
`FILE_SHARE_DELETE`, so the kernel pins the directory for the life of the token.
That refusal is now asserted, and the replacement is driven the way the platform
allows one — by repointing a directory reparse point in the path prefix — so both
runners reach the same `Recheck` assertion.

## Ledger

`.github/ci/platform-cases.tsv` and `.github/ci/skip-classes.tsv` are untouched.
No case is deleted, skipped or platform-excluded; `TestGlobalInstall` and the
global rollback sweep now assert on Windows a class they previously skipped
there.

🤖 Generated with [Claude Code](https://claude.com/claude-code)
