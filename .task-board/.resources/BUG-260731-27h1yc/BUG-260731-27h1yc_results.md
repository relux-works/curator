# BUG-260731-27h1yc — curator-windows-lane-masked-test-failures

PR: https://github.com/relux-works/curator/pull/12 (`task/BUG-260731-27h1yc-windows-lane` → `main`)
Signed commits (GitHub-signed, `verification.verified=true`): `2a02da9`, `c7bc890`

Scope, per the ownership map on BUG-260731-11bpa4: `internal/buildsource`,
`internal/install`, `internal/install/atomicity`.

---

## 1. Baseline: what was actually failing

The bug description named five cases. It understates the work by ~14x, which the
solution-architect had already recorded on this item. Measured, not estimated,
from the gate evidence artifact of a real windows-latest job:

**Baseline** — `main` @ `3a047d5` (PR 9, 10 and 11 all merged), run `30624569953`,
artifact `test-evidence-windows-latest`, `test/go-test.json`:

| package | failing top-level cases | owner |
|---|---|---|
| `internal/install` | **60** | this bug |
| `internal/install/atomicity` | **8** | this bug |
| `internal/buildsource` | **2** | this bug |
| `cmd/curator` | 9 | BUG-260731-33v6zz |
| `internal/transaction` | 5 | BUG-260731-lepevi |
| `internal/managerlock` | 2 | BUG-260731-33v6zz |
| `internal/buildcache` | 2 | BUG-260731-33v6zz |
| `internal/globalbins` | 1 | BUG-260731-33v6zz |
| `internal/godriver` | 1 | BUG-260731-lepevi |
| `internal/staging` | 1 | BUG-260731-33v6zz |
| **total** | **91** | |

**70 of those 91 are in this bug's three packages.**

None of this is a Windows regression. `go vet ./...` aborted the Windows job at an
earlier step on every run in this repository's history, so these packages had
never executed on a Windows runner at all. BUG-260731-11bpa4 repaired that step;
this is the first measurement.

## 2. Root causes

### A. `Platform: "unix"` pinned in every install fixture — 68 of 70 cases

`internal/install` and `internal/install/atomicity` set `opts.Platform = "unix"`
in their harnesses and in ~25 individual call sites. On a Windows host that is not
a portable default; it is a shape the platform cannot build:

* `runtimestore/scripts.go:172` validates a staged **unix** script runtime for a
  POSIX execute bit. No file on a Windows filesystem carries one, so every
  install failed with `validate staged script runtime: script command is not
  executable: <cmd>` — 69 occurrences across the two packages.
* `runtimestore/targets.go:96` refuses a natively compiled artifact against a
  unix shim: `compiled artifact target OS does not match shim platform` — 10
  occurrences. The fake toolchain already reports the **host** target, so a
  Windows build correctly produced `bin/<cmd>.exe`; only the pinned shim platform
  disagreed.

Every baseline install therefore failed before the case asserted anything. The
ledger's Windows rows for these packages were nominal coverage, not real.

**Fix.** The platform is derived from the host (`runtimestore.Platform()`), shim
assertions name the launcher the host actually writes (`shimName()` → `.cmd` on
Windows), and the user-bin forwarding-shim cases prepare their PATH on every
runner instead of exempting Windows.

### B. `internal/identity` refused a local Windows checkout — real product defect

Surfaced by `TestProjectDryRunKeepsEveryEphemeralPathInOneOperationPrivateRoot`
and its global twin, which declare a cloned source by its local path:

```
invalid source for build-skill: C:\Users\...\origin\build-skill;
malformed or unsupported network source
```

`identity.Parse` reaches an scp-style fallback whose path group is `[^\\]+` — it
forbids a backslash. `C:\repos\skill` therefore never matched, fell through to the
"contains a colon, so it is a malformed remote" branch, and was refused.
`C:/repos/skill` was already local by the existing single-letter-host rule; the
backslash spelling was not.

**This is a product bug, not a test expectation: `curator` could not install a
skill from a local path on Windows at all.** `TestCanonical` had a
``C:\repos\skill`` case that passed only because `Canonical` discards `Parse`'s
error.

**Fix.** A drive-letter prefix in either spelling is a local filesystem source
with no network identity, alongside the existing `/`, `./`, `../`, `~` and
`file:` forms. The carve-out is exactly one letter wide, so a real hostname
before a colon stays a network remote and stays subject to the allowlist; a new
`Parse`-level test pins both halves.

### C. `internal/buildsource` — two vectors that do not exist on NTFS

**`TestRejectsInvalidPathsLinksAndCollisions/invalid_protocol_path`** wrote
`bad:name`. On NTFS a colon opens an *alternate data stream*, so the entry that
actually landed was the perfectly portable name `bad`, `Validate` returned `nil`,
and the case asserted nothing. It also leaked the returned token, whose open root
handle then failed the temporary directory's own cleanup.

