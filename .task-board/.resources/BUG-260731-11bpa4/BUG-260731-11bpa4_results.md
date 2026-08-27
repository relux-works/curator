# BUG-260731-11bpa4 — outcome

PR: https://github.com/relux-works/curator/pull/10
Branch: `task/BUG-260731-11bpa4-windows-vet`, base `bd6ba08` (head of PR 9)
Signed commits: `31720f1`, `8a68692` (both `verification.verified=true`)

## 1. The reported defect is fixed and proven in real CI

`decodeHelperOutput` was **never declared anywhere in the repository's history**.
`cfffd7c` added the call site at `internal/runtimestore/targets_windows_test.go:97`
without the function, so the Windows build of `internal/runtimestore` has never
compiled. It stayed invisible while `toolchain-identity.sh` failed every Go job at
step 4; PR 9 repaired that gate and uncovered it.

Reproduced locally before the change:

```
GOOS=windows go vet ./internal/runtimestore/
vet: internal/runtimestore/targets_windows_test.go:97:14: undefined: decodeHelperOutput
exit 1
```

Real CI, run 30620739038, job `Test (windows-latest)`:

| step | conclusion |
|---|---|
| `go vet` | **success** (this is the step that was failing) |
| `Ledger consistency` | **success** |
| `go test + platform-case gate` | failure — see §3 |

The Windows case was neither deleted nor skipped. It now compiles and executes
for the first time.

### Design of the fix

`parseHelperOutput` (pure, returns `(map[string]any, error)`) lives in the
platform-neutral `targets_test.go` with its own table tests, so linux and macOS
compile and exercise the parsing contract. Only the thin `*testing.T` wrapper
`decodeHelperOutput` is windows-only — it has to be, because `golangci-lint`'s
`unused` check runs on ubuntu where the calling test is not in the build. A
shared copy fails `Lint` with `func decodeHelperOutput is unused`; this was
observed, not predicted.

## 2. Four further Windows failures in this package, also fixed

With the package compiling, its tests ran on `windows-latest` for the first time.
Five failed. Four were repaired:

- `TestShimTransitionMatrixIsDeterministicAndManagerScoped` — described its
  compiled artifact with the host GOOS but named it and its shims for unix, so on
  Windows it tripped the `.exe` suffix rule. Artifact name and shim platform are
  now both derived from the host.
- `TestPrepareScriptRuntimeStagesIncompleteReplacementWithoutBuildRoots`,
  `TestPrepareScriptRuntimeReusesOnlyCompleteManagedTree`,
  `TestPrepareSingleScriptStagesManagedBinWithoutCopyingSnapshotTree` — these
  declare `Platform: "unix"`, and staged unix commands are validated for a POSIX
  execute bit (`scripts.go:173`) that no file on a Windows host can carry. They
  are unconstructible there rather than merely unasserted, so they skip with the
  **existing** `platform-control` classification, matching
  `TestCompiledShimsStageWithoutLaunchThenPostInstallForwardExactly` in the same
  file. No new rows were added to `platform-cases.tsv` or `skip-classes.tsv`.

`internal/runtimestore` went from 5 Windows failures to 1.

## 3. STOP-THE-LINE — the remaining failure is a contradiction, not a bug

`TestWindowsPostInstallWrappersForwardArgumentsPathAndExitCode` still fails, and
it cannot be made to pass without a decision that is not mine.

### The constraint

The generated Windows launcher (`runtimestore.go:191`) is:

```
call "<runtimePath>" %*
exit /b %ERRORLEVEL%
```

`cmd.exe` runs a **second percent-expansion pass** over a `call` line after `%*`
has been substituted. Two consequences follow, and both are unavoidable:

1. A literal `%` in `runtimePath` pairs with a `%` coming from the forwarded
   arguments, and everything between them is deleted.
2. An argument containing `%VAR%` is expanded on that second pass. The shim
   itself defines `PATH` immediately above, so a forwarded `percent%PATH%value`
   can never survive verbatim.

Observed directly in CI. Run 30620739038 reports the command name as

```
'"C:\...\001\immutable cache PATHvalue"' is not recognized as an internal or external command
```

The fixture's artifact lives in `immutable cache % Юникод` and one forwarded
argument is `percent%PATH%value`. `immutable cache PATHvalue` is exactly those
two strings after the intervening percent-pairs were consumed.

### Why no escaping scheme fixes it

Quadrupling `%` in `escapeCMDValue` repairs the *path* side (pass 1 `%%%%`→`%%`,
pass 2 `%%`→`%`) but does nothing for the *argument* side: `%*` inserts the
arguments during pass 1, so `%PATH%` inside them is expanded by pass 2 regardless
of how the path is escaped. Verbatim `%VAR%` forwarding is unreachable while
`call` is on the line.

