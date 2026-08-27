# TASK-260720-jrrgw9 cmd/curator timing diagnosis

Date: 2026-07-29
Role: tester
Disposition: development handback; diagnosis only

## Scope and provenance

Audited the exact candidate at `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-jrrgw9/worktree`, the accepted integrated comparison at `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-2kaopg/worktree`, `TASK-260720-jrrgw9_full-verifier-results.md`, the raw failed full-suite log, accepted baseline timing artifacts, and the exact 22-file task delta. `rsync -nrc` against the accepted worktree exited 0 and identified 20 added tests plus the two modified tests `internal/buildcache/conformance_test.go` and `internal/closure/conformance_test.go`; no product delta belongs to this task. The two `git diff --no-index` inspections exited 1 as expected because those two files differ; these were read-only diff inspections, not green gates.

## Diagnosis

The full repository command exited 1 after 601 seconds because the `cmd/curator` package reached its unchanged 10-minute package deadline at 600.435 seconds. `TestInstallRepairsUntrustedCompiledStateAndPreservesTheOldInstall` was merely the test active when the package alarm fired and had run for 28 seconds. Historical measurement of that test is 34.69 seconds, so the stack does not attribute the cumulative overrun to it.

The accepted comparison already ran close to the limit: two accepted high-load repository gates measured `cmd/curator` at 554.967 and 545.195 seconds, while the accepted comparable-load clean derivation was 465.944 seconds. This task introduces only one `cmd/curator` file, `lifecycle_conformance_test.go`; its focused group measured 12.847 seconds in the producer evidence and 15.158 seconds in the independent focused barrier. Adding that group to an inherited package with only 45 to 55 seconds of high-load margin makes ordinary host variance sufficient to cross 600 seconds. The failure is therefore cumulative package duration plus insufficient margin, not a semantic assertion failure and not a hang in the named repair test.

## Ranked cumulative cost owners

1. Inherited `TestStatusReportsCompiledCurrentnessAndFailsCheck`: 184.57 seconds in the preserved per-test trace. It performs one real compiled install, then three fresh compiled-scope plans for each of 14 drift cases: JSON, fail-closed check, and human rendering.
2. Inherited repair and rollback coverage: the current focused four-test group measured 190.706 seconds. Earlier comparable attribution measured `TestInstallAndUpgradeRepairCorruptCompiledState` at 104.52 seconds, `TestInstallRepairsUntrustedCompiledStateAndPreservesTheOldInstall` at 34.69 seconds, and `TestDryRunNeverClaimsACompletedCompilerCheck` at 6.53 seconds; the accepted integrated tree also adds the real commit-failure rollback test. These are dedicated lifecycle assertions and must remain real.
3. Inherited global-status currentness matrix: reduced by previously accepted immutable-plan replay from 110.541 to about 43.0 seconds. This is the proven pattern for the proposed project-status change.
4. Inherited transitive project/global currentness: 26.42 plus 25.72 seconds.
5. Inherited GC and unusable-toolchain cases: 16.49, 13.26, and 12.72 seconds.
6. Introduced by this task: authoritative bootstrap/upgrade lifecycle group, 12.847 to 15.158 seconds. It is not the largest owner, but it consumes the remaining package margin.
7. The other 21 task delta files run in separate package test binaries. Their failed-run package totals were bounded and below their package deadlines: install 329.547 seconds, godriver 56.897 seconds, runtimestore 19.273 seconds, closure 9.496 seconds, and every other affected package below 5 seconds. They cannot consume the independent `cmd/curator` package deadline.

## Smallest semantics-preserving patch

Change only `cmd/curator/status_test.go`, inside `TestStatusReportsCompiledCurrentnessAndFailsCheck`; no product file or authoritative consumer file needs modification.

1. For every drift case, invoke the real command once as `status app --json --check`. Assert the expected non-zero check exit and decode the same JSON row. This replaces separate JSON and check plans without losing either assertion.
2. Assert every state and cause in the human form through the same `buildReport.Describe()` method used by production. Keep representative end-to-end plain-human command invocations for at least one no-cause marker drift, one cause-bearing input drift, and one cache-boundary drift.
3. For the two cache-damage cases, call the already accepted package helper `snapshotBuildCacheAfter` before mutation. Its cleanup restores every byte and permission bit. Remove only the two `reinstall` callbacks used solely to reset this reporting fixture. Do not alter the dedicated install/repair/rollback tests.
4. Keep the initial clean real install, clean JSON report, and clean check. Keep all 14 drift cases and every existing row-field, skill-demotion, outcome, detail, state, cause, and fail-closed assertion.

## Non-vacuity

The authoritative lifecycle file remains byte-identical. It still reads `manager-lifecycle.json` at runtime, rejects an empty case list, derives CLI flags from published fields, runs all three real upgrade selections, observes real Git fetch reports, checks exclusions, and requires exactly one fetch for deduplicated repositories. The proposed patch touches none of this.

The project-status matrix also stays non-vacuous: each live tamper still precedes a real combined CLI acquisition, every stable Curator code is asserted from decoded output, and a wrong state/cause/outcome or a zero check exit fails. Human formatting remains executable through the production `Describe()` method, with representative complete command-path checks. Cache reset compilation is not evidence owned by a reporting test; dedicated real repair and rollback tests remain unchanged and are included in focused gates.

## Hazards

Do not use `t.Parallel`. `capture` swaps process-global stdout/stderr, `t.Setenv` changes process-global `CURATOR_CONFIG` and toolchain selection, and the subtests mutate one shared marker/cache fixture with LIFO cleanup. Register marker restoration before `snapshotBuildCacheAfter` so the cache cleanup runs first. Keep cache cases sequential and verify the snapshot helper restores directory modes after payloads. Do not share mutable manager homes or Git working copies across top-level tests.

## Expected saving

Combining JSON and check plus replacing most human CLI repetitions removes about 25 to 28 compiled plans. At the accepted post-fingerprint estimate of about 2.15 seconds per plan this saves roughly 54 to 60 seconds. Removing two fixture-only restore installs saves at least another 17 seconds from the prior measured model. Expected total is 70 to 80 seconds, with medium-high confidence. Applied to the accepted high-load baseline plus the new lifecycle group, the projected package range is about 480 to 500 seconds, restoring roughly 100 seconds of deadline margin; comparable-load projection is about 400 to 410 seconds.

## Producer command allowlist

Run each command standalone and report its real exit code. No broader package, repository, race, Windows, cache-clearing, install, publication, stage, commit, or pin command is authorized in the producer rework.

```text
gofmt -l cmd/curator/status_test.go
git diff --check
go vet ./cmd/curator
CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 go test -count=1 ./cmd/curator -run ^TestStatusReportsCompiledCurrentnessAndFailsCheck$
CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 go test -count=1 ./cmd/curator -run ^(TestAuthoritativeBootstrapCasesAreExecutable|TestAuthoritativeUpgradeCasesAreExecutable)$
CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 go test -count=1 ./cmd/curator -run ^(TestInstallAndUpgradeRepairCorruptCompiledState|TestInstallRepairsUntrustedCompiledStateAndPreservesTheOldInstall|TestInstallAndUpgradeRestoreTheCacheWhenTheCommitFails|TestDryRunNeverClaimsACompletedCompilerCheck)$
```

The producer must attach exact focused timings and remain in development for independent full-gate verification. This diagnosis did not run a whole `cmd/curator` package, full repository, race, Windows, coverage, or product validation command and does not claim the full gate is fixed.