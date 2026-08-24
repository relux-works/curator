# TASK-260824-1kzj22 — developer outcome: parallelize the cmd/curator test suite

Role: developer. Change is **test code only**; no production file was touched.

## Summary

36 of the 70 `cmd/curator` test functions were genuinely independent and are now
marked `t.Parallel()`. They hold **1.3 %** of the package's wall-clock. The other
33 hold **98.7 %** and are serial *by construction*, not by omission: every one
of them drives the CLI through `run()`, which resolves the manager configuration
from the process-global `CURATOR_CONFIG` environment variable and writes through
the process-global `os.Stdout` / `os.Stderr`.

**The 4-minute acceptance target is not reachable with `t.Parallel()` alone, and
this is provable without reference to machine speed:** the single test
`TestCompiledProjectStatusRepairRollbackRecovery` measures 230–268 s on its own,
above the 240 s target, and its five subtests corrupt and repair one shared
installed fixture, so they cannot be split either.

Per the brief ("report how close it got and what the residual serial bottleneck
is — do NOT silently exceed scope into sub-packaging without saying so"), the
work stops at the safe parallelization and the bottleneck is documented below
with the options that would actually move it. Each of those options needs a
decision outside this task's scope.

## Measured before / after

Reference host: darwin/arm64, 16 CPUs, go1.25.5.
Command: `go test -timeout 30m -count=1 ./cmd/curator`.

**Caveat that matters for every wall-clock number here:** during the measurement
window other agent sessions on this host were running their own `go test` passes
(up to four foreign `curator.test` / `install.test` processes, 1-minute load
average between 11 and 51). Load is recorded per run. The absolute numbers are
therefore contention-inflated; the *ratio* between the parallelizable and serial
share is not, because it is derived from per-test times inside a single run.

| Run | Tree | Wall-clock | Exit | 1-min load at start |
| --- | --- | ---: | ---: | ---: |
| baseline 1 | HEAD (no `t.Parallel()`) | 326.6 s | 0 | ~15 |
| baseline 2 | HEAD | 305.6 s | 0 | ~15 |
| baseline 3 | HEAD | 522.5 s | 0 | ~15 |
| after 1 | patched | 543.7 s | 0 | 14.50 |
| after 2 | patched | 374.0 s | 0 | 15.17 |
| after 3 | patched | 470.3 s | 0 | 22.87 |

An earlier attempt at run 3 was killed by the tooling's own 10-minute call cap,
not by a test failure; it was re-run to completion in the background and that
completed run is the 470.3 s row. A fourth patched run, the verbose one used for
the per-test breakdown, took 383.8 s and also exited 0.

Baseline median 326.6 s, after median 470.3 s — the spread between runs of the
*same* tree (305.6 s … 522.5 s) is larger than any effect this change can have,
which is itself the honest headline: at 98.7 % serial share there is nothing for
`t.Parallel()` to overlap.

The noise-resistant measurement, taken from one verbose run
(`.temp/TASK-260824-1kzj22/after-run1-verbose.log`):

| Group | Cases | Cost if serial | Cost when parallel |
| --- | ---: | ---: | ---: |
| parallelized | 36 | 5.18 s | 2.52 s (batch floor = slowest member) |
| serial by construction | 33 | 381.03 s | 381.03 s |

Isolated confirmation on the parallelized subset only, 3 runs each:

| `-parallel` | run 1 | run 2 | run 3 |
| ---: | ---: | ---: | ---: |
| 1 | 5101 ms | 4823 ms | 5009 ms |
| 16 | 4215 ms | 3860 ms | 4129 ms |

So the change is real and works — it just cannot matter at the package level.

Coverage is provably untouched: the two profiles agree block for block
(695 / 1218 statements, 778 blocks, zero differing blocks), which is what you
would expect when the only edits are `t.Parallel()` calls and a doc comment.

### Where the time actually is

Top serial costs from the same verbose run:

| Test | Time |
| --- | ---: |
| `TestCompiledProjectStatusRepairRollbackRecovery` | 230.63 s |
| `TestGlobalStatusReportsCompiledCurrentnessAndFailsCheck` | 33.90 s |
| `TestStatusReportsATransitivelyResolvedCompiledCommand` | 22.73 s |
| `TestGlobalStatusReportsATransitivelyResolvedCompiledCommand` | 19.76 s |
| `TestCLICompatibleVerifiedProviderOwnsBuildDispatchAndReceipt` | 13.95 s |
| `TestGCRetainsAndReportsReferencedCompiledState` | 13.41 s |

`TestCompiledProjectStatusRepairRollbackRecovery` breaks down (earlier
low-contention run, `.temp/TASK-260824-1kzj22/baseline-chunkAC.log`) into five
subtests of 80.85 s, 72.11 s, 64.95 s, 20.07 s and 17.02 s.

That cost is not Go test overhead — it is cold Go compilation. `godriver` gives
every build session a hermetic `GOPATH` / `GOMODCACHE` / `GOCACHE` under its own
root (`internal/godriver/session.go:411`), and there is an explicit guard case
rejecting a shared cache (`internal/godriver/guards_test.go:60`). Each compiled
command fixture therefore pays a full stdlib compile.

## Isolation review

### Parallelized (36) — why each is safe

All 36 satisfy the same four conditions: no `t.Setenv` / `os.Setenv` (directly or
through a helper), no `capture()` and no other `os.Stdout` / `os.Stderr` swap, no
`os.Chdir`, and every filesystem fixture is a fresh `t.TempDir()`. None of them
mutates the one package-level seam in this package, `resolveCLIProvider`
(`cmd/curator/assurance.go:17`) — the two cases that do swap it are both in the
serial set. No test in this package binds a port.

By file:

- `builds_test.go` (21): `TestClassifyBuildCommandMapsEveryCacheOutcomeToADistinctCode`,
  `TestUnusableToolchainRowsCarryTheDriverBoundaryCode`,
  `TestClassifyBuildCommandDetectsRecordedIdentityDrift`,
  `TestBuildInputDriftIsAttributedOnlyAsFarAsTheMarkerProves`,
  `TestInputCausesAreDistinctAndDocumented`,
  `TestClassifySkillBuildsAcceptsOnlyAnExactlyCurrentInstallation`,
  `TestClassifySkillBuildsWithoutCompiledStateStaysSilent`,
  `TestClassifySkillBuildsDetectsContextExposure`,
  `TestClassifySkillBuildsDetectsCommandDriftInBothDirections`,
  `TestClassifySkillBuildsRefusesAMarkerThatCannotDescribeABuild`,
  `TestMarkerRefusalSeparatesUnsupportedFromInvalid`,
  `TestCurrentnessCodesAreDistinctAndOnlyExactStatesPass`,
  `TestEveryCurrentnessCodeIsDocumented`, `TestCheckFailsForEveryNonCurrentCode`,
  `TestBuildReportsNeverPublishAnAbsolutePath`, `TestBuildReportDetailIsBounded`,
  `TestGoToolchainGuidanceNamesTheAcceptedSelectionAndTestedFamilies`,
  `TestRepairNoticesDistinguishRepairFromAPreservedInstallation`,
  `TestMarkerDigestsDetectAConcurrentInstallMarkerChange`,
  `TestStatusReportMarksCompiledStateThatMovedDuringTheCheck`,
  `TestStatusReportReportsCompiledCommandsOfAnUninstalledSkill`.
  These call the classification functions directly with in-memory markers and
  `t.TempDir()` roots. `TestEveryCurrentnessCodeIsDocumented` and
  `TestInputCausesAreDistinctAndDocumented` additionally read
  `../../README.md` — read-only, so concurrent reads are safe.
- `main_test.go` (14): `TestProductionExternalDepsBindTrustedGitAndAudit`,
  `TestProductionExternalAuditReturnsAdvisoryWarnings`,
  `TestUsageEnumeratesDocumentedCommands`, `TestRunVersionExitsZero`,
  `TestRunNoArgsPrintsUsage`, `TestRunUnknownCommand`, `TestShellInitPrintsHooks`,
  `TestSkillCheckOnTempDir`, `TestInstallFlagsAcceptTrailingOptions`,
  `TestInstallAuditFlagAcceptsOptionalMode`,
  `TestSelectProjectTargetsUsesAliasesAndStableAllOrder`,
  `TestStatusDriftDetectsContentTampering`,
  `TestHiddenWorkerModeIsNotAUserVisibleCommand`,
  `TestProductionBinaryDispatchesRustOracleBeforeAmbientCargoDiscovery`.
- `toolchain_host_test.go` (1): `TestHostGoToolchainIsSelectableOnAnInventoryPlatform`
  — a read-only `godriver.Probe` against a `t.TempDir()` config root.

Six of these (`TestRunVersionExitsZero`, `TestRunNoArgsPrintsUsage`,
`TestRunUnknownCommand`, `TestShellInitPrintsHooks`, `TestSkillCheckOnTempDir`,
plus `TestHiddenWorkerModeIsNotAUserVisibleCommand`) call `run()` and let it
write to the *real* `os.Stdout`. That is safe for a specific reason worth stating
because it is the load-bearing assumption of this whole change: Go releases
paused parallel tests only after the entire serial pass of the package has
finished, so a parallel case can never overlap a serial case that has swapped
the streams via `capture()`. The invariant is now documented in the `capture`
doc comment (`cmd/curator/status_test.go`) so the next person who adds a case
does not have to rediscover it.

### Left serial (33) — why

| Test | Reason | File |
| --- | --- | --- |
| `TestAuthoritativeBootstrapCasesAreExecutable` | CURATOR_CONFIG env + os.Stdout swap | lifecycle_conformance_test.go |
| `TestAuthoritativeUpgradeCasesAreExecutable` | CURATOR_CONFIG env + os.Stdout swap | lifecycle_conformance_test.go |
| `TestBootstrapAndProjectCommands` | CURATOR_CONFIG env | main_test.go |
| `TestBootstrapIfMissingCreatesAbsentConfig` | CURATOR_CONFIG env | main_test.go |
| `TestBootstrapIfMissingKeepsExistingConfigWithoutParsing` | CURATOR_CONFIG env | main_test.go |
| `TestBootstrapIfMissingRejectsForce` | CURATOR_CONFIG env | main_test.go |
| `TestCLICompatibleVerifiedProviderOwnsBuildDispatchAndReceipt` | CURATOR_CONFIG env + os.Stdout swap + swaps `resolveCLIProvider` | main_test.go |
| `TestCLIEndToEndInstallStatusAndTamperCheck` | CURATOR_CONFIG env | main_test.go |
| `TestCLIExecutionAssuranceSelectionIsPortableDefaultAndVerifiedFailClosed` | CURATOR_CONFIG env | main_test.go |
| `TestCLIVerifiedCapabilityDriftStartsNothingAndAdoptsNoCache` | CURATOR_CONFIG env + os.Stdout swap + swaps `resolveCLIProvider` | main_test.go |
| `TestCompiledInstallFollowsTheNativeControlInventoryExactly` | CURATOR_CONFIG env + os.Stdout swap | status_test.go |
| `TestCompiledProjectStatusRepairRollbackRecovery` | CURATOR_CONFIG env + os.Stdout swap; 5 subtests share one installed fixture they corrupt and repair | status_test.go |
| `TestDryRunNeverClaimsACompletedCompilerCheck` | CURATOR_CONFIG env + os.Stdout swap | status_test.go |
| `TestGCPrunesDeadConsumersUnderTheHomeLock` | CURATOR_CONFIG env | gc_test.go |
| `TestGCRetainsAndReportsReferencedCompiledState` | CURATOR_CONFIG env + os.Stdout swap | status_test.go |
| `TestGCRunsSerializedAcrossConcurrentInvocations` | CURATOR_CONFIG env | gc_test.go |
| `TestGCWaitsForTheHomeLock` | CURATOR_CONFIG env | gc_test.go |
| `TestGlobalStatusFailsCheckWhenTheClosureCannotBeProven` | CURATOR_CONFIG env + os.Stdout swap | global_status_test.go |
| `TestGlobalStatusFailsCheckWhenTheUserHomeCannotBeResolved` | mutates `HOME`/`USERPROFILE`/`HOMEDRIVE`/`HOMEPATH` + os.Stdout swap | global_status_test.go |
| `TestGlobalStatusKeepsTheDeclaredSkillSurfaceWithoutCompiledCommands` | CURATOR_CONFIG + HOME env + os.Stdout swap | global_status_test.go |
| `TestGlobalStatusRejectsPositionalArguments` | CURATOR_CONFIG + HOME env + os.Stdout swap | global_status_test.go |
| `TestGlobalStatusReportsATransitivelyResolvedCompiledCommand` | CURATOR_CONFIG + HOME env + os.Stdout swap | global_status_test.go |
| `TestGlobalStatusReportsCompiledCurrentnessAndFailsCheck` | CURATOR_CONFIG + HOME env + os.Stdout swap | global_status_test.go |
| `TestGlobalStatusWithoutASkillfileStaysSilentAndCurrent` | CURATOR_CONFIG + HOME env + os.Stdout swap | global_status_test.go |
| `TestGlobalUpgradeDryRunDoesNotCreateSkillsRoot` | CURATOR_CONFIG env | main_test.go |
| `TestShellInitInstallCachesHookWithoutConfig` | mutates `CURATOR_CONFIG` and `SHELL` | main_test.go |
| `TestStatusAcceptsAnUnchangedLegacyMarkerSchema` | CURATOR_CONFIG env + os.Stdout swap | status_test.go |
| `TestStatusExplainsAnUnusableGoToolchain` | CURATOR_CONFIG + `CURATOR_GO` selection env + os.Stdout swap | status_test.go |
| `TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands` | CURATOR_CONFIG env + os.Stdout swap | status_test.go |
| `TestStatusKeepsLegacyBehaviourWhenAPlanFailsWithoutCompiledCommands` | CURATOR_CONFIG env + os.Stdout swap | status_test.go |
| `TestStatusReportsATransitivelyResolvedCompiledCommand` | CURATOR_CONFIG env + os.Stdout swap | status_test.go |
| `TestStatusReportsAnUnusableToolchainPerCompiledCommand` | CURATOR_CONFIG + selection env + os.Stdout swap | status_test.go |
| `TestUpgradeDryRunDoesNotCreateOrFetchSkillsRoot` | CURATOR_CONFIG env | main_test.go |

Two independent Go rules make the env column fatal rather than merely risky:
`t.Setenv` panics if the test or any ancestor is parallel, and `t.Parallel()`
panics if that test already called `t.Setenv`. There is no ordering that satisfies
both.

Subtests were deliberately **not** parallelized. In the serial cases they either
inherit the same two couplings or share a mutable fixture with their siblings
(`TestCompiledProjectStatusRepairRollbackRecovery` is the clearest: each subtest
corrupts the installation and the next one relies on the repaired state). In the
parallelized cases every subtest table is sub-second pure computation, where
subtest scheduling costs more than it saves.

## Residual bottleneck and the options that would move it

The bottleneck is one sentence: **`cmd/curator`'s CLI surface is only reachable
through process-global state, so its end-to-end cases cannot share a process.**

`run()` (`cmd/curator/main.go:94`) takes only `args`; it reaches configuration via
`loadConfig()` → `config.Load("")`, which reads `CURATOR_CONFIG` from the
environment, and it emits through `fmt.Print*` to `os.Stdout` / `os.Stderr`.

| Option | Reaches ≤ 4 min? | Cost / blocker |
| --- | --- | --- |
| A. Make the config path and output streams injectable (`run(args, stdout, stderr, configPath)`), tests stop owning process globals | Yes | Production change — explicitly forbidden by this task's scope. Needs an architecture decision. |
| B. Drive the CLI as a subprocess from the tests, reusing the existing `TestMain` worker-mode re-exec | Yes, no production change | In-package coverage attribution for `cmd/curator` collapses unless the suite moves to `GOCOVERDIR`-based subprocess coverage. The AC requires unchanged coverage, so this needs the coverage-mode decision first. Also a large rewrite of ~5 000 lines of test code. |
| C. Let the compiled-command fixtures share a Go build cache | Yes, and it attacks the real cost directly | Contradicts the hermetic build isolation `internal/godriver/guards_test.go:60` exists to enforce. Security/design decision, not a test change. |
| D. Split the compiled-command end-to-end cases into their own package so `go test ./...` runs them in a separate process | Reduces `./...` wall-clock, not `./cmd/curator` | Requires `run()` to be importable, i.e. a production change. Also does not satisfy the AC as written, which measures `./cmd/curator`. |

Recommendation: **A**. It is the smallest change that removes both couplings at
once, it is ordinary Go CLI design, and it leaves coverage attribution untouched.
B is the fallback if production code must stay frozen, but it should not be
started until someone accepts the coverage-mode change.

## Validation

| Gate | Command | Result |
| --- | --- | --- |
| gofmt | `gofmt -l cmd/curator/` | exit 0, no output |
| vet | `go vet ./cmd/curator/` | exit 0 |
| lint | `golangci-lint run ./cmd/curator/...` (pinned v2.12.2) | exit 0, `0 issues.` |
| whitespace | `git diff --check` | exit 0 |
| race, parallelized subset | `go test -race -timeout 30m -count=1 -run '^(<36 cases>)$' ./cmd/curator` | exit 0, 36 PASS, no `DATA RACE` |
| full suite | `go test -timeout 30m -count=1 ./cmd/curator` | 3 consecutive runs, all exit 0; see the run table above |
| coverage, HEAD | `go test -timeout 30m -count=1 -coverprofile=... ./cmd/curator` | exit 0, `coverage: 57.1% of statements` (388.7 s) |
| coverage, patched | same | exit 0, `coverage: 57.1% of statements` (521.7 s) |

Logs are in `.temp/TASK-260824-1kzj22/` and attached to the board:
`before-baseline-run{1,2,3}.log`, `after-run1-verbose.log`, `after-run{1,2,3}.log`,
`race-parallel-subset.log`, `lint.log`, `baseline-chunk{AC,DP,QZ}.log`.

## Acceptance criteria status

- [x] Independent cases marked `t.Parallel()` with correct per-test isolation; test code only, no production change.
- [x] Isolation review documented (both directions, above).
- [ ] **Wall-clock ≤ 4 min — not met, and not reachable in scope.** Closest observed full run on the patched tree: 374.0 s under 1-min load 15.2
  (three consecutive completed runs: 543.7 s, 374.0 s, 470.3 s, all exit 0). The residual bottleneck is named above; the single heaviest test alone is 230–268 s.
- [x] Zero new flaky failures — three consecutive `-count=1` runs on the patched tree exited 0 (plus a fourth verbose run), and all three baseline runs exited 0; no test changed its outcome, no test was skipped that was not already skipped at HEAD.
- [x] Coverage unchanged — verified, not asserted. Both trees report `coverage: 57.1% of statements`, and a block-level diff of the two profiles is exact: 778 blocks on each side, **695 / 1218 statements covered on both**, zero blocks whose covered state differs, no block present in one profile and not the other. Profiles are in the evidence bundle (`cover-before.out`, `cover-after.out`).
- [x] Focused `-race` run on the parallelized subset green.
- [x] gofmt, `go vet`, pinned golangci-lint v2.12.2 clean; `git diff --check` clean.
- [x] Evidence logs and this outcome attached.
