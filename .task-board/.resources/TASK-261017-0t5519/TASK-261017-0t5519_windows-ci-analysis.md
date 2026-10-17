# Windows CI duration analysis

Task: TASK-261017-0t5519 — windows-test-duration-analysis. Read-only research, 2026-10-17 UTC. Empty repository delta is expected: no product, test, configuration or workflow changes.

The Windows Test job really is the critical path: **49m41s–57m09s, median 52m02s** across the three latest successful main CI runs returned at collection. Most time is actual test execution, especially serial CLI/profile fixture work; it is not toolchain installation or a missing outer Go cache. The last package is `cmd/curator` in all three Windows samples. `internal/envprofile` and `internal/install` overlap it and become the next bottlenecks if only CLI tests are accelerated.

Recommend first reducing repeated immutable fixture setup across these three packages and preparing balanced test-level shards. For a dedicated Windows runner, start with **8 modern dedicated vCPU, 32 GB RAM, 250 GB local NVMe/SSD**, one job at a time. Buying more cores without addressing serial work is unlikely to deliver a proportional improvement. The causal split among NTFS metadata operations, process launch, Defender scanning and CPU is **not measured** by the available artifacts; do not describe antivirus as a demonstrated cause.

## Sample, provenance and method

All three workflow conclusions and the nine selected Test job conclusions are `success`. A/B/C below are chronological, not interchangeable repetitions of one SHA. All run dates are 2026-10-17. The first two Test streams have identical test inventories; the third adds tests, including additional mutation cases. These are observational comparisons, not a controlled benchmark. Every lane uses Go 1.25.5; logs identify Windows Server 2025 / `windows-2025-vs2026`, macOS arm64, and self-hosted Linux amd64. No private runner identifiers are reproduced.

