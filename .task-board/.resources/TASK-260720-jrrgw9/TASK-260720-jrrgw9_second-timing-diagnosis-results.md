# TASK-260720-jrrgw9 — second `cmd/curator` timing diagnosis

Date: 2026-07-29  
Role: tester  
Verdict: development handback  
Scope: read-only diagnosis; no candidate, test, schema, golden, pin, configuration, or board acceptance claim was edited

## Finding

The verifier-2 failure is cumulative package duration, not a deadlock.

- The exact full gate failed with exit 1 at `cmd/curator 600.591s`.
- `TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands` had been active for only 1 second when the unchanged 10-minute package alarm fired.
- Its goroutine was runnable in `transaction.namespaceIdentity`/`validateJournal` during the first real `capture("install", "app")`.
- The two `capture` reader goroutines were in the expected `io.ReadAll` waits while `run` still owned the pipe writers; neither was a lock cycle.
- The same status/lifecycle/repair surface passed in the focused verifier-2 barrier at `cmd/curator 267.207s`.
- Three additional uncached narrow probes in this diagnosis all passed. The cache-movement test progressed through every subtest and returned normally.

The named legacy-shape test is therefore only the test that happened to start at the deadline. Optimizing or weakening that script-only compatibility assertion would not address the cause.

## Ranked timing inventory

| Rank | Surface | Observed time | Real work and duplication |
| ---: | --- | ---: | --- |
| 1 | `TestInstallAndUpgradeRepairCorruptCompiledState` | 85.75s | Two separately constructed compiled fixtures, four assertion-bearing corrupt-cache repairs. The two outer fixture/setup residuals were 11.02s and 11.23s after subtracting the named repair subtests. |
| 2 | `TestInstallAndUpgradeRestoreTheCacheWhenTheCommitFails` | 76.75s | Two separately constructed compiled fixtures; each case performs an assertion-bearing failed repair publication/reversal and a successful recovery repair. |
| 3 | `TestStatusReportsCompiledCurrentnessAndFailsCheck` | 69.816s in the producer's patched focused run | One real compiled fixture and 17 retained production plans. The prior patch already removed 27 duplicate plans and two cleanup rebuilds; no further assertion reduction is recommended here. |
| 4 | `TestStatusReportsProtectedCacheStateThatMovedDuringTheCheck` | 59.68s | One real compiled fixture plus four cleanup reconciliations. The cleanups alone cost 45.11s: removed 12.97s, unprovable boundary 12.81s, self-consistent replacement 6.32s, corrupt artifact 13.01s. They run after the moved-state assertions and prove no behavior owned by this test. |
| 5 | `TestInstallRepairsUntrustedCompiledStateAndPreservesTheOldInstall` | not isolated in this run | One separately constructed compiled fixture plus the assertion-bearing untrusted repair. Verifier evidence shows it was progressing normally at 28s in the earlier full timeout and it passes in the 267.207s focused group. Only its redundant initial fixture is a candidate for reuse. |
| 6 | `TestGCRetainsAndReportsReferencedCompiledState` | 13.30s | One separately constructed compiled fixture followed by the assertion-bearing GC and status check. Not needed for the minimum patch. |
| 7 | `TestStatusReportsAnUnusableToolchainPerCompiledCommand` | 11.10s | One separately constructed compiled fixture followed by three refused-toolchain report paths. Not needed for the minimum patch. |
| 8 | Authoritative lifecycle pair | 13.125s in the producer focused run | Real Git bootstrap/upgrade fixtures, no Go compiler artifacts. This is the authoritative rc.5 representative and should remain unchanged. |

Static graph findings:

- `global_status_test.go` is already in the desired shape: one compiled global fixture for the currentness matrix, one acquired production plan replayed for renderings, `snapshotBuildCacheAfter` for both cache-damage cases, and only one additional compiled fixture for the distinct transitive-closure topology.
- `lifecycle_conformance_test.go` uses temporary real Git upstream/working repositories and three script-only upgrade selections. It adds about 13 seconds and contains no duplicate Go builds.
- `capture` and `captureReport` drain both pipes concurrently and restore the process-global streams after `run`/`reportGlobalStatus` returns. They are not the timeout source and must not be parallelized.
- The three legacy project/global compatibility tests perform script-only installs. They are low-yield optimization targets and preserve distinct JSON, marker-schema, and unresolvable-closure contracts.

## Smallest robust patch

Patch only `cmd/curator/status_test.go`.

1. Add a small `compiledProjectFixture` value (`project`, `home`, `installed`) and a `newInstalledCompiledProject` helper that calls `compiledProject` and performs one real `capture("install", "app")`.
2. Add one sequential parent test named `TestCompiledProjectStatusRepairRollbackRecovery`.
3. Extract the existing bodies of these five tests into helpers that accept that installed fixture, and invoke them as named subtests of the parent:
   - `TestStatusReportsCompiledCurrentnessAndFailsCheck`
   - `TestInstallAndUpgradeRepairCorruptCompiledState`
   - `TestInstallAndUpgradeRestoreTheCacheWhenTheCommitFails`
   - `TestStatusReportsProtectedCacheStateThatMovedDuringTheCheck`
   - `TestInstallRepairsUntrustedCompiledStateAndPreservesTheOldInstall`
