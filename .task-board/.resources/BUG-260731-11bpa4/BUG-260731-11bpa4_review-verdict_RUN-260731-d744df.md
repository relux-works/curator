# BUG-260731-11bpa4 — review verdict: **ACCEPTED**

Reviewer run `RUN-260731-d744df` (Opus 5, read-only). Reviewed head
`3cee1c192d5c313d602ee814e7686e837b4127cd` on PR
[#10](https://github.com/relux-works/curator/pull/10), base `main`, diffed against
`bd6ba08` (PR 9 head, the mandated base).

Every claim below was re-derived from primary sources by this review — CI job
steps, the uploaded gate artifacts, the GitHub API and the source tree. Nothing
is carried over from the developer's report on trust.

---

## 1. Verdict

**Accepted.** The acceptance criterion — *"go vet and go test pass for
`internal/runtimestore` on windows-latest in Curator CI"* — is met and proven on
the runner. The diff introduces zero new failures anywhere.

One action remains and it is **not** the reviewer's: **PR 10 is still open.**
See §7.

---

## 2. The AC, verified on the runner

CI run `30622852198`, job `Test (windows-latest)` `91131187109`, step conclusions
read from the Actions API:

| # | Step | Conclusion |
|---|---|---|
| 5 | Verify Go toolchain identity | success |
| 7 | **`go vet`** | **success** ← the reported defect |
| 8 | Ledger consistency | success |
| 9 | go test + platform-case gate | failure (other packages, §4) |

`internal/runtimestore` in `test-evidence-windows-latest/test/go-test.json`,
re-parsed by this review: **37 case results, 0 failures.**

```
pass  TestWindowsPostInstallWrappersForwardArgumentsPathAndExitCode   <- the AC target
pass  TestWindowsLauncherEscapesEachExpansionPassOnce                 <- new
pass  TestWindowsLauncherCarriesRuntimePathAndExitStatus
pass  TestCandidateManagerLauncherContract                            <- conformance
pass  TestAuthoritativeLauncherCasesForwardArgvPathRolesAndExitStatus <- conformance (+2 sub)
pass  TestParseHelperOutput* (+8 sub), TestShimTransitionMatrix* (+3 sub),
      TestInstallSingleCommandAndShims, TestInstallRuntimeRootsOncePerCommit,
      TestRemoveStaleShims, TestRuntimeCommandPath,
      TestStagingRejectsLiveOverlapAndUnsafePathEntries,
      TestPrepareScriptRuntimeRejectsInvalidTypedInputs (+4 sub)
skip  TestCompiledShimsStageWithoutLaunchThenPostInstallForwardExactly  (pre-existing)
skip  TestPrepareScriptRuntime* x3, TestUnixLauncherCarries*            (§5)
```

Both conformance tests pass **on the Windows runner**, so the pinned protocol
shape is intact under the new escaping.

---

## 3. The fix is correct, not merely green

`escapeCMDCallValue` (`internal/runtimestore/runtimestore.go:216`) is applied
**only** to the `call` line (`:191`); `set "PATH=..."` (`:182`) stays at one pass.
Checked against the emitted template:

- The `set` lines sit inside a plain `if defined PATH ( … )` block. A
  parenthesised block is percent-expanded once at parse time → **doubling is
  correct**.
- `call` re-parses the remainder of its own line after `%*` substitution → **two
  passes → quadrupling is correct**. `%%%%` → `%%` (pass 1) → `%` (pass 2).
- Writing it as `escapeCMDValue(escapeCMDValue(v))` rather than a literal `%%%%`
  states the rule ("one escape per expansion pass") instead of a magic constant.
  Good call.
- The escaping is the **identity** on `%`-free paths, which is what keeps the
  pinned conformance vectors byte-identical. Verified twice: both conformance
  tests green on windows-latest, and green locally against `SPEC_PIN`
  `00b1688a9b2457ca397a0bb550acf47cad8ee967`.

Cross-contamination check done by hand: after pass 1 the path region holds `%%`,
which pass 2 consumes as one escaped literal. It can no longer pair with a stray
`%` arriving from `%*`. That is precisely the old corruption
(`immutable cache PATHvalue`), and it is closed at the source rather than papered
over.

---

## 4. Nothing in this diff broke anything

Baseline `8a68692` (run `30620739038`) vs head `3cee1c1` (run `30622852198`),
both re-parsed from the uploaded artifacts by this review:

```
baseline failures: 117
head     failures: 117
FIXED: internal/runtimestore :: TestWindowsPostInstallWrappersForwardArgumentsPathAndExitCode
NEW:   cmd/curator :: TestCompiledInstallFollowsTheNativeControlInventoryExactly
```

The one new failure is **not from this change**, and this review proved it
independently rather than accepting the assertion. Via the GitHub contents API:

- `cmd/curator/status_test.go` on **`main`** contains
  `TestCompiledInstallFollowsTheNativeControlInventoryExactly` (3 occurrences).
- **No file on `task/BUG-260731-11bpa4-windows-vet` contains it at all.**

It arrived through `main` (`b6e523b`, sibling PR 11) into the merge commit GitHub
tests, and fails with the pre-existing
`go_toolchain_missing: trusted GOROOT is not a real directory` cluster. This is
exactly what LOGBOOK 1402 predicted. `internal/globalbins ::
TestSafeSelectionFeedsStagedForwardingTargetWithoutLiveMutation` fails
identically at baseline and head, so that package — which consumes
`WindowsShimContent` — shows no escaping fallout.

---

## 5. The non-goal was honoured

The task forbade deleting or skipping the Windows case to make the gate pass.
Checked at source:

- The case is **not** deleted, **not** skipped, **not** reclassified.
  `platform-cases.tsv` row 61 is byte-identical to base
  (`must_run_on=windows`, `skip_allowed_on=-`, `class=-`); `skip-classes.tsv` is
  untouched; `git diff bd6ba08 3cee1c1 -- .github/ internal/interop/` is empty.
- Exactly **one fixture argument** (`percent%PATH%value`) was dropped, under the
  explicit ARCH-RESOLVED authorisation. Everything row 61 names is still
  asserted: arguments (space, embedded quote, Unicode, empty), PATH, exit code 37.
- **The `%VAR%` coverage on the path side is fully retained** and is stronger
  than the report claims: `helperPath` = `dependency path %PATH% Юникод` is
  asserted verbatim through the `set` line, `artifactDir` =
  `immutable cache % Юникод` exercises the `call` line, and all three shim
  destinations carry a `%`. Only the *argument* side lost its `%VAR%` case, which
  is the narrowed contract.

The three `PrepareScriptRuntime*` skips are justified at source, not asserted:
`scripts.go:172` demands `Perm()&0o111 != 0` when `platform == "unix"`, and no
file on a Windows host carries a POSIX execute bit — the fixture is
unconstructible there, not merely unasserted. They classify as
`allowed-platform-control` under an existing `skip-classes.tsv` regex, none of
the three is in `platform-cases.tsv`, and none ever ran on Windows before
(the package did not compile). No coverage regression.

---

## 6. Fit with the project

- **Ownership boundary held.** Four files, all in `internal/runtimestore`. No CI
  script, ledger, spec pin, protocol or sibling-scope file touched.
- **Test placement is the right answer to a real constraint.** `parseHelperOutput`
  lives in the platform-neutral test file with 8 table subcases so linux/darwin
  compile and exercise the parsing contract; only the thin `*testing.T` wrapper
  stays windows-only, because `unused` runs on ubuntu where the calling test is
  not in the build. That is the correct shape, arrived at from an observed lint
  failure rather than guessed.
- **New coverage earns its place.** `TestWindowsLauncherEscapesEachExpansionPassOnce`
  pins the arity *and* the identity case on every host, so a future escaping
  change that rewrote `%`-free paths — which would silently break the protocol
  pin — fails on Linux instead of waiting for the conformance job.
- **The globalbins consequence was decided, not discovered.** Verified at source:
  `ownedTarget` (`globalbins.go:347`) byte-compares, and both call sites
  (`globalbins.go:74`, `stage.go:84`) append a non-fatal advisory and `continue`.
  A pre-existing shim under a `%`-bearing path is provably non-functional, so
  declining to adopt it — loudly — is the right behaviour. Accepting it with no
  `internal/globalbins` change is correct and stays inside the boundary.
- Signatures: all three commits GitHub-verified `valid`. PR 10 targets `main`.
- Non-Windows: every other job on run `30622852198` is green, including **Lint**.
  Independently re-run locally at `3cee1c1`: `go build ./...` 0;
  `GOOS=windows go vet ./...` 0; `GOOS=linux go vet ./...` 0;
  `GOOS=windows go test -c` 0; `go test ./internal/runtimestore ./internal/globalbins` ok;
  both conformance tests against the pinned spec ok; `golangci-lint` on the
  package with a fresh cache **0 issues**.

---

## 7. The landing decision — the reviewer's call, made here

DoD item 4 reads *"Obtain independent Opus review and land only after required CI
is green."* The review half is this document. On the landing half:

- **`main` has no branch protection** — `GET /repos/relux-works/curator/branches/main/protection`
  returns `404 Branch not protected`. There is no GitHub-enforced required-check
  set; "required CI green" is a project-policy notion, not a gate.
- `Test (windows-latest)` is red on 117 cases. **Zero are attributable to this
  diff** (§4): 116 are pre-existing and owned by `BUG-260731-27h1yc` /
  `BUG-260731-33v6zz`, and 1 arrived from `main` via sibling PR 11.
- Holding PR 10 **blocks both of those bugs**. Until it lands, `main`'s Windows
  job still dies at step 7 `go vet`, so `go test` never runs there and neither
  owner can obtain a baseline. This PR is the prerequisite that makes the whole
  Windows lane observable.

**Recommendation: land PR 10.** It is a correct, package-scoped, fully proven fix
that regresses nothing and unblocks the two bugs that own the remaining redness.

**I did not merge it.** The reviewer archetype is read-only and does not supply
`commit_ack`; merging to `main` is the commit-owning mover's action. This section
is the acceptance evidence for that mover.

---

## 8. Checklist honesty

**Item 4 left UNCHECKED.** The independent Opus review is complete, but PR 10 is
not landed, so the item as worded is not satisfied. Ticking it would be false.
Every other item is checked against evidence that this review re-derived.

## 9. Non-blocking notes for later, not rework

1. `parseHelperOutput` returns `fmt.Errorf("helper output carried no JSON payload")`
   with no format verbs; `errors.New` is idiomatic and `errors` is already
   imported. Test-only, lint-clean. Not worth a rework cycle.
2. `scripts.go` supports `platform == "windows"` script runtimes (`:51`, `:124`,
   `:223`, `:231`), but no test constructs one — the three unix fixtures now skip
   there. **Pre-existing gap, not created by this change** (the package never
   compiled on Windows). Belongs with the Windows-lane bugs.
3. With the `%`-bearing argument gone, the *path-`%` pairs with argument-`%`*
   scenario is no longer exercised. That is fine: correctly escaped, the path's
   `%%` self-consumes and cannot pair, and an argument carrying `%` is now
   out of contract by design. The surviving lone-`%`-in-path case is the one that
   proves the fix. Noted so nobody re-derives it as a gap.
