# BUG-260731-11bpa4 — call-line escaping fix, Windows and non-Windows evidence

Role: developer. Implements ARCH-RESOLVED checklist items 19–23 from
`BUG-260731-11bpa4_architecture-decision.md`. Signed commit
[`3cee1c1`](https://github.com/relux-works/curator/commit/3cee1c192d5c313d602ee814e7686e837b4127cd)
on PR [#10](https://github.com/relux-works/curator/pull/10), base `main`.

**Headline: the derived quadrupling rule is now empirically proven on
`windows-latest`. `TestWindowsPostInstallWrappersForwardArgumentsPathAndExitCode`
PASSES, and `internal/runtimestore` has zero failing cases on the Windows
runner. The bug's own acceptance criterion is met.**

The Windows *job* is still red, on 117 failures in other packages that are
pre-existing and out of this bug's package-scoped AC. That is quantified below,
not asserted.

---

## 1. What changed

Three files, all in `internal/runtimestore`. No CI script, ledger, spec pin or
protocol change.

| File | Change |
|---|---|
| `runtimestore.go` | `escapeCMDCallValue` (= `escapeCMDValue` applied twice) used on the `call` line only; `set "PATH=..."` stays at one pass; `WindowsShimContent` documents the two-pass arity and the out-of-contract `%VAR%` argument forwarding |
| `runtimestore_test.go` | `TestWindowsLauncherCarriesRuntimePathAndExitStatus` corrected from `%%` to `%%%%` on the call line — it had pinned the broken rule; new platform-neutral `TestWindowsLauncherEscapesEachExpansionPassOnce` |
| `targets_windows_test.go` | drops **only** the `percent%PATH%value` argument |

`escapeCMDCallValue` is deliberately expressed as `escapeCMDValue(escapeCMDValue(v))`
rather than a literal `%%%%` replacement, so the code states "one escape per
expansion pass" instead of a magic constant.

### Held to the architecture decision

- **Kept** in the Windows case: the space, embedded-quote, Unicode and
  empty-string arguments; the `%`-bearing artifact directory
  `immutable cache % Юникод`; the PATH assertion; the exit-code 37 assertion.
- **Not** deleted, skipped or reclassified. `platform-cases.tsv` row 61 is
  byte-identical to base (`must_run_on=windows`, `skip_allowed_on=-`, `class=-`),
  and `skip-classes.tsv` is untouched — `git diff bd6ba08` on both is empty.
- The retained `%`-bearing directory is what proves the fix: had the call line
  stayed at one pass, that directory alone would have failed the case.

### New test coverage

`TestWindowsLauncherEscapesEachExpansionPassOnce` runs on every host, not only
Windows. It pins two things:

1. the arity — `%%` on the set line, `%%%%` on the call line;
2. the **identity case** — a percent-free path is emitted unchanged, which is
   precisely what keeps the pinned conformance vectors byte-identical. A future
   change to the escaping that accidentally rewrote percent-free paths would
   break the protocol pin; this test catches it on Linux and macOS instead of
   waiting for the conformance job.

---

## 2. Windows evidence — CI run 30622852198, job 91131187109

Extracted from the uploaded gate artifact `test-evidence-windows-latest`
(`test/go-test*.json`), not from the job log, which prints only stage exit codes.

| Step | Result |
|---|---|
| 5 Verify Go toolchain identity | success |
| 7 `go vet` | **success** — the originally reported defect |
| 8 Ledger consistency | success |
| 9 `go test` + platform-case gate | failure (other packages, §3) |

`internal/runtimestore` on `windows-latest` — **37 case results, 0 failures**:

```
pass  <PACKAGE>
pass  TestWindowsPostInstallWrappersForwardArgumentsPathAndExitCode   <-- the AC target
pass  TestWindowsLauncherEscapesEachExpansionPassOnce                 <-- new
pass  TestWindowsLauncherCarriesRuntimePathAndExitStatus
pass  TestCandidateManagerLauncherContract                            <-- conformance
pass  TestAuthoritativeLauncherCasesForwardArgvPathRolesAndExitStatus <-- conformance
      (+ its two subcases)
pass  TestShimTransitionMatrixIsDeterministicAndManagerScoped (+3 subcases)
pass  TestParseHelperOutput* (+8 subcases), TestInstallSingleCommandAndShims,
      TestInstallRuntimeRootsOncePerCommit, TestRemoveStaleShims,
      TestRuntimeCommandPath, TestStagingRejectsLiveOverlapAndUnsafePathEntries,
      TestPrepareScriptRuntimeRejectsInvalidTypedInputs (+4 subcases)
skip  TestCompiledShimsStageWithoutLaunchThenPostInstallForwardExactly
skip  TestPrepareScriptRuntime* (3 cases)   } pre-existing platform-control class
skip  TestUnixLauncherCarriesRuntimePathArgumentsAndExitStatus
```

The four `skip`s are the pre-existing platform-control classification from
commit `8a68692`, unchanged by this commit.

**Both conformance tests are green on the Windows runner.** The escaping change
is the identity on their percent-free fixture paths, exactly as the architecture
decision predicted.

### The derived rule survived the runner

The architecture note flagged the quadrupling rule as *derived, not executed*,
and said the runner wins if it disagrees. It did not disagree — the rule is now
executed and green on `windows-latest`. No revision needed.

---

## 3. The Windows job is still red — none of it from this change

Baseline = commit `8a68692` (run 30620739038), head = `3cee1c1` (run 30622852198),
both from the same uploaded gate artifacts:

```
baseline 8a68692 failures: 117
head     3cee1c1 failures: 117

FIXED by this commit:
  - internal/runtimestore :: TestWindowsPostInstallWrappersForwardArgumentsPathAndExitCode

NEW at head:
  + cmd/curator :: TestCompiledInstallFollowsTheNativeControlInventoryExactly
```

**The one new failure is not from this change, and this is checked rather than
argued.** That test does not exist anywhere on this branch:

```
git grep -c TestCompiledInstallFollowsTheNativeControlInventoryExactly 3cee1c1 -- cmd/curator/status_test.go  -> ABSENT
git grep -c TestCompiledInstallFollowsTheNativeControlInventoryExactly 8a68692 -- cmd/curator/status_test.go  -> ABSENT
```

It arrived through `main`: the sibling lane's PR #11 ("Align the Linux lane with
the native control inventory") merged as `b6e523b` while this work was in flight,
and GitHub tests a merge commit of PR head + base. Its failure mode is
`go-v1 go_toolchain_missing: trusted GOROOT is not a real directory`, joining a
pre-existing cluster of 8 tests that emit the same error on the Windows runner in
the baseline run. It refuses at build planning, before any mutation — a stage
that never reaches shim bytes at all, which this change is the only thing to touch.

The remaining 116 are the untriaged Windows lane recorded in
`BUG-260731-11bpa4_windows-ownership-map.md` and owned by `BUG-260731-27h1yc`
and `BUG-260731-33v6zz`. `internal/globalbins ::
TestSafeSelectionFeedsStagedForwardingTargetWithoutLiveMutation` is worth naming
explicitly because `internal/globalbins` consumes `WindowsShimContent`: it fails
**identically in both runs**, so it is pre-existing and not a consequence of the
escaping change.

---

## 4. Non-Windows evidence

Every command below was run as a standalone process — no `tee`, no pipe — and
the exit code reported is the real one.

CI run 30622852198, all other jobs:

| Job | ID | Result |
|---|---|---|
| Test (ubuntu-latest) | 91131187045 | success |
| Test (macos-latest) | 91131187073 | success |
| **Lint** | 91131187116 | **success** |
| Race (ubuntu-latest) | 91131187120 | success |
| Interop conformance gate | 91131187101 | success |
| Naming gate | 91131187055 | success |
| Gate self-test ×3 (ubuntu/windows/macos) | 91131187082 / 91131187100 / 91131187162 | success |
| Test (windows-latest) | 91131187109 | failure — §3 |

Lint is now green. The two `unused` findings that the earlier developer run
correctly reported as unfixable from inside this ownership boundary
(`internal/godriver/controls_other.go`, `internal/transaction/namespace.go`)
belonged to the sibling scope and were resolved by PR #11 landing on `main`.

Local (macOS host, worktree at `3cee1c1`):

| Command | Exit |
|---|---|
| `go build ./...` | 0 |
| `go vet ./...` (host) | 0 |
| `GOOS=linux go vet ./...` | 0 |
| `GOOS=windows go vet ./...` | 0 |
| `GOOS=windows go vet ./internal/runtimestore/` | 0 |
| `GOOS=windows go test -c -o /dev/null ./internal/runtimestore/` | 0 |
| `go test ./internal/runtimestore/ ./internal/globalbins/ -count=1` | 0 |
| both conformance tests, `CURATOR_CONFORMANCE_ROOT` at `SPEC_PIN` `00b1688` | 0 |
| `bash .github/ci/ledger-consistency.sh` | 0 — 49 rows across linux darwin windows |
| `bash .github/ci/test-gate.sh` (darwin lane) | 0 — go test exit=0, platform-case gate exit=0 |
| `golangci-lint v2.12.2 run ./...` (host) | 0 — "0 issues" |
| `GOOS=linux golangci-lint v2.12.2 run ./...` | 1 — 2 pre-existing sibling findings, since fixed on `main` by PR #11 |

Two local-methodology notes, recorded so they are not re-derived:

- A first attempt used `go test ./... -count=1` with `CURATOR_CONFORMANCE_ROOT`
  exported for **every** package. That produced 8 spurious package failures
  (`vectors/build-drivers.json: no such file or directory`). It is the wrong
  gate: `.github/ci/test-gate.sh` deliberately splits served packages (root
  exported) from deferred ones (`env -u CURATOR_CONFORMANCE_ROOT`), because the
  pinned spec `00b1688` does not publish those vectors. Running the real gate
  script gives exit 0. Reproduce with `test-gate.sh`, never with a blanket
  `go test ./...`.
- `golangci-lint` caches by module path, and every task worktree here shares the
  module path `github.com/relux-works/curator`. A stale cache reported findings
  against *another agent's worktree* by absolute path. Re-run with a fresh
  `GOLANGCI_LINT_CACHE` when paths look wrong.

---

## 5. globalbins integration point — decision (checklist item 22)

`internal/globalbins/globalbins.go:353` decides shim ownership by byte-comparing
the stored file against a fresh `WindowsShimContent(canonical, nil)`.

**Decision: accept the reclassification. No `internal/globalbins` change.**

- Both sides recompute, so every shim written by this manager version stays
  self-consistent. There is no drift for new installs.
- The only artifact affected is a shim written by an *older* manager under a
  `%`-bearing canonical path. Such a shim is provably non-functional: its call
  line carries a single-doubled `%`, which the second expansion pass deletes,
  so it points at a mangled path. It has never launched anything. Declining to
  adopt a broken artifact is the correct reading, not a regression.
- The consequence is bounded and non-fatal. Both call sites
  (`globalbins.go:74`, `stage.go:84`) append an advisory message and continue;
  `unmanagedConflict` reports `command %q was not published … not managed by
  Curator`. No hard failure, no data loss, recoverable by deleting the file.
  Loud beats silent overwrite here.
- Teaching `ownedTarget` the legacy byte form would widen the accepted-bytes set
  to a form that never worked, for a rare path shape, and requires editing
  `internal/globalbins` — outside this bug's stated ownership boundary
  (`internal/runtimestore` only).

Recorded here so the behaviour is decided rather than discovered in the field.
If a future report shows real users hitting it, the cheap remedy is the
precedent already in that same function, which keeps recognising the older
symlink form specifically so an older manager's installation is adopted and
replaced.

---

## 6. Honest checklist state

**Met:** the bug's AC — `go vet` and `go test` pass for `internal/runtimestore`
on `windows-latest` in real Curator CI. Both conformance tests green. Ledger row
61 unchanged. Windows case not deleted, not skipped.

**Left unchecked — "land only after required CI is green":** the required
`Test (windows-latest)` job is red on 116 pre-existing out-of-scope failures plus
one that arrived from `main`. This change fixes the only one of them inside its
package scope. Landing is blocked on `BUG-260731-27h1yc` / `BUG-260731-33v6zz`,
not on anything in this diff. Merging PR #10 is therefore **not** done here — it
is a review decision, and it belongs to whoever owns the trade-off between
landing a correct scoped fix and holding for a fully green Windows lane.

**Also unchecked — independent Opus review:** not performed by this role.