### Why `call` cannot simply be dropped

`call` is bound by the published protocol, not chosen by the implementation:

- `launcher_conformance_test.go:308` asserts `call "<runtimePath>" %*` for the
  `implementation_runtime` path role of the **authoritative launcher case**.
- `conformance_test.go:59` asserts the same literal as the "candidate forwarding
  contract".
- That conformance fixture's runtime path is
  `C:\manager home\runtime\launcher-skill\0123\scripts\launcher-tool.cmd` — a
  batch file. `call` is present precisely because the runtime target may be one;
  without `call`, a batch target never returns and `exit /b %ERRORLEVEL%` is
  unreachable.

### The contradiction

Three requirements are mutually unsatisfiable on `cmd.exe`:

1. the protocol-mandated launcher shape `call "<path>" %*`,
2. verbatim forwarding of an argument containing `%VAR%`,
3. a runtime path containing a literal `%`.

`targets_windows_test.go` asserts (2) and (3) together, and the ledger requires it
on windows with `skip_allowed_on = -`. The protocol requires (1). Adding
quadruple-escaping, delayed-expansion toggles, or per-argument special cases would
be compensating hacks around an invalid model, and the bug explicitly forbids
weakening the case to make the gate pass. Stopping here rather than building that
tower.

### Exact decision needed (human / architecture)

Pick one:

- **A — change the protocol launcher shape.** Replace `call "<path>" %*` with a
  form that does not re-expand (e.g. direct invocation), and decide what happens
  to batch-file runtime targets, which currently rely on `call` to return. This
  is a `curator-spec` change plus a `SPEC_PIN` re-pin.
- **B — declare runtime paths containing `%` unsupported on Windows.** Then the
  fixture directory `immutable cache % Юникод` is renamed. This narrows the
  ledgered Windows case, so it needs explicit approval.
- **C — declare verbatim `%VAR%` argument forwarding out of contract on Windows.**
  Then the fixture's `percent%PATH%value` argument is changed. Same caveat as B.

I can implement any of the three once chosen. I am not choosing between them,
because each changes either a published protocol contract or the meaning of an
acceptance case.

## 4. Out of scope — the Windows lane is broken repo-wide

`Test (windows-latest)` has never run in this workflow, so every Windows failure
in the repository is newly visible. **11 packages fail**, only one of which is
mine:

```
cmd/curator, internal/buildcache, internal/buildsource, internal/globalbins,
internal/godriver, internal/install, internal/install/atomicity,
internal/managerlock, internal/runtimestore, internal/staging,
internal/transaction
```

`internal/godriver` and `internal/transaction` belong to the concurrent sibling
scope BUG-260731-lepevi. The remaining eight have no owner. This is a
program-level finding and needs its own board items; it is not this bug.

## 5. Lint

`golangci-lint v2.12.2` (the version pinned in CI) reports **0 issues** against
this branch locally. The CI `Lint` job is red on exactly two findings —
`internal/godriver/controls_other.go:35:30` and
`internal/transaction/namespace.go:310:6`, both `unused`. They are **byte-identical
on base PR 9** (job 91108467255) and on this PR (job 91121004304), so this change
introduces neither, and both files are in the sibling's scope, which the
orchestrator context forbids me from touching. Checklist item 7 is therefore left
unchecked.

## 6. Gate evidence

Every command below was run as a standalone process; the exit code is the real one.

| gate | exit |
|---|---|
| `GOOS=windows go vet ./internal/runtimestore/` (before fix) | **1** — expected red, the reproduction |
| `GOOS=windows go vet ./internal/runtimestore/` (after) | 0 |
| `GOOS=windows go vet ./...` | 0 |
| `GOOS=linux go vet ./...` | 0 |
| `go vet ./...` (darwin) | 0 |
| `gofmt -l cmd internal` | empty |
| `golangci-lint run` (v2.12.2) | 0 issues |
| `ledger-consistency.sh` | 0 — 49 rows across linux darwin windows |
| `test-gate.sh` (darwin, SPEC_PIN root) | 0 — go test 0, platform-case gate 0 |
| `test-gate.sh` with `GO_TEST_FLAGS=-race` (darwin) | 0 |
| `gate-selftest.sh` | 0 — 75 passed, 0 failed |
| `no-broad-suppression.sh` | 0 |

Real CI on PR 10: `Test (windows-latest)` `go vet` **success** and `Ledger
consistency` **success**; `Gate self-test` green on all three runners; `Naming
gate` and `Interop conformance gate` green.

Windows execution could not be reproduced locally — this is a macOS host and no
Windows machine was available, so every Windows claim above comes from real CI
runs 30619686990 and 30620739038, whose evidence artifacts were downloaded and
parsed rather than read from a summary.
