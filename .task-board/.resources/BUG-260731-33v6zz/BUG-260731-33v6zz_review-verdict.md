# BUG-260731-33v6zz — independent review verdict

**Verdict: ACCEPTED.**

Reviewed at the *actual PR head*, not the head the implementer's artifact names.

- PR: https://github.com/relux-works/curator/pull/13 (open, base `main`, `MERGEABLE`)
- Head: `11603469e090adf404020b717a30605f1cb134ee`
- Base: `origin/main` = `b30773b8ceb81a83a683418bee35cdc73d90da81` (PR 12 merged)
- Evidence: CI run `30638128300`, artifact `test-evidence-windows-latest` (`8797212256`),
  reviewer-parsed from raw `test/go-test.json` — not from a job summary.
- Baseline for the set-diff: run `30636020675`, current `main` on native windows-latest.

## AC — verified independently

> Curator `Test (windows-latest)` reports no failures for `cmd/curator`,
> `internal/buildcache`, `internal/globalbins`, `internal/managerlock` or
> `internal/staging`, with `.github/ci/platform-cases.tsv` and
> `.github/ci/skip-classes.tsv` unchanged or strengthened rather than relaxed,
> and macOS and Linux behavior preserved.

| package | main `b30773b` | head `1160346` | verdict |
|---|---|---|---|
| `cmd/curator` | 8 fail | **0 fail** | 63 pass, 0 top-level skip |
| `internal/buildcache` | 1 fail | **0 fail** | 51 pass, 4 skip (all pre-existing `root-unset`) |
| `internal/globalbins` | 1 fail | **0 fail** | 5 pass, 2 skip (all pre-existing `platform-control`) |
| `internal/managerlock` | 2 fail | **0 fail** | 17 pass, 0 skip |
| `internal/staging` | 1 fail | **0 fail** | 4 pass, 0 skip |
| **owned total** | **13** | **0** | |

All 14 originally-named cases re-checked individually in the raw JSON: every one is
`pass`. None is `skip`, none is absent.

Whole-lane residual at head is exactly **6** cases, none in an owned package:
`internal/transaction` (5) and `internal/godriver TestFingerprintReportsUnreadableDirectoryIdentically` (1).
Both were already red on `main` at the same commit-range baseline and are owned by
`BUG-260731-38dz6m` (backlog).

### No relaxation

- `git diff origin/main -- .github/` is **empty**. `platform-cases.tsv` and
  `skip-classes.tsv` are byte-unchanged.
- Artifact `ledger/ledger-consistency.txt`: `ok`, 56 rows across linux/darwin/windows;
  every required `cmd/curator` row reports `ok` on windows-latest.
- Reviewer skip set-diff, base vs head, at **every** nesting level: the only skips
  appearing at head that are absent at base are the two
  `TestCompiledProjectStatusRepairRollbackRecovery/.../{install,upgrade}` subtests.
  They are **not new**: `denyWrites` at `cmd/curator/status_test.go:1100` is
  byte-identical to `main:cmd/curator/status_test.go:1098` and is untouched by the
  diff. They were merely unreachable at base because the parent test failed first.
  `test/skips-observed.tsv` classifies both `allowed-host-capability`.
- The 4 `t.Skip` calls the branch adds are host-capability guards inside **new**
  Windows-only tests. None fired: all 7 new regression tests are `pass` in the raw JSON.

### No regressions

Reviewer set-diff over all `pass`/`fail`/`skip` events, base vs head: **zero** cases
that passed at base are anything other than `pass` at head.

### Other platforms

Run `30638128300`: `Test (ubuntu-latest)`, `Test (macos-latest)`, `Race (ubuntu-latest)`,
`Race (macos-latest)`, `Lint`, `Naming gate`, `Interop conformance gate` and
`Gate self-test` on all three OS are **green**. `Test (windows-latest)` is the only
red job, on the 6 unowned cases above.

Unix behavior is preserved *structurally*, not just empirically:
`internal/buildcache/protection_windows.go` is `//go:build windows`;
`godriver.physicalPath` on unix is `filepath.EvalSymlinks(filepath.Clean(path))`,
i.e. exactly the call it replaced; `runtimestore.Platform()` returns `"unix"` off
Windows, so the `globalbins` fixture change is a no-op on Linux and macOS.

Reviewer-local re-verification on darwin/arm64: `go build ./...` exit 0,
`GOOS=windows go vet ./...` exit 0.

## Scope

Out-of-scope packages verified byte-unchanged against `main`:
`git diff --stat origin/main -- internal/install/ internal/buildsource/ internal/transaction/ internal/runtimestore/`
produces **no output**.

**`internal/godriver` was touched, contrary to the literal scope exclusion — ratified.**
The exclusion existed because `BUG-260731-lepevi` was concurrent. That bug is `done`,
its PR 11 is merged, this branch is rebased onto it, and PR 11 touched
`controls_other.go` while this touches `session.go` / `platform_*.go` — no overlap.
7 of the 8 `cmd/curator` failures were the GOROOT junction defect, and only
`cmd/curator` exercises the real `build.Default.GOROOT` path, so the fix was not
reachable from inside the owned packages. The implementer flagged it in the board
notes and the PR body rather than leaving review to find it. Accepted.

## Implementation quality

Product surface is 4 files; everything else is tests.

- **`createProtectedDirectory`** — replaces create-then-protect with
  `windows.CreateDirectory` + `SECURITY_ATTRIBUTES` carrying `SE_DACL_PROTECTED`.
  This is the correct fix for the TOCTOU, and it is the Windows equivalent of the
  mode argument `mkdir(2)` already takes. `ERROR_ALREADY_EXISTS` is mapped to
  `os.ErrExist`, so the contract of the `os.Mkdir` it replaced is preserved — and
  pinned by `TestWindowsCreateProtectedDirectoryRefusesAnExistingName`.