| Sample / run | Exact source SHA | Windows job | macOS job | Linux job |
| --- | ---: | ---: | ---: | ---: |
| [A: 39029649557](https://github.com/relux-works/curator/actions/runs/39029649557) | 78d7bf05f2975889e451f9bdcb4ba9b6b906f0da | [117387368850](https://github.com/relux-works/curator/actions/runs/39029649557/job/117387368850) | [117387368839](https://github.com/relux-works/curator/actions/runs/39029649557/job/117387368839) | [117387368913](https://github.com/relux-works/curator/actions/runs/39029649557/job/117387368913) |
| [B: 39029714957](https://github.com/relux-works/curator/actions/runs/39029714957) | 26d8ddbf1278ac025a3333aee1053adbdb37f1bf | [117398102557](https://github.com/relux-works/curator/actions/runs/39029714957/job/117398102557) | [117398102497](https://github.com/relux-works/curator/actions/runs/39029714957/job/117398102497) | [117398102506](https://github.com/relux-works/curator/actions/runs/39029714957/job/117398102506) |
| [C: 39042501825](https://github.com/relux-works/curator/actions/runs/39042501825) | 499b539a54efe3699581702120fdd256afe44caf | [117424639211](https://github.com/relux-works/curator/actions/runs/39042501825/job/117424639211) | [117424639027](https://github.com/relux-works/curator/actions/runs/39042501825/job/117424639027) | [117424639092](https://github.com/relux-works/curator/actions/runs/39042501825/job/117424639092) |


The companion `TASK-261017-0t5519_evidence-manifest.json` records all nine artifact IDs, original merged-stream hashes, byte counts and event counts. The workflow uploads `.temp/ci-evidence/` as `test-evidence-windows-latest`, `test-evidence-macos-latest` and `test-evidence-ubuntu-latest`; timings come from each artifact's `test/go-test.json`. Use only that merged stream; adding `go-test-served.json` again would double count. No deferred packages occur in these samples. Windows/macOS serve 85 packages; Linux serves 84 and runs the explicit godriver platform-refusal assertion separately. Each lane reports 80 passing packages and five packages without tests.

Full data are attached as `TASK-261017-0t5519_packages.csv` (85 package rows) and `TASK-261017-0t5519_tests.csv` (11,739 distinct package/test rows, including nested subtests and skipped tests). Each contains status, reported `elapsed_s`, and timestamp-derived `wall_s` for **every** run/platform column. `absent` means no event under that name in that lane, not a zero-duration pass. `skip` is not a fast successful exercise of the behavior.

`elapsed_s` is Go's terminal-event Elapsed value. `wall_s` is start/run-to-terminal timestamp difference and therefore includes a parallel test's scheduling pause. For example the compiled-status test in C reports 450.64s Elapsed but 548.892019s run-to-pass wall time. Top-level durations include their children: never add parents to subtests. Parallel tests and package durations overlap: never add them to estimate job wall time. Subtracting package start from finish does not include compilation before that package's start event, and can include TestMain work. Go documents these [event semantics](https://pkg.go.dev/cmd/test2json).

## Job and critical-path time

| Metric | A | B | C |
| --- | ---: | ---: | ---: |
| Windows total job | 57m09s | 49m41s | 52m02s |
| Windows test-gate step | 54m20s | 46m57s | 49m22s |
| macOS total job | 17m32s | 21m55s | 16m05s |
| macOS test-gate step | 16m19s | 20m02s | 14m56s |
| Linux total job | 2m20s | 2m38s | 2m06s |
| Linux test-gate step | 1m48s | 2m01s | 1m36s |


Job time uses API started/completed timestamps, excluding queue delay. The macOS Test job is 16m05s–21m55s here. The macOS Race jobs, read from job metadata only, are 43m14s / 26m30s / 35m29s (jobs 117387368763 / 117398102520 / 117424639173). Even eliminating Windows time would leave that independent workflow bottleneck. This task does not claim per-test Race analysis.

| Windows decomposition | A | B | C |
| --- | ---: | ---: | ---: |
| Windows cmd/curator elapsed | 3161.97s | 2705.87s | 2841.92s |
| CLI serial phase, start to first top-level cont | 2481.74s | 2170.91s | 2301.88s |
| CLI parallel phase, first cont to package finish | 680.29s | 534.95s | 540.04s |
| Job outside cmd/curator elapsed | 267.03s | 275.13s | 280.08s |


The serial phase is 78.5% / 80.2% / 81.0% of CLI package wall time. A/B have 291 top-level tests without `pause` and 123 with it; C has 294 and 123. This is observed event scheduling, not an assumption from source spelling. Every one of envprofile's 356 top-level tests runs without `pause`; their Elapsed sum nearly equals its full package duration.

The Windows gate spends approximately 95s / 109s / 117s before the first package-start event (planning, compilation and process startup together; no finer attribution is available). Total job time outside the CLI package is only 267s / 275s / 280s. Removing setup alone therefore cannot turn a 52-minute job into a short one. On C, envprofile starts 106s after CLI and finishes only 315s before it. With other package timings unchanged, optimizing CLI alone saves at most **5m15s** in that run (A 1m53s; B 4m37s).

## Package timings

Values are Go package Elapsed, in A / B / C order. Full timestamp wall durations and all remaining packages are in the package CSV. Rows are ranked by median Windows Elapsed.

| Package | Windows A / B / C | macOS A / B / C | Linux A / B / C |
| --- | ---: | ---: | ---: |
| cmd/curator | 3161.97s / 2705.87s / 2841.92s | 760.74s / 998.48s / 739.78s | 99.75s / 104.47s / 85.71s |
| internal/envprofile | 2940.60s / 2330.10s / 2420.52s | 257.97s / 344.03s / 266.44s | 64.34s / 67.65s / 40.78s |
| internal/install | 2677.01s / 1906.01s / 2046.80s | 743.10s / 872.10s / 657.62s | 64.50s / 66.67s / 43.25s |
| internal/crossconformance | 455.15s / 320.96s / 334.80s | 122.91s / 179.13s / 118.62s | 47.92s / 56.07s / 39.17s |
| internal/install/atomicity | 379.29s / 276.52s / 265.84s | 40.44s / 49.87s / 43.27s | 14.15s / 18.85s / 12.82s |
| internal/shell | 192.20s / 189.52s / 185.26s | 43.52s / 64.47s / 36.95s | 14.76s / 15.30s / 12.75s |
| internal/transaction | 167.92s / 128.67s / 170.08s | 14.49s / 11.09s / 6.55s | 2.86s / 3.43s / 2.21s |
| internal/pnpmsource | 144.97s / 102.62s / 107.50s | 15.01s / 16.36s / 14.34s | 25.77s / 29.58s / 22.60s |
| internal/godriver | 122.52s / 93.68s / 103.77s | 20.78s / 28.78s / 21.71s | 0.001s (refusal assertion) / 0.001s (refusal assertion) / 0.001s (refusal assertion) |
| internal/npmsource | 129.08s / 86.47s / 89.27s | 13.56s / 14.38s / 14.49s | 25.58s / 29.47s / 22.67s |
| internal/closure | 85.00s / 69.27s / 77.47s | 9.90s / 12.62s / 13.19s | 3.17s / 3.48s / 4.45s |
| internal/swiftpminterop | 84.24s / 75.59s / 77.28s | 30.23s / 31.26s / 21.18s | 29.05s / 28.87s / 21.72s |
| internal/buildrepo | 58.24s / 48.30s / 60.82s | 33.67s / 43.41s / 32.77s | 17.10s / 17.62s / 17.16s |
| internal/artifactpolicy | 45.31s / 48.85s / 48.98s | 53.59s / 74.45s / 58.59s | 56.02s / 65.00s / 45.36s |
| internal/closuregraph | 43.63s / 48.17s / 49.80s | 31.05s / 37.22s / 28.73s | 44.01s / 50.20s / 37.38s |


Median CLI is 2841.92s Windows versus 760.74s macOS and 99.75s Linux. Median envprofile is 2420.52s versus 266.44s and 64.34s. Median install is 2046.80s versus 743.10s and 64.50s. These ratios mix platform implementations, hardware and feature qualification. In particular Linux deliberately does not execute the native compiled-build lifecycle exercised by Windows/macOS. Do not promise Linux's two-minute time on Windows or use Linux as a coverage-equivalent replacement.

## Test timings

These are top-level Elapsed values, ranked by median Windows time. They locate work; they are not additive critical-path contributions. The full CSV additionally records every subtest, actual run-to-terminal wall time and platform status.

| Package / test | Windows A / B / C | macOS A / B / C | Linux A / B / C |
| --- | ---: | ---: | ---: |
| cmd/curator / `TestCompiledProjectStatusAndUntrustedRecovery` | 543.99s / 501.08s / 450.64s | 235.27s / 312.07s / 202.46s | skip / skip / skip |
| cmd/curator / `TestProfileInstallActivationMatrix` | 286.27s / 263.86s / 262.98s | 22.97s / 24.83s / 21.87s | 1.40s / 1.39s / 1.40s |
| cmd/curator / `TestGlobalStatusReportsCompiledCurrentnessAndFailsCheck` | 220.71s / 179.11s / 160.65s | 54.99s / 89.28s / 49.32s | skip / skip / skip |
| cmd/curator / `TestStatusReportsATransitivelyResolvedCompiledCommand` | 142.78s / 132.74s / 150.28s | 44.87s / 60.04s / 44.11s | skip / skip / skip |
| internal/shell / `TestShellHookRejectsMalformedRecords` | 127.33s / 133.53s / 125.25s | 29.40s / 43.06s / 26.26s | 10.65s / 10.49s / 8.91s |
| cmd/curator / `TestGlobalStatusReportsATransitivelyResolvedCompiledCommand` | 133.33s / 116.72s / 108.24s | 34.10s / 45.45s / 29.94s | skip / skip / skip |
| internal/install / `TestDraftAuditRenewalMarkersNeverRenewable` | 178.50s / 112.14s / 106.35s | 6.86s / 9.44s / 7.26s | 3.74s / 3.77s / 2.19s |
| internal/install / `TestDraftFailureAtEveryTargetClassRestoresPriorState` | 148.02s / 100.01s / 100.03s | 9.97s / 12.37s / 11.33s | 3.53s / 6.30s / 3.71s |
| cmd/curator / `TestMuseRootContextNarrowingMutants` | 118.63s / 98.32s / 95.64s | 30.56s / 31.29s / 29.38s | 11.40s / 12.81s / 10.60s |
| cmd/curator / `TestGCRetainsAndReportsReferencedCompiledState` | 101.20s / 87.75s / 97.38s | 25.07s / 37.16s / 26.74s | skip / skip / skip |
| internal/install / `TestAuthoritativeCacheRejectionsAreRebuiltNeverAdopted` | 191.09s / 96.46s / 91.53s | 9.61s / 13.73s / 10.16s | 2.62s / 1.97s / 1.21s |
| cmd/curator / `TestNativeBlackboxExternalBuildLifecycle` | 101.28s / 94.59s / 92.32s | 29.19s / 40.38s / 37.62s | skip / skip / skip |
| internal/install / `TestLegacyLaneRejectsUnboundPreviousGeneration` | 98.54s / 68.18s / 89.28s | 14.04s / 17.44s / 9.42s | 1.32s / 2.99s / 0.81s |
| cmd/curator / `TestEnvResolvePreserveLockRefusesCurrentV1AndPlainRepairMigrates` | 84.57s / 103.32s / 88.84s | 2.43s / 3.72s / 3.14s | 0.68s / 0.66s / 0.36s |
| cmd/curator / `TestCompiledInstallFollowsTheNativeControlInventoryExactly` | 98.35s / 66.09s / 80.73s | 20.22s / 32.60s / 20.94s | 1.88s / 1.83s / 1.86s |
| internal/envprofile / `TestEnvironmentsCodexSeedVectors` | 101.15s / 80.08s / 78.78s | 5.02s / 7.22s / 4.83s | 1.51s / 1.59s / 0.56s |
| cmd/curator / `TestStatusReportsAnUnusableToolchainPerCompiledCommand` | 88.29s / 69.76s / 77.82s | 20.81s / 28.24s / 21.71s | skip / skip / skip |
| internal/envprofile / `TestRC14ResolveRepairMigratesSiblingHomes` | 70.95s / 65.83s / 72.16s | 4.94s / 7.44s / 4.84s | 1.44s / 0.93s / 0.57s |
| internal/envprofile / `TestEnvironmentsReadFailureVectors` | 68.15s / 97.30s / 70.53s | 11.46s / 14.27s / 11.86s | 0.93s / 1.70s / 0.69s |
| internal/install / `TestLegacyLaneDependencyTagMovement` | 69.27s / 66.33s / 66.61s | 13.01s / 14.36s / 8.38s | 1.74s / 1.78s / 0.80s |
| cmd/curator / `TestMuseCredentialPolicyNarrowingMutants` | 66.12s / 65.58s / 131.59s | 33.78s / 40.04s / 55.97s | 15.73s / 14.09s / 16.53s |
| cmd/curator / `TestProfileGitReinstallHonoursUseAndTakeover` | 69.52s / 65.53s / 60.61s | 7.63s / 6.64s / 6.38s | 0.36s / 0.36s / 0.36s |
| cmd/curator / `TestSecurityPostureVectorsThroughCLI` | 49.24s / 64.06s / 68.31s | 5.83s / 8.32s / 5.45s | 1.06s / 1.06s / 1.05s |
| internal/envprofile / `TestStoreBoundaryResolveVectorsDriveResolve` | 97.67s / 45.22s / 61.21s | 5.91s / 7.29s / 6.35s | 1.16s / 0.39s / 0.30s |
| internal/envprofile / `TestRC14IdentityMigrationRollsBackEveryEntry` | 66.95s / 59.94s / 60.47s | 4.11s / 4.76s / 4.19s | 1.26s / 1.16s / 0.48s |


## Causes and evidential limits

1. **Repeated serial filesystem and Git fixture work is the largest measured bucket.** CLI serial work takes 36m11s–41m22s; envprofile's entirely serial process takes 38m50s–49m01s. `TestProfileInstallActivationMatrix` exercises 3 source shapes × 3 installation states × 4 flag combinations = 36 isolated scenarios; every scenario installs another profile and sets up state, and Git shapes initialize/add/commit/tag real local repositories. Its 4m23s median Windows time compares with 23s macOS and 1.4s Linux. `profileHome` uses process-global `t.Setenv`; blindly adding `t.Parallel` is invalid. Envprofile's `pinHomes` and network identities do the same. File tree protection and transaction/managed-state publication add native ownership/ACL work. Source and timings establish repeated I/O/process work; they do not measure the fraction in NTFS versus Git startup versus security scanning.

[cmd/curator/profile_install_matrix_test.go:41](https://github.com/relux-works/curator/blob/499b539a54efe3699581702120fdd256afe44caf/cmd/curator/profile_install_matrix_test.go#L41)
; 
[cmd/curator/profile_install_matrix_test.go:235](https://github.com/relux-works/curator/blob/499b539a54efe3699581702120fdd256afe44caf/cmd/curator/profile_install_matrix_test.go#L235)
; 
[cmd/curator/profile_test.go:18](https://github.com/relux-works/curator/blob/499b539a54efe3699581702120fdd256afe44caf/cmd/curator/profile_test.go#L18)
; 
[internal/envprofile/envprofile_test.go:42](https://github.com/relux-works/curator/blob/499b539a54efe3699581702120fdd256afe44caf/internal/envprofile/envprofile_test.go#L42)
; 
[internal/envprofile/network_fixture_test.go:65](https://github.com/relux-works/curator/blob/499b539a54efe3699581702120fdd256afe44caf/internal/envprofile/network_fixture_test.go#L65)


2. **Real compiled-status, repair and rollback paths repeat a deliberately expensive toolchain trust operation.** The largest single test is `TestCompiledProjectStatusAndUntrustedRecovery` (450.64s–543.99s). Its 19 direct named subtests account for 277.65s in C; approximately 173s more is outside those child intervals in the parent (fixture installation, other assertions and teardown, not a measured compiler-only figure). `newInstalledCompiledProject` builds a vendored real command through `install`; status and repair then inspect it. `godriver.Establish` runs telemetry-off, version and env subprocesses, fingerprints the whole toolchain; Probe does not memoize, and Close revalidates. The fingerprint traverses and reads files through anchored handles. Each operation also has private Go caches, so the workflow cache cannot warm these independent cold builds. The function chain is real, but no profile splits its 7–9 minutes among hashing, compilation, process start and ACL checks. An optimization must retain toolchain-mutation and refused-cache guarantees rather than memoizing trust across mutable state.

[cmd/curator/status_test.go:514](https://github.com/relux-works/curator/blob/499b539a54efe3699581702120fdd256afe44caf/cmd/curator/status_test.go#L514)
; 
[cmd/curator/status_test.go:535](https://github.com/relux-works/curator/blob/499b539a54efe3699581702120fdd256afe44caf/cmd/curator/status_test.go#L535)
; 
[internal/godriver/session.go:175](https://github.com/relux-works/curator/blob/499b539a54efe3699581702120fdd256afe44caf/internal/godriver/session.go#L175)
; 
[internal/godriver/session.go:350](https://github.com/relux-works/curator/blob/499b539a54efe3699581702120fdd256afe44caf/internal/godriver/session.go#L350)
; 
[internal/godriver/fingerprint.go:70](https://github.com/relux-works/curator/blob/499b539a54efe3699581702120fdd256afe44caf/internal/godriver/fingerprint.go#L70)
; 
[internal/godriver/session.go:422](https://github.com/relux-works/curator/blob/499b539a54efe3699581702120fdd256afe44caf/internal/godriver/session.go#L422)


3. **Nested `go test` mutation jobs are regular suite work, not an external CI phase.** On C the two Muse mutation parents consume 131.59s + 95.64s = 227.23s of CLI serial time; A/B consume 184.75s / 163.90s. Each overlay launches a new `go test -count=1` and compiles a changed package graph. C has additional mutants, so its 131.59s credential-policy parent is not a same-inventory regression against A/B. This is a concrete repeated process/compile cost, though its internal build/start split is unmeasured. Moving only these two parents off the PR path would save at most 3m47s of C's CLI work, not 40 minutes, and envprofile may cap the job saving further.

[cmd/curator/muse_credential_mutants_test.go:54](https://github.com/relux-works/curator/blob/499b539a54efe3699581702120fdd256afe44caf/cmd/curator/muse_credential_mutants_test.go#L54)
; 
[cmd/curator/muse_root_context_test.go:428](https://github.com/relux-works/curator/blob/499b539a54efe3699581702120fdd256afe44caf/cmd/curator/muse_root_context_test.go#L428)


4. **Shell startup is visible but is not the last package.** `TestShellHookRejectsMalformedRecords` runs 14 malformed-record shapes through real POSIX interpreters and PowerShell, each with both trust rollout profiles. Its Windows time is 125.25s–133.53s versus 26.26s–43.06s macOS and 8.91s–10.65s Linux. Every activation starts an interpreter; the test has no deliberate sleep. This supports optimizing repeated shell/process setup, not deleting malformed-record or PowerShell coverage. The shell package ends far before CLI, so optimizing it alone has approximately zero current critical-path gain.

[internal/shell/shell_hook_trust_test.go:821](https://github.com/relux-works/curator/blob/499b539a54efe3699581702120fdd256afe44caf/internal/shell/shell_hook_trust_test.go#L821)
; 
[internal/shell/shell_hook_trust_test.go:357](https://github.com/relux-works/curator/blob/499b539a54efe3699581702120fdd256afe44caf/internal/shell/shell_hook_trust_test.go#L357)


5. **Outer cache misses are not the explanation.** All three Windows logs show an exact-key `actions/setup-go@v5` cache hit, a successfully restored approximately 145 MB cache, Go 1.25.5, and no cache save because the primary key matched. Setup-go takes 39s / 36s / 35s, compared with gate steps of 3260s / 2817s / 2962s. macOS also hits its cache; Linux deliberately uses job-private Go state without action-cache reuse and is still much faster. Native driver builds intentionally use operation-private caches; reusing them would change an isolation/reproducibility contract. `-count=1` deliberately disables test-result reuse. Keep build/dependency caching separate from caching test verdicts.

[.github/ci/test-gate.sh:113](https://github.com/relux-works/curator/blob/499b539a54efe3699581702120fdd256afe44caf/.github/ci/test-gate.sh#L113)
; 
[.github/ci/isolate-go-state.sh:8](https://github.com/relux-works/curator/blob/499b539a54efe3699581702120fdd256afe44caf/.github/ci/isolate-go-state.sh#L8)


6. **No evidence supports real sleeps/timeouts as the main 52-minute cost.** The slow compiled-status and profile matrix bodies have no unconditional sleep. Windows' host-GOROOT lock is a no-op; it serializes relevant macOS work only. Existing cancellation/deadline tests remain genuine coverage: C's `TestMaterializeCaptureWaitReselectsAndTimeout` is 3.29s, `TestGCWaitsForTheHomeLock` 0.27s, and `TestBuildEnforcesOutputDeadlineDiskAndNonzeroExit` 1.80s. `TestLandlockProbeModeTerminatesAffectedTestBinaries` takes 20.14s, but compiles five real package test binaries and checks immediate hidden-mode exits; its 3s watchdog is a failure ceiling, not five intentional sleeps. The 120-minute Windows package timeout does not create a 120-minute wait in green runs. Indirect waits and contention within production calls cannot be excluded without profiling.

[internal/testtoolchain/lock.go:31](https://github.com/relux-works/curator/blob/499b539a54efe3699581702120fdd256afe44caf/internal/testtoolchain/lock.go#L31)
; 
[internal/scriptworker/probe_reexec_test.go:17](https://github.com/relux-works/curator/blob/499b539a54efe3699581702120fdd256afe44caf/internal/scriptworker/probe_reexec_test.go#L17)


7. **Antivirus scanning remains an unmeasured hypothesis.** Fresh test/overlay/worker executables and many file opens create a plausible scan workload. None of these nine uploaded evidence sets contains Defender performance events, scan duration, disk queue depth, child-process CPU or ETW traces; long `.exe` tests do not identify Defender as the cause. A bounded follow-up can record the slow compiled-status test and the path-only and Git-only activation cases with Defender performance recording plus CPU/file-I/O tracing, with policy unchanged. [Microsoft's analyzer reference](https://learn.microsoft.com/en-us/defender-endpoint/performance-analyzer-reference) documents `New-MpPerformanceRecording` and `Get-MpPerformanceReport`. No antivirus exclusions or protection changes are recommended from these observations alone.

## Ranked options

All gains below are planning estimates against C's 52m02s Windows job, **not measured improvements**, and are not additive. Bounds explain their basis. Any experiment must compare the same source/test/configuration and image identity, keep the case ledger, and report repeated green samples.

| Rank | Option | Expected gain / bound | Risk and retained coverage |
| --- | --- | --- | --- |
| 1 | Reduce repeated immutable fixture setup in CLI, envprofile and install: prepare an immutable source/template once, clone into isolated writable state, reuse unchanged derived plans only where their contracts permit | If serial work in all three packages drops 20–30%, CLI saves 460–691s in C; plan roughly **5–10m job gain**, subject to envprofile/install and compilation costs. The percentage is an assumption to validate | Medium. Preserve fresh mutating state, rollback/ABA tests, ACL ownership and negative controls. Sharing live mutable caches or memoizing trust would invalidate tests. Profile setup before changing it |
| 2 | Balance **top-level tests**, not just packages, over 2–4 independent Windows shards, splitting each of the three heavy packages | Timing replay below suggests **30–40m total for two shards**, **20–30m for four** (roughly 12–22m / 22–32m saved). These assume additional independent runner capacity and unchanged per-test speed, with setup/compile margin | Medium/high workflow complexity and extra runner minutes. Process isolation permits existing t.Setenv tests. Require exhaustive inventory partition, merged JSON, and a single full platform-case gate after all shards; missing/failed shards must reject. A package-only split leaves the 47m CLI package intact |
| 3 | Pilot a faster dedicated Windows runner; retain suite unchanged | If measured test work accelerates 1.5–2×, C's 47.37m CLI plus unchanged 4.67m overhead gives **28.4–36.2m job time**, about **16–24m saved**. No measured speedup exists yet | Medium cost/maintenance. Fast single-thread and low-latency local disk matter more than core count alone; preserve Windows feature qualification. Warm cache alone cannot explain this gain |
| 4 | Remove process-global test dependencies in selected helpers, then parallelize independent scenarios; tune `-p` and `-parallel` separately | Amdahl ceiling for merely eliminating the existing CLI parallel tail is **9m00s** in C; the envprofile finish reduces the current job ceiling to **5m15s**. Plan **0–5m** for flag-only tuning; gains can be negative. Converting the large serial phase is a larger refactor, covered by ranks 1–2 | Medium/high for helper refactor; low change size but contention risk for flags. Do not add t.Parallel around t.Setenv/global probes. Trial bounded combinations such as -p=2/4 and -parallel=2/4/8 with real timings, not assumed core scaling |
| 5 | Keep required affected-package Windows checks on PRs, full suite and full mutation sweep on main; optionally separate expensive mutation parents | Potentially remove most **52m PR latency** for a change proven disjoint from heavy packages; **0m full-main-job gain**. For changes selecting CLI/envprofile/install the gain can be zero. Moving just the two named mutation parents bounds C's CLI reduction at **227s**, not the whole suite | High coverage-selection risk. Follow the repository's affected-package/reverse-dependency/validation-input ownership contract and approval policy. Never substitute Linux. Omitting Windows native compiled recovery/status, DACL/reparse, `.cmd` launchers or PowerShell would defer those exact supported-feature checks until main |
| 6 | Improve ordinary dependency/build-cache reuse and startup scheduling | **0–2m plausible**, all cache hits already observed. Even erasing all non-CLI work bounds C's gain at **4m40s**; changing no test work cannot save 40m | Low for correct build cache keys; high if test verdicts or operation-private driver caches are reused. Leave -count=1 and the production isolation contract intact |
| 7 | Optimize repeated shell activation fixtures or tiny deadline tests | About **0m current job gain**: these packages finish earlier. Useful only after major bottlenecks move | Keep malformed approvals, both rollout profiles, real PowerShell and Windows spelling behavior. Shorter deadlines can increase flaky failures rather than reduce successful execution |

Go distinguishes package/build concurrency (`-p`, default GOMAXPROCS) from tests that explicitly opt into concurrency (`-parallel`); higher settings can introduce contention. [Go command documentation](https://pkg.go.dev/cmd/go#hdr-Testing_flags). The [testing contract](https://pkg.go.dev/testing#T.Setenv) forbids Setenv in parallel tests or those with parallel ancestors. Actual runner GOMAXPROCS, utilization and memory peaks were not logged; default runner specifications are not proof of the effective concurrency.

### Sharding estimate, with its assumptions exposed

For each heavy package, sort top-level tests by observed Elapsed descending and greedily assign each to the least-loaded bin. Treat each bin as serial (conservative about current in-package overlap). Use the sum of top-level Elapsed, never child durations or run-to-pass scheduling waits. This replay uses current durations; it does not simulate CPU, filesystem or compilation contention, TestMain overhead, changed caches or interaction between packages. It is a capacity-planning estimate, not evidence that a proposed sharded gate is correct.

| Largest replayed bin | A | B | C |
| --- | ---: | ---: | ---: |
| cmd/curator, 2 bins | 2086.72s | 1761.74s | 1808.24s |
| cmd/curator, 4 bins | 1043.36s | 880.87s | 904.12s |
| internal/envprofile, 2 bins | 1470.23s | 1165.02s | 1210.20s |
| internal/envprofile, 4 bins | 735.12s | 582.51s | 605.10s |
| internal/install, 2 bins | 1701.94s | 1171.62s | 1228.64s |
| internal/install, 4 bins | 850.97s | 585.81s | 614.32s |


These bins apply to each heavy package, not a claim that one global four-job matrix can run every package in 904s. One implementation could use separate heavy-package shards and a light-package lane; another could multiplex package/test selections across fewer workers and remeasure contention. Splitting only CLI exposes envprofile's 40–49 minutes. Whole-package isolation can reduce interference, but has no measured benefit here. Test-level sharding is what breaks the serial package floor.

## Windows runner sizing recommendation

| Component | Recommended starting point | Reason / limit |
| --- | --- | --- |
| CPU | **8 dedicated modern x64 vCPU**, strong single-thread performance, no burst-only allocation | Headroom for Go compilation and independent package processes. Most CLI/profile work is serial, so 16–32 cores are not the first purchase |
| RAM | **32 GB** | Headroom for parallel toolchains, OS/security services and filesystem cache. No RSS/commit peak evidence justifies 64 GB yet |
| Disk | **250 GB local NVMe/SSD**, NTFS, retain at least 50 GB free; keep job temp and Go caches on local disk | Source trees, test executables, multiple pinned toolchains, private cold-build directories and small-file metadata dominate the workload shape. Avoid network-backed scratch; measure actual latency/queue depth |
| Concurrency | **One active full Windows job per machine initially** | Extra jobs can multiply compile and filesystem contention. Increase only after measuring p95 time, peak memory and I/O |
| Tooling/state | Same Windows image family and Go 1.25.5, pinned Node/pnpm/Rust as sampled; job-private writable state | Maintains comparable feature and toolchain identity. Persist only appropriately keyed build/dependency caches, not cross-job mutable fixture state |

A lower-cost pilot with 4 modern dedicated vCPU / 16 GB RAM / 150 GB SSD can establish whether single-thread/disk improvement alone suffices. Prefer 8/32/250 for a new machine; defer larger capacity until telemetry shows CPU saturation or memory pressure. This is a workload-based recommendation, not a measured minimum. Current [GitHub public-repository runner specifications](https://docs.github.com/en/actions/reference/runners/github-hosted-runners) list Windows x64 at 4 CPU / 16 GB RAM / 14 GB SSD and macOS arm64 at 3 CPU / 7 GB RAM. Sample logs do not independently establish physical CPU model, available space or oversubscription. Compare three unchanged-suite green runs on the pilot before promising the conditional 28–36m target.

## Evidence checks, scope and handoff

| Check performed in this run | Real exit | Result / bound |
| --- | ---: | --- |
| `gh run list --branch main --workflow ci.yml --status success --limit 5 --json ...` | 0 | Latest three returned green main runs selected; no failed run was substituted |
| Nine Test job-log downloads, retry with `gh api --allow-escape-sequences .../logs` | 0 each | Initial downloads each exited 1 because gh refused terminal escape sequences; successful retry logs were used. The initial empty files were not accepted as evidence |
| Three `gh run download RUN -p 'test-evidence-*' -D ...` calls | 0 each | All nine selected uploaded test streams present |
| `python3 analyze.py` | 0 | Nine complete streams, paired lifecycle records, no fail actions, no deferred packages |
| `python3 verify.py` (revision 2) | 0 | All 11,824 package/test table rows reconciled against raw terminal events; elapsed, status and timestamp wall values checked |
| `python3 verify.py missing-terminal` (revision 2) | **1** | Expected red: `validate_lifecycle` rejects `missing_terminal` after removing C's CLI package pass. This is a failing control command, not a passing suite |
| `python3 verify.py wrong-timing` (revision 2) | **1** | Expected red: `validate_csv` rejects `elapsed_mismatch` after changing C's CLI time to 1s. This is a failing control command, not a passing suite |

The report's companion sanitized evidence log preserves source hashes, selected hosted log facts and validation outcomes without raw path-bearing test output. The analyzer/verification scripts are temporary research utilities; Python is used for JSON/CSV analysis, not as a new product test language. No new test or gate implementation is proposed by this research delta.

**Reused, not rerun locally:** all nine sampled Test stages and their platform-case gates report exit 0 in the hosted logs. Their evidence is bound to the three listed SHAs, original test/configuration inventories, Go version and recorded platform family. This does not attest current unsampled trunk, future edits, unknown runner hardware or a future optimized implementation. No product build, full-suite replay, race run, new mutation run or self-hosted Windows benchmark was executed here: the task explicitly prohibits repository/workflow/test changes and asks for existing hosted evidence. Existing reviewer tests remain untouched. Code/build/landing property gates for a repository-changing candidate are inapplicable to this empty delta; the observed hosted property gates are green on their respective exact samples.

**Brief gap:** no attack-surface table was supplied. No coverage-map rows or newly killed product mutants are claimed; the read-only AC is covered by the full timing CSVs, cause evidence, ranked estimates and sizing recommendation. Out of contract under the brief's “Read-only analysis. No product, test or workflow changes”: implementing any optimization, changing CI or security policy, adding/sharding product tests, benchmarking a newly provisioned runner, or proving the unrecorded OS-level causal split. No other supplied row was excluded. This is an initial analysis, not rework; repository diff is empty, including outside the analyzed modules.

### Existing mutation-suite evidence (performance context only)

These rows are **reused C hosted observations**, not locally executed mutations. Named failing child tests and child exit 1 come from the original go-test JSON logs; the containing parents pass. The two skipped Windows rows are survivors for a Windows-only coverage claim: neither has a named Windows failure. They are not counted among the 19 observed kills. This table narrows the meaning of the mutation timing bucket; it does not introduce a mutation gate for this research task.

| Mutant | Gate narrowed to / change exercised | Named failing test in C | Observation and survival bound |
| --- | --- | --- | --- |
| fragment-reobserves-muse-release | Reobserve vendor release during materialization | None on Windows | **Survivor (skipped)**; Unix shell sentinel not exercised on Windows |
| muse-keychain-admitted | Admit keychain only for Muse | `TestMuseCredentialPolicyMatrix/keychain/shared` | Killed; child exit 1 |
| missing-shared-source-admitted | Admit absent native shared credential source | `TestMuseMissingNativeFilePolicy/shared` | Killed; child exit 1 |
| isolated-muse-native-link | Link native credentials only for isolated Muse | `TestMuseCredentialPolicyMatrix/file/isolated` | Killed; child exit 1 |
| renderer-drops-only-backend | Drop only mandatory backend environment variable | `TestMuseCLIStorePolicy` | Killed; child exit 1 |
| header-only-root-fact-dropped | Drop fact only for header-only root context | `TestMuseRootContextVectors/zero-applicable` | Killed; child exit 1 |
| boundary-admits-HOME | Admit HOME only for Muse fragment | `TestMuseFragmentBoundary/HOME` | Killed; child exit 1 |
| boundary-admits-missing-literal | Admit missing backend literal while preserving other validation | `TestMuseFragmentBoundary/missing-literal` | Killed; child exit 1 |
| boundary-admits-false-root | Admit false exclusive root fact only for Muse | `TestMuseFragmentBoundary/false-root` | Killed; child exit 1 |
| unrelated-root-mints-fact | Mint root fact from unrelated home-level file | `TestMuseRootContextVectors/no-context` | Killed; child exit 1 |
| linked-root-fact-dropped | Drop root fact only for symlink form | `TestMuseRootContextVectors/monolithic` | Killed; child exit 1 |
| denied-root-admitted | Ignore permission denial only for Muse root | None on Windows | **Survivor (skipped)**; POSIX chmod unreadability cannot establish Windows ACL denial |
| shared-status-misreports-isolation | Report shared isolation as isolated | `TestMuseCredentialPolicyStatus/shared` | Killed; child exit 1 |
| readlink-denial-admitted | Admit unreadable auth symlink | `TestMuseAuthRepairAdditionalStates/readlink-error` | Killed; child exit 1 |
| configured-form-ignored | Ignore only Muse referenced form configuration | `TestMuseRootContextVectors/referenced-refused` | Killed; child exit 1 |
| referenced-fallback | Map only Muse referenced form to monolithic | `TestMuseRootContextVectors/referenced-refused` | Killed; child exit 1 |
| no-context-admitted | Admit no-context package only for Muse | `TestMuseRootContextVectors/no-context` | Killed; child exit 1 |
| wrong-target | Move only Muse root target | `TestMuseRootContextVectors/monolithic` | Killed; child exit 1 |
| wrong-bytes | Append unexpected bytes only to Muse header | `TestMuseRootContextVectors/zero-applicable` | Killed; child exit 1 |
| changed-root-admitted | Ignore only changed Muse root bytes | `TestMuseRootContextDrift/changed` | Killed; child exit 1 |
| removed-root-admitted | Ignore only absent Muse root | `TestMuseRootContextDrift/removed` | Killed; child exit 1 |


Ready for review with analysis as the outcome and an empty repository delta. No implementation or CI optimization has been applied.