4. In the cache-movement helper, replace each `t.Cleanup(func() { reinstall(t) })` with `snapshotBuildCacheAfter(t, home)` registered before the cache mutation. Keep the marker-before/after byte assertion and reacquire the production dry-run plan per case.
5. Re-read the marker/fingerprints at the start of every repair/rollback subtest. Each successful repair already returns the shared fixture to a current state; the snapshot cleanup returns every cache byte and mode after each movement case. Do not share a stale baseline across cases.
6. Do not add `t.Parallel`, change `capture`, change production behavior, change timeouts, or replace the assertion-bearing repair/rollback/recovery executions.

The five current tests construct seven real installed compiled fixtures: one currentness, two corrupt-repair command fixtures, two rollback command fixtures, one cache-movement fixture, and one untrusted-repair fixture. The parent needs one. This removes six initial install/build cycles. Replacing the four cache-movement cleanup reconciliations removes three more rebuilds and one cache-hit reconciliation.

Measured cleanup saving is 45.11s. Six redundant initial compiled fixtures at the observed approximately 11.1s setup cost contribute about 66.6s. Snapshot/helper overhead is sub-second for this one-entry cache. Expected saving is therefore 108–112s, leaving at least 90 seconds of package margin under comparable concurrent load without deleting an assertion or changing a real lifecycle representative.

## Assertion-preservation matrix

| Existing surface | Preserved executable assertions | What is removed |
| --- | --- | --- |
| Compiled currentness | Full identity row, private-path exclusion, current check, all 14 drift state/cause/outcome/detail/skill-demotion assertions, production `Describe`, three representative plain CLI paths | No assertion-bearing call; only the fixture install is supplied by the parent |
| Corrupt-cache repair | Both `install` and `upgrade`; receipt and artifact corruption; pre-repair status; unusable-toolchain refusal with byte-identical cache/marker/install; real repair; repair notice; current check; stable key; one live entry | Two per-command initial fixture builds become the parent's one already-current fixture |
| Commit-failure rollback/recovery | Both `install` and `upgrade`; real repair publication before failure; exact quarantine/stage counts; cache, marker, launcher/adapters/runtime/ledger fingerprints; prior corrupt status; successful ordinary recovery; current check | Two per-command initial fixture builds |
| Cache moved during status | Corrupt, removed, unprovable-boundary, and self-consistent replacement cases; fresh production dry-run plan; unchanged marker bytes; `build-state-changed` row/detail/demotion/fail-closed assertions | Four post-assertion cleanup reconciliations; exact cache bytes/modes are restored by the already-established snapshot helper |
| Untrusted cache | Initial marker, compiler-free dry-run vocabulary, real untrusted rebuild, stable logical key, current status, later toolchain failure preserving marker and installed bytes | Its separate initial fixture build |

## Exact Go commands and exits

No Go commands overlapped. Total Go wall time was **248.150s**, below the five-minute cap.

1. Exit **0**, package 60.312s:

   ```text
   CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 go test -count=1 -v ./cmd/curator -run '^TestStatusReportsProtectedCacheStateThatMovedDuringTheCheck$'
   ```

2. Exit **0**, package 162.920s:

   ```text
   CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 go test -count=1 -v ./cmd/curator -run '^(TestInstallAndUpgradeRepairCorruptCompiledState|TestInstallAndUpgradeRestoreTheCacheWhenTheCommitFails)$'
   ```

3. Exit **0**, package 24.918s:

   ```text
   CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 go test -count=1 -v ./cmd/curator -run '^(TestStatusReportsAnUnusableToolchainPerCompiledCommand|TestGCRetainsAndReportsReferencedCompiledState)$'
   ```

The identical pre-probe process query exited 1 before each command because it found no matching Go/test process:

```text
pgrep -af '(^|/)(go|.*\.test)( |$)|go-build|cmd/curator'
```

One later read-only query briefly returned external PIDs 14418 and 14450 (exit 0); the immediate exact `ps -p 14418,14450 ...` check exited 1 because both were already terminal. No further Go command was started.

## Literal narrow producer allowlist

After applying the one-file patch, the producer may run exactly these Go commands, sequentially, with no `-timeout` token:

```text
CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 go test -count=1 -v ./cmd/curator -run '^TestCompiledProjectStatusRepairRollbackRecovery$'
```

```text
CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 go test -count=1 -v ./cmd/curator -run '^(TestAuthoritativeBootstrapCasesAreExecutable|TestAuthoritativeUpgradeCasesAreExecutable|TestGlobalStatusReportsCompiledCurrentnessAndFailsCheck|TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands|TestStatusAcceptsAnUnchangedLegacyMarkerSchema)$'
```

Also run `gofmt` on exactly `cmd/curator/status_test.go` and `git diff --check -- cmd/curator/status_test.go`.

Forbidden in the producer pass: broad `go test ./cmd/curator` without `-run`, `go test ./...`, race, coverage, Windows execution, any timeout flag, overlapping Go commands, production edits, cache clearing, host installation, stage, commit, publication, or pin changes.

## Handback

Return the task to development for the one-file producer patch. This diagnosis does not claim repository conformance acceptance. After producer rework, an independent verifier must establish the exact uncached full-suite result before full race or native Windows gates are attempted.
