# TASK-260729-2kaopg — integrated timeout diagnostic

Date: 2026-07-29
Role: tester
Run: RUN-260729-2655bc
Candidate: `.temp/TASK-260729-2kaopg/worktree`

## Verdict

The integrated global-status candidate is not ready for the required review
handoff. Its first literal default-timeout `go test -count=1 ./...` gate
exited 1 because `cmd/curator` reached the standard 10-minute timeout at
602.193s. The second attempted gate was cancelled and has empty output, so it
is not evidence.

This is a test-runtime regression, not a semantic failure in the accepted
fingerprint patch or the global-status assertions. The narrow rework is to
separate global-status plan acquisition from report rendering so the CLI tests
can reuse one immutable read-only plan while still exercising the real
classifier, JSON renderer, human renderer, and fail-closed check decision.

No product or test file was edited in this tester pass, and no additional full
repository suite was run, as required by the recovery directive.

## Provenance validation

- The candidate differs from the accepted currentness tree by exactly:
  - the owned global-status surface and call-site rewires;
  - the accepted cycle-3 fingerprint patch.
- Inventory comparison found 23 added top-level tests and no missing tests:
  - 8 `TestGlobalStatus*` tests;
  - 15 fingerprint equivalence/mutation tests.
- The accepted fingerprint patch still hashes to
  `a7e0906612ce6f007bfdb3776de632dd9c7a673e9b501443be5fb3eced8f1beb`.
- The board-hosted cycle-3 reviewer verdict says `accepted` and independently
  records the exact patch, focused gates, and a 441.177s default-timeout
  `cmd/curator` pass on the accepted currentness-plus-fingerprint candidate.
- The owned global-status delta still hashes to
  `1c45ac4c1f1cce15dd871b38eaf90dec0fb214280545feec14030696d48ed65d`.
- The preserved integration evidence reconstructs both fingerprint files from
  the accepted patch byte-for-byte and reports no accepted currentness file
  reverted or absent.

## Exact gate evidence

| Evidence | Exit | Result |
| --- | ---: | --- |
| bounded `go test -count=1 -timeout 30m -json ./cmd/curator` diagnostic | 0 | package passed in 533.463s |
| focused global-status coverage run on the integrated candidate | 0 | package passed in 110.541s; owned functions remain at the recorded 80–100% coverage |
| literal `go test -count=1 ./...`, gate 1 | 1 | `cmd/curator` timed out at 602.193s; every other package passed |
| literal `go test -count=1 ./...`, gate 2 | not evidence | cancelled by the orchestrator; both output files are empty |

Gate 1's timeout stack happened to show
`TestStatusExplainsAnUnusableGoToolchain` creating its fixture, but the
extended JSON trace proves that test takes only 0.26s. It was simply the next
test when the package-wide alarm fired.

At this pass's checkpoint there were no foreign Go/test/build processes.
Available space on `/System/Volumes/Data` was 7,484,236 KiB. No additional
suite was started under that low-disk condition.

## Timing attribution

The preserved extended trace attributes 533.463s as follows:

| Top-level test | Seconds |
| --- | ---: |
| `TestStatusReportsCompiledCurrentnessAndFailsCheck` | 184.57 |
| `TestInstallAndUpgradeRepairCorruptCompiledState` | 104.52 |
| `TestGlobalStatusReportsCompiledCurrentnessAndFailsCheck` | 99.93 |
| `TestInstallRepairsUntrustedCompiledStateAndPreservesTheOldInstall` | 34.69 |
| project transitive currentness | 26.42 |
| global transitive currentness | 25.72 |
| project unusable toolchain | 13.26 |
| global unusable toolchain | 12.72 |

The global-status test file performs 26 expensive status plans over compiled
scopes:

- 20 in `TestGlobalStatusReportsCompiledCurrentnessAndFailsCheck`;
- 2 in the unusable-toolchain test;
- 4 in the transitive-command test.

Each CLI invocation necessarily re-fingerprints the whole trusted GOROOT.
The accepted cycle-3 benchmark measures the optimized fingerprint at
1.081s/op, and the prior task diagnostic measured one compiled-scope status at
about 3.1s before that 30.7% improvement, or about 2.15s after it. The two
cache-tamper subtests additionally run a real compiled reinstall solely to
restore their shared fixture.

The global tests therefore add repeated planning cost for output-mode
assertions even though the logical source, toolchain, command selection, and
planned build facts do not change between `--json`, `--check`, and human
rendering.

## Recommended rework

Keep production behavior unchanged: one invocation of `curator global status`
must still parse flags, acquire exactly one fresh read-only global plan, bracket
marker/cache classification, render the selected format, and apply `--check`.

Refactor the implementation into these internal phases:

1. parse global-status options;
2. acquire `install.Result` once with `globalStatusPlan`;
3. classify and render from that supplied result.

Then adjust only `cmd/curator/global_status_test.go`:

1. Keep representative end-to-end CLI calls for unchanged, source drift,
   cache drift, context drift, unusable toolchain, and transitive compiled
   state. Combine `--json --check` on the same invocation so the emitted JSON
   and real exit code are asserted together.
2. For the other stable-code cases, reuse the immutable planned result and
   call the same production classify/render phase after each marker, cache, or
   context mutation. Preserve every state, cause, cache outcome, detail, JSON,
   human-line, and fail-closed assertion.
3. Reuse the already installed default compiled fixture for the unusable
   toolchain subcase instead of building a second identical installation.
4. Snapshot and restore the protected cache entry byte-for-byte in test
   cleanup. Do not invoke the real reconciliation path merely to repair a
   fixture; install/repair semantics remain covered by their dedicated tests.

This reduces expensive compiled global plans from 26 to at most 6 and removes
three unnecessary compiled installations/restorations. Using the measured
post-patch plan cost, the conservative recoverable time is more than 60s:
approximately 43s from 20 avoided plans plus 17s or more from two avoided
restore installs, before counting the eliminated duplicate toolchain install.
Applied to the 533.463s green diagnostic, that projects below 475s and restores
the requested default-timeout headroom without skipping tests, changing a
timeout, weakening a stable-code assertion, or altering production semantics.

## Required replay after rework

Run, as standalone foreground processes with real exit codes:

1. focused global-status tests and coverage;
2. the accepted fingerprint mutation/equivalence and conformance/vector gates;
3. build, vet, changed-file gofmt, and diff checks;
4. after a no-foreign-Go-process and sufficient-disk barrier, two consecutive
   literal `go test -count=1 ./...` gates with the standard default timeout.

The integration checklist items for two consecutive default-timeout gates and
independent accepted review must remain unchecked until those exact events
occur.