- **`validateWindowsDACL`** — widening from `AceFlags != 0` to
  `AceFlags &^ windowsPropagationFlags != 0` is sound. `INHERITED_ACE` (0x10) and the
  audit flags are outside the permitted mask, the non-owner SID rejection is
  unchanged, and rights accumulation now skips `INHERIT_ONLY_ACE` ACEs so an
  inherit-only DACL still fails the mutation-rights requirement. The pre-existing
  negative cases, including `inherit-only owner allow`, are retained verbatim in
  `TestValidateWindowsSecurityPolicy`.
- **`prepareProtectedDirectory` / `ownedInheritingDirectory`** — repairs a genuine
  total blocker: the manager home is made by `os.MkdirAll`, whose perm argument
  Windows ignores, so no publication on Windows could ever satisfy the boundary.
  The repair is narrow and fails closed: reparse points refused, foreign owners
  refused, only an unprotected-DACL defect repaired, and full validation re-run
  afterwards rather than assumed. The negative test proves the boundary was not
  traded away for the repair.
- **`godriver.physicalPath`** — the directory-junction diagnosis is correct and
  well-evidenced (`IO_REPARSE_TAG_MOUNT_POINT` is a reparse-tag name surrogate, so
  Go ≥1.23 `os.Lstat` reports `ModeIrregular` and `EvalSymlinks` will not follow it).
  Anchoring the fingerprint and the `os.SameFile` root check on the physical target
  is strictly stricter than anchoring on the link. The
  `GetFinalPathNameByHandle` buffer-retry loop is correct: the too-small return
  includes the terminating null, so the second attempt always fits.
- Test edits are portability corrections, not weakenings. `staging` keeps every
  negative case and preserves the absolute-but-unclean fixture as absolute-but-unclean.
  `globalbins` is a no-op off Windows.

## Non-blocking findings — for follow-up, not rework

1. **Stale outcome artifact.** `BUG-260731-33v6zz_windows-lane-results.md` documents
   head `c6cb2df` / run `30630564858` and carries a "still red at 74 cases" table
   that predates PR 12 merging. At the reviewed head the residual is 6, and
   `internal/install`, `internal/install/atomicity` and `internal/buildsource` are
   green on `main`. The board notes carry the corrected figures and
   `BUG-260731-38dz6m` already owns the real residual, so nothing was misrouted —
   but the artifact should be refreshed via `update_resource` to head `1160346` /
   run `30638128300`. This verdict supersedes it as the current evidence of record.
2. **Residual TOCTOU in the repair path (low).** `ownedInheritingDirectory` proves
   its facts through a handle opened with `FILE_FLAG_OPEN_REPARSE_POINT`, then
   `protectWindowsPath` re-resolves the same *path* through
   `SetNamedSecurityInfo`, which follows reparse points. A swap in that window
   would redirect the DACL write. Exploitation needs write access to the parent of
   the manager home — i.e. the user's own privilege level or above — and the final
   `openWindowsProtected` re-validation fails closed, so impact is low. The clean
   form is `windows.SetSecurityInfo` on the already-open handle (reopened with
   `WRITE_DAC`), which is the same discipline `createProtectedDirectory` applies
   one function up.
3. **Test name overstates coverage.** `TestWindowsPublicationRefusesAHomeOwnedByAnotherPrincipal`
   grants World mutation rights on a home the current user still owns; it exercises
   the explicit-non-owner-grant branch, not the foreign-owner branch. The
   foreign-owner branch is covered only indirectly through
   `TestValidateWindowsSecurityPolicy`'s `wrong owner` case. A CI runner cannot
   create a foreign-owned directory without elevation, so the gap is understandable —
   the name should match what is proven.
4. **Partially tautological expectation.** `managerlock.identityBelow` spells `want`
   through the production `canonicalWithExistingPrefix`, so on Windows the
   identity-*spelling* assertion no longer pins anything independently. Mitigated:
   on unix the helper is a plain `filepath.Join`, so those hosts lose nothing; the
   Windows folding rule is asserted behaviorally by `identity_windows_test.go`
   (case-alias contention); and the branch adds a genuinely independent
   alias-twin assertion.
5. **First-run asymmetry on Windows (observation).** Only `ensureProtectedBase`
   repairs the home; `openProtectedEntry` does not. A cold `Inspect` on a
   never-published-into home therefore reports `UntrustedProvenance` rather than a
   plain miss until the first publication self-heals it. Strictly better than the
   prior state, where nothing on Windows worked at all, and both classifications
   lead to a rebuild — but establishing the boundary where the home is created
   (`managerlock.prepare`) would be the tidier home for it.

## DoD

| item | result |
|---|---|
| Classify all 14 Windows failures from raw go test JSON | met |
| Fix the five packages without skips | met — 0 failures, 0 added firing skips |
| Focused Windows regression coverage; Linux/macOS preserved | met — 7 new tests, all ran and passed; unix paths structurally unchanged |
| Signed PR to `main` with native windows-latest evidence | met — PR 13, all 9 commits `Good signature` |
| Outcome evidence attached, handed to independent review | met (see finding 1 on freshness) |
| Tests green / lint clean / build not broken | met — CI `Lint` green on head; local `go build` and `GOOS=windows go vet` exit 0 |
| Implementation matches AC | met |
| Solution fits project architecture | met — 4 product files, platform-split preserved, ledgers untouched |