The Windows vector is a trailing dot — `PortableComponent` refuses it precisely
because Win32 path normalisation *strips* it, which would alias two distinct
protocol paths onto one file — created through the extended-length `\\?\` prefix
that bypasses that normalisation, and removed the same way. That is a sharper
case than the POSIX one: it proves the validator, not the path layer, is what
says no.

**`TestFrozenTokenRejectsRootReplacement`** (required, `skip_allowed_on = -`)
failed in its *setup*: `os.Rename` of the validated root returned
`The process cannot access the file because it is being used by another process`.
`os.OpenRoot` → `syscall.Open` requests `FILE_SHARE_READ|FILE_SHARE_WRITE` and
never `FILE_SHARE_DELETE` (Go 1.25.5, `syscall/syscall_windows.go:396`), so the
kernel pins the validated directory for the life of the token — and pins every
enclosing directory too, because Windows refuses to rename a directory with an
open handle in its subtree.

The fixture is now platform-split. Windows asserts that refusal — it is the
platform's own half of "a frozen build source rejects replacement of its root" —
and then drives the replacement the way the platform *does* allow one, by
repointing a directory reparse point in the path prefix, so both runners reach
the same `Recheck` assertion. A junction needs no privilege; the helper prefers a
symbolic link and falls back to `mklink /J`, and **fails** rather than skips,
because this row tolerates no skip.

### D. `TestReadDocumentBindsGenerationToBytesReplacedByRename`

Same family: `os.Open` holds the declaration document without
`FILE_SHARE_DELETE`, so Windows refuses the atomic rename over it while the read
is in flight. The schedule is now platform-split — POSIX replaces inside the read
window, Windows asserts the refusal and settles the replacement the moment the
read closes. All three properties the case asserts are unchanged.

### E. `TestAuditGateBlocksUndeclaredNetwork`

The `skill-net` fixture declares `unix_path` alone. The spec permits that
(`skillspec/parse.go:263`: a script command requires `unix_path` **or**
`win_path`), and a skill that names no path for the host is refused during
runtime validation, before the audit verdict the case exists to assert is
consulted. The fixture now exports the command on both platforms. The
one-platform-only shape is a separate rule, already covered by the two POSIX
launcher cases in the same file, which skip on Windows under the existing
`platform-control` classification.

## 3. Evidence

**Windows, `Test (windows-latest)`, run `30626331508`, head `c7bc890`** — artifact
`test-evidence-windows-latest`:

| package | baseline (`main` @ `3a047d5`) | this PR | verdict |
|---|---|---|---|
| `internal/install` | 60 failing | **0 failing** | **pass** |
| `internal/install/atomicity` | 8 failing | **0 failing** | **pass** |
| `internal/buildsource` | 2 failing | **0 failing** | **pass** |

Required ledger rows, from the platform-case gate's own report on the Windows
runner (`test/platform-cases.txt`):

```
ok    internal/buildsource :: TestFrozenTokenRejectsMutation
ok    internal/buildsource :: TestFrozenTokenRejectsRootReplacement
ok    internal/buildsource :: TestWithValidatedOrdersCallbackAndRejectsMutation
ok    internal/install :: TestDryRunTouchesNothing
ok    internal/install :: TestEndToEndInstall
ok    internal/install/atomicity :: TestFailureAtEveryTargetClassRestoresPriorStateInReverseOrder
ok    internal/install/atomicity :: TestStaleAdapterRemovalRollsBackToTheExactPriorEntry
ok    internal/install/atomicity :: TestAdapterMirrorLinksAreJournaledAndRestoredExactly
```

`Test (windows-latest)` as a whole stays red: `cmd/curator`,
`internal/managerlock`, `internal/buildcache`, `internal/staging`,
`internal/globalbins` (BUG-260731-33v6zz) and `internal/transaction`
(BUG-260731-lepevi) are other owners' packages and were never in this bug's
scope. That is the state the ownership map predicted.

**Other lanes, same run:** `Test (ubuntu-latest)`, `Test (macos-latest)`,
`Race (ubuntu-latest)`, `Race (macos-latest)`, `Lint`, `Naming gate`,
`Interop conformance gate` and all three `Gate self-test` jobs — all **success**.

**Local (darwin/arm64, go1.25.5):** `go test -count=1 ./internal/install/...
./internal/identity/ ./internal/buildsource/ ./internal/runtimestore/
./internal/globalbins/` — all `ok`. `gofmt -l internal/` clean. `go vet ./...`
and `GOOS=windows go vet ./...` clean.

## 4. Ledger

`.github/ci/platform-cases.tsv` and `.github/ci/skip-classes.tsv` are **byte
identical to main**. No case is deleted, skipped or platform-excluded.

`test/skips-observed.tsv` on the Windows runner is identical to the `main`
baseline's — same rows, same classes, same reasons; the only textual difference
is a `t.TempDir` nonce inside one unrelated `internal/scopes` reason string. No
skip was added.

The ledger is **strengthened**, not relaxed: `TestGlobalInstall` and the global
half of `TestFailureAtEveryTargetClassRestoresPriorStateInReverseOrder` now
assert the PATH-visible forwarding shim and the user-bin mirror ledger on
Windows, a class they previously bypassed there with a `runtime.GOOS != "windows"`
guard.

## 5. Deviations from the recorded directive, and why

The directive said to branch from `main`. The branch was cut from PR 10's head
(`8a68692`) instead, because on `main` at that moment `go vet ./...` still aborted
the Windows job before the test gate and **no Windows evidence was producible for
this change at all** — which the Definition of Done requires. PR 10 has since
merged; `main` was merged into the branch (`a164dca`) and the PR's diff against
`main` is now this task's alone. `internal/runtimestore` and the Linux inventory
files were never modified here.

`internal/identity` is not named in the ownership map. It is changed because two
of `internal/install`'s own failing cases fail *inside* it, and because the defect
is real user-facing Windows behaviour rather than a test expectation.
