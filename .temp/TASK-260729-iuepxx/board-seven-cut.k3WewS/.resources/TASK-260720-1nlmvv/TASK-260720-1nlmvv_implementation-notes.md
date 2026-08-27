# TASK-260720-1nlmvv — Expose build diagnostics and repair behavior

Worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1nlmvv/worktree`
(base `origin/main` 17804ce; nothing staged, committed, or published).
Platform: darwin/arm64, Go 1.25.5, golangci-lint v2.4.0 (pinned).

## Provenance of the composed base

The story's `done` blockers landed as uncommitted working trees, not commits.
The accepted tree of the last one, `TASK-260720-1ljev5` (collect-build-cache-safely),
lives in `.temp/TASK-260720-1ljev5/worktree` and already contains the accepted
state of every earlier phase. It was copied verbatim into this task's worktree
(`rsync -a --delete`, verified by `diff -r`, predecessor left untouched) and
verified green before any edit: `go build ./...` exit 0.

Task-only delta, verified by directory diff against the composed base:

| File | Change |
|---|---|
| `cmd/curator/builds.go` | new — stable currentness vocabulary, classification, redaction-aware rendering, repair notices, toolchain guidance |
| `cmd/curator/main.go` | `status` reports compiled state; `--check` fails closed; install/upgrade print repair notices; legacy marker schema accepted again |
| `internal/install/diagnostics.go` | new — `RedactDiagnostic`, the one bounded/redacted rendering of untrusted detail |
| `internal/install/plan.go` | `Driver()`, `ReceiptSHA256()`, `Artifact()` accessors; `Describe()` redacts its reason |
| `internal/install/install.go`, `global.go` | `Result.BuildDiagnostic` carries the stable go-v1 boundary code |
| `internal/godriver/session.go` | `SelectionCuratorGo`, `SelectionGOROOT`, `TestedFamilies()` — the accepted selection mechanisms and tested families as data |
| `README.md` | documents every code, the dry-run vocabulary, the repair path, and toolchain selection |

Plus three new test files (`cmd/curator/builds_test.go`, `cmd/curator/status_test.go`,
`internal/install/diagnostics_test.go`).

## Design

### The currentness vocabulary is one closed set

`currentnessCodes()` enumerates every code `status` can produce; `currentCode()`
admits exactly two of them (`up-to-date` for a skill, `current` for a compiled
command). `checkFailed()` is the whole `--check` verdict and consults both the
declared-skill map and the compiled-command rows, because a build can belong to
a transitively resolved node that no project declaration names. Everything —
invalid, unsupported, missing, corrupt, drifted, moved, unknown — fails closed.
`TestCheckFailsForEveryNonCurrentCode` walks the full set on both surfaces, and
`TestEveryCurrentnessCodeIsDocumented` keeps README and code in sync.

### Classification compares identities first, then the cache

`classifyBuildCommand` checks the recorded driver, then the recorded
build-source identity, then the recorded logical key, and only then interprets
the planner's read-only cache inspection. That ordering is what makes "wrong
target or toolchain" reportable as exactly that: if the source identity still
matches but the derived key does not, the only remaining inputs are target,
toolchain, and the fixed policy.

A `cache-hit` is not accepted on its own. The entry's receipt identity, artifact
path, and artifact hash must equal the ones the marker recorded, which is how
`corrupt-build-receipt` and `build-artifact-drift` separate from a plain hit.
The planner now exposes those two values (`ReceiptSHA256()`, `Artifact()`), set
only for a hit, so a refused entry can never supply evidence about itself.

### Compiled state can only demote, never promote

`statusReport` computes the historical drift map first and lets compiled state
lower a skill only from `up-to-date`. A skill that is missing, tampered,
unresolvable, or behind its declaration keeps that more actionable code. This is
also the legacy-compatibility guarantee: the human lines and the `skills` object
are unchanged for every installation without compiled commands, and the `builds`
key is absent entirely rather than empty.

### Concurrent change closes the whole window

Every install marker below the project and hybrid stores is fingerprinted before
the read-only plan runs and again after the last classification. A marker that
differs between the two makes every verdict for that skill `build-state-changed`
rather than an authoritative-looking result derived from a plan that was already
stale. `TestStatusReportMarksCompiledStateThatMovedDuringTheCheck` drives that
path directly; `markerDigests`/`markerMoved` are pinned separately.

### One redaction rule, applied at both surfaces

`install.RedactDiagnostic` is the single rendering of untrusted detail — cache
reasons, receipt bytes, compiler output. It drops invalid UTF-8, replaces every
non-printable rune (control characters, escape sequences, line breaks,
zero-width format runes) with a space, replaces every Unix, Windows, and UNC
absolute path with `<path>`, and bounds the result to 240 runes. Protocol-relative
declaration paths such as a build root survive, because they are declaration
state rather than private location state. It is applied by `PlannedBuild.Describe()`
so the installation report is covered too, not only `status`.

### Toolchain guidance is generated from the driver's own facts

`godriver` now exports `SelectionCuratorGo`, `SelectionGOROOT`, and
`TestedFamilies()`; `ConfigFromEnvironment` reads the same constants, so there is
one source of truth. `Result.BuildDiagnostic` carries the stable `go-v1` code out
of a failed plan or staging pass, and `goToolchainGuidance` turns it into text
that names both selection mechanisms and every tested release family. It never
mentions `PATH` or a download outside the one sentence that denies them — the
test strips that sentence and then forbids both words in the remainder.

### Repair is install and upgrade, and it says which one happened

There is no new repair command; the landed rc.4 CLI contract explicitly does not
prescribe one, and install/upgrade already rebuild untrusted, corrupt,
wrong-target, and missing entries after every gate. `printRepairNotices`
distinguishes "rebuilt untrusted build cache state into a new protected entry"
from "did not repair … the previous installation, its consumers, and the live
build cache are unchanged", and from a refusal "before any mutation" for a
blocking outcome. Dry runs print no notice at all.

### Legacy marker schema fix

`scopeStatusDrift` reported `unsupported-marker` for every marker below the
current schema, so since marker v2 landed an unchanged schema 1 through 5
installation was misreported and `status --check` failed for it. The check now
accepts `marker.LegacySchemaVersion` as well; a schema-1 marker that would have
to describe a compiled command is demoted to `needs-install` by the build
classification, which is the only place that can see the plan. Pinned by
`TestStatusAcceptsAnUnchangedLegacyMarkerSchema` and
`TestClassifySkillBuildsRefusesAMarkerThatCannotDescribeABuild`.

## Evidence

Reported exit codes are the real status of each command, run standalone.

### macOS (darwin/arm64), primary

| Gate | Exit |
|---|---|
| `gofmt -l .` | 0 (no output) |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `go test ./... -count=1` (40 packages) | 0 |
| `golangci-lint run ./...` (pinned v2.4.0) | 0, **0 issues** |
| `go test -race -timeout 40m ./internal/install -count=1` | 0 (709s) |
| `go test -race -timeout 20m ./cmd/curator ./internal/godriver -count=1` | 0 (181s + 44s) |
| `GOOS=windows GOARCH=amd64 go build ./...` | 0 |
| `GOOS=linux GOARCH=amd64 go build ./...` | 0 |
| `GOOS=linux GOARCH=amd64 go vet ./...` | 0 |

### Expected-red gate, honestly reported

`GOOS=windows GOARCH=amd64 go vet ./...` exits **1**:

```
vet: internal/runtimestore/targets_windows_test.go:97:14: undefined: decodeHelperOutput
```

This is **pre-existing and sibling-owned** (`internal/runtimestore`,
`TASK-260720-29hi1h`). The identical command on the untouched accepted base
(`.temp/TASK-260720-1ljev5/worktree`) also exits 1 with the identical message.
Excluding that one package, `GOOS=windows GOARCH=amd64 go vet` over this tree
exits **0**.

### Native Windows

Host `win` (Windows 10.0.19045.6456, `DESKTOP-3PBO632\admin`). The host has
**no Go toolchain and no `git`**, so the test binaries were cross-compiled on
macOS and executed natively. Binary provenance was verified before the run
(`strings` confirms this task's strings are present).

| Suite | Exit | Result |
|---|---|---|
| `install.test.exe -test.run "TestRedactDiagnostic\|TestPlanLinesRedact"` | **0** | all 5 tests and 6 subtests pass, including the native Windows-path and UNC redaction cases |
| `curator.test.exe -test.run "<the 16 classification, vocabulary, guidance, repair and marker-digest tests>"` | **0** | every platform-neutral test of this task passes natively |
| `curator.test.exe` (whole package) | **1** | 8 failures, all `exec: "git": executable file not found in %PATH%` |

The whole-package failure is an environment limit, not a product defect: the
eight tests are the pre-existing `TestCLIEndToEndInstallStatusAndTamperCheck`
plus the seven CLI tests of this task, and every one of them builds a git
repository fixture (the compiled ones additionally need a real Go toolchain).
Measured, not assumed: the **accepted base** binary
(`.temp/TASK-260720-1ljev5/worktree`, cross-compiled and run on the same host in
the same session) fails `TestCLIEndToEndInstallStatusAndTamperCheck` with the
identical message and exit 1.

`TestEveryCurrentnessCodeIsDocumented` skips with an explicit reason when the
checkout is not reachable from the test binary, which is exactly the
standalone-binary case; it runs and passes in any real checkout.

### Not run

Native Linux execution was not run. The change is platform-neutral presentation
code; `GOOS=linux GOARCH=amd64 go build` and `go vet` are both clean, and the
only platform-specific behaviour it touches — absolute-path redaction for Unix,
Windows and UNC forms — is covered by dedicated cases that pass on both macOS
and native Windows.

## Acceptance criteria

| Criterion | Where |
|---|---|
| Dry-run miss says `would-preflight-and-build`, never a completed compiler check | `TestDryRunNeverClaimsACompletedCompilerCheck` (install and upgrade) |
| Untrusted protected state says `would-rebuild-untrusted-cache` | `TestInstallRepairsUntrustedCompiledStateAndPreservesTheOldInstall` |
| Cache hits and every non-current class are distinct in JSON and human output | `TestClassifyBuildCommandMapsEveryCacheOutcomeToADistinctCode`, `TestClassifyBuildCommandDetectsRecordedIdentityDrift`, `TestStatusReportsCompiledCurrentnessAndFailsCheck` (5 real end-to-end classes, JSON and human) |
| `status --check` non-zero for every invalid/unsupported/missing/corrupt/unknown state, zero only for exact current | `TestCheckFailsForEveryNonCurrentCode`, `TestStatusReportsCompiledCurrentnessAndFailsCheck`, `TestStatusExplainsAnUnusableGoToolchain` |
| Go diagnostics name the accepted selection mechanisms and tested family, without PATH or download | `TestGoToolchainGuidanceNamesTheAcceptedSelectionAndTestedFamilies`, `TestStatusExplainsAnUnusableGoToolchain` |
| Install and upgrade repair invalid cache/marker state only after gates, preserving the old install on failure | `TestInstallRepairsUntrustedCompiledStateAndPreservesTheOldInstall` (real rebuild, unchanged logical key, then a failing run that leaves the previous installation byte-intact) |
| Compiler diagnostics bounded and redacted | `TestRedactDiagnostic*` (4), `TestBuildReportsNeverPublishAnAbsolutePath`, `TestBuildReportDetailIsBounded`, `TestPlanLinesRedactAnUntrustedReason` |
| Legacy status output and CLI exit codes compatible | `TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands`, `TestStatusAcceptsAnUnchangedLegacyMarkerSchema`, `TestStatusDriftDetectsContentTampering`, `TestCLIEndToEndInstallStatusAndTamperCheck` (both unchanged and passing) |
| `gc` behaviour for compiled state | `TestGCRetainsAndReportsReferencedCompiledState`, plus the existing `TestGC*` set unchanged |

`cmd/curator` gained a `TestMain` that dispatches the fixed hidden worker mode
exactly as the installed manager does, so the CLI tests above run **real**
compiled installations through the identity-verified process graph rather than a
mock. `TestHiddenWorkerModeIsNotAUserVisibleCommand` still passes: the dispatch
is on `os.Args`, not on the parsed command surface.

## Findings worth carrying forward

1. **Legacy compatibility regression, fixed here.** Since marker v2 landed,
   `status` reported `unsupported-marker` for every schema-1 marker, so an
   unchanged schema 1 through 5 installation failed `status --check`. Fixed and
   pinned; worth a note in the rc.4 conformance task.
2. **A repeated install of a compiled skill is not idempotent.** `stageNode`
   calls `marker.Current(live, expected)` without a `marker.BuildCurrentness`
   value, and that call fails closed for any marker carrying build state, so the
   context directory is re-staged on every run (verified: a second `curator
   install app` over an unchanged compiled project reports `installed`, not
   `up-to-date`, while the plan line still reports `outcome=cache-hit`, so
   nothing recompiles). Safe but wasteful, and it is install-staging behaviour
   owned by `TASK-260720-3itlly`/`TASK-260720-2284br`, not by this task. Worth
   its own task: wire `BuildCurrentness` into `stageNode`.
3. **`curator global status` is not build-aware.** It renders `scopeStatusDrift`
   directly and has no `--json` or `--check`. The global scope does plan and
   commit compiled commands, so a compiled global installation currently has no
   currentness surface. Out of this task's listed CLI surface (install, upgrade,
   dry-run, status, `status --json`, `status --check`, gc); worth its own task.
4. **`internal/install` exceeds the default `go test` timeout under `-race`.**
   It needs `-timeout 40m` (709s observed); the default 10m produces a timeout
   panic that reads like a failure but reports no data race. CI should pin the
   timeout — relevant to `TASK-260720-1pvfj5`.
5. **The Windows validation host still has neither `git` nor a Go toolchain.**
   Every CLI-level test that builds a repository fixture is therefore
   unrunnable there, including one that already existed. Installing `git` on
   that host would recover eight tests' worth of native coverage — relevant to
   `TASK-260720-1pvfj5`.
6. **One flake observed, not reproducible.** `TestRegistryRevocationDeniesInstall`
   failed once during a loaded parallel `go test ./...` with "registry test-reg
   snapshot timestamp is too far in the future"; it passed in isolation, in a
   full `./internal/install` run, on the untouched base, and in two later full
   `./...` runs (exit 0 each). The registry snapshot clock-skew tolerance looks
   timing-sensitive under load. Worth its own task rather than a silent retry.
