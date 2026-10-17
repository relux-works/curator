# BUG-261017-ygfdal developer evidence

Candidate: signed remote snapshot `23a4f4ee7dee14319f6215df7073dcef5b2ba900`, based on `612960ee90bedc7a4a113eda13aaa6cf2a2cbe9f`. Story HEAD is unchanged; the eight-file candidate is uncommitted in the managed worktree. The remote snapshot has a verified G signature. No existing test file was modified or removed.

## Behavior

The scriptworker package initializes its fixed hidden Landlock entry point before an importing binary can enter TestMain or main. Exact requests execute the existing probe and exit with its result; missing or extra arguments exit 1 silently. No product environment marker or user-visible command was added. Linux probeScriptLandlock additionally refuses re-probing when argv identifies a probe child. The production landlockProbeCommand constructor preserves scriptWorkerEnvironment sanitization, creates a private process group, and replaces CommandContext cancellation with SIGKILL for the group.

## Coverage

**4 of 5 AC rows driven by named behavioral tests.** AC5 includes an external full-CI obligation, declared separately below rather than counted as fully driven by a committed test. There are no out-of-contract AC rows. Brief gap: no surface table was supplied, so no catalog-row coverage map can be claimed; the AC and binary maps here are the measured scope.

| AC | Production call site | Named tests and evidence |
| --- | --- | --- |
| 1 | scriptworker.init -> RunLandlockProbe | TestLandlockProbeModeTerminatesAffectedTestBinaries: all five binaries, four argument shapes each, silent exit before the 3s bound |
| 2 | probeScriptLandlock -> isLandlockProbeChild; scriptworker.init | TestLandlockProbeChildCannotReprobe: normal caller succeeds, valid and malformed probe callers return absent; TestProductionBinaryDispatchesLandlockProbe retains the hidden usage-surface assertion |
| 3 | probeScriptLandlock -> landlockProbeCommand -> exec.Cmd.Run | TestLandlockProbeCancellationKillsProcessGroup: actual shell child and sleep grandchild; deadline and explicit cancellation; isolated Linux subreaper confirms SIGKILL, reaps the descendant, and confirms ESRCH; normal exit remains successful |
| 4 | RunLandlockProbe; probeScriptInventory -> probeScriptLandlock | Existing TestLinuxLandlockProbeFindsPresentControls and TestLinuxLandlockProbeChildAnswersAtTheBoundary pass on self-hosted Linux; owned directory exits 0, missing directory exits 1; off-Linux binary cases preserve exit 1 |
| 5 | affected binary entry; scripts/remote-gate.sh | Same named binary regression red on baseline and green on candidate. Full remote gate is external execution, not a simulated or locally mirrored status |

Every binary below imports scriptworker transitively in the go list -test dependency inventory. Its named subtests under TestLandlockProbeModeTerminatesAffectedTestBinaries exercise owned, missing, no-directory, and extra-argument requests; all 20 cases passed on self-hosted Linux.

| Test binary package | Shared dispatch and subtest prefix |
| --- | --- |
| cmd/curator | scriptworker.init; cmd/curator |
| internal/install | scriptworker.init; internal/install |
| internal/install/atomicity | scriptworker.init; internal/install/atomicity |
| internal/crossconformance | scriptworker.init; internal/crossconformance |
| internal/scriptworker | scriptworker.init; internal/scriptworker |

Stated bounds: the process-group regression calls the exact production command constructor with a controlled executable to force a deadline/cancellation; it does not force the real fast Landlock syscall sequence to take 30 seconds. The safety guarantee covers descendants remaining in the inherited process group, as required by this task; unrelated malicious process-group/session escape is outside the fixed self-probe model. Linux tests require a Landlock-capable host, matching the existing suite contract. Non-Linux cleanup is explicitly skipped because those platforms never launch this probe. Future test packages beyond the five enumerated here were not measured, although import-time dispatch also protects them. No source-text acceptance gate was introduced: source anchors only apply the overlays, and the harness requires a named behavioral failure plus its exact diagnostic.

## Narrowing mutants

TestLandlockProbeNarrowingMutants runs real Go source overlays without editing the checkout. Each retained gate is weakened, its behavioral suite runs with -count=1, and the expected-red child exits 1. Build/setup failures cannot count because both the named test and its behavioral diagnostic must occur.

| Mutant | What it narrows the gate to | Named test that fails | Result / survivor bound |
| --- | --- | --- | --- |
| dispatch-admits-extra-argument | Argument rejection only rejects too few arguments, accepting an extra argument | TestLandlockProbeModeTerminatesAffectedTestBinaries/internal/scriptworker/extra-argument | Killed, exit 1: probe exits 0 where exit 1 is required |
| reprobe-admits-malformed-child | Recursion rejection applies only to exactly three argv entries | TestLandlockProbeChildCannotReprobe | Killed, exit 1: argc=2 child reports present |
| cleanup-kills-group-only-on-deadline | Group cleanup applies to deadlines; explicit cancellation kills only the direct child | TestLandlockProbeCancellationKillsProcessGroup/cancel | Killed, exit 1: descendant survives until fixture cleanup |

**3 of 3 narrowing mutants killed; no survivors.** The hidden-mode token and the actual guards remain present in every relevant overlay. The full Linux served-suite JSON records all three exact diagnostics.

## Red-on-baseline evidence

Baseline is main/checkpoint 612960ee90bedc7a4a113eda13aaa6cf2a2cbe9f, with zero commits behind local main at inspection. Before the product edit, the newly added affected-binary regression ran against that baseline and failed as intended. It was repeated with the exact final test source, using a Go overlay that restores the original Linux product file and removes the added dispatch; the checkout was not changed. Both real regression runs exited 1. This was a bounded mode-dispatch witness on macOS, not an attempt to saturate a Linux host.

Command: go test -count=1 -timeout=2m ./internal/scriptworker -run '^TestLandlockProbeModeTerminatesAffectedTestBinaries$/internal/install/owned$' (with baseline overlay inherited by nested builds on the repeat).

Observed named failure: TestLandlockProbeModeTerminatesAffectedTestBinaries/internal/install/owned; the hidden-mode child ran TestProjectManifestABAAroundTheReadRestartsClosure and other tests instead of probing; the 3s watchdog terminated its process group. The failure diagnostic was `probe ran tests or hung: context deadline exceeded`. The sanitized probe environment also caused the incorrectly executed test to report git unavailable. This is test-execution output and a wrong entry point, not a setup failure in the regression parent.

The first draft regression invocation failed to compile due to an unused test import (exit 1); it is not counted as red regression evidence. First host lint exited 2 from a deprecated new-test API and stale cross-worktree cache; the API was corrected and independent caches were used for subsequent passing runs.

## Validation

Local commands were run directly with logs under the task scratch directory, without tee pipelines. No full local replay is claimed; the full remote gate supplies the complete cross-platform execution. Required Linux behavior and all narrowing mutants were actually executed by the self-hosted Linux Test job, not inferred from cross-compilation.

| Command | Real exit | Scope |
| --- | ---: | --- |
| go test -count=1 -timeout=3m ./internal/scriptworker -run 'Test(LandlockProbeMode\|LandlockProbeCancellation\|LinuxLandlockProbe)' | 0 | Dispatch passes; Linux-only tests skip with a stated reason |
| go test ./cmd/curator -run '^TestProductionBinaryDispatchesLandlockProbe$' -count=1 -timeout=3m | 0 | Production executable dispatch and hidden surface |
| GOOS=linux GOARCH=amd64 go test -c -o .temp/BUG-261017-ygfdal/scriptworker-linux.test ./internal/scriptworker | 0 | Final Linux test binary cross-build; not Linux execution |
| go build ./... | 0 | Repository compilation |
| go vet ./... | 0 | Repository static validation |
| make lint with a task-private lint cache | 0 | Host repository lint: 0 issues |
| GOOS=linux GOARCH=amd64 golangci-lint run ./internal/scriptworker with a task-private cache | 0 | Final Linux scope lint: 0 issues |
| git diff --check | 0 | Whitespace validation |

## Remote gate

Current run: https://github.com/relux-works/curator/actions/runs/38997554340
Snapshot: 23a4f4ee7dee14319f6215df7073dcef5b2ba900.
Self-hosted Linux Test job 117293322669 passed and published test-evidence-ubuntu-latest (artifact 12087593346). Artifact: https://github.com/relux-works/curator/actions/runs/38997554340/artifacts/12087593346. Its full go-test-served.json was downloaded and parsed: all 20 binary cases, recursion guard, cancellation regression, three narrowing mutants, existing present/absent checks, and production dispatch passed.

`sh scripts/remote-gate.sh` exited **0**. The hosting API confirms the exact snapshot head and overall success: **20 successful jobs, 2 optional skipped jobs**. No local mirror or synthesized status was used. The table records runner registration names, not machine hostnames.

| Job | Runner registration name | Result |
| --- | --- | --- |
| [Race (macos-latest)](https://github.com/relux-works/curator/actions/runs/38997554340/job/117293322455) | GitHub Actions 1000035624 | success |
| [Go 1.25.5 driver (macos-latest)](https://github.com/relux-works/curator/actions/runs/38997554340/job/117293322602) | GitHub Actions 1000035625 | success |
| [Test (macos-latest)](https://github.com/relux-works/curator/actions/runs/38997554340/job/117293322620) | GitHub Actions 1000035626 | success |
| [Interop conformance gate](https://github.com/relux-works/curator/actions/runs/38997554340/job/117293322645) | relux-rise1-curator-1 | success |
| [Gate self-test (macos-latest)](https://github.com/relux-works/curator/actions/runs/38997554340/job/117293322655) | GitHub Actions 1000035627 | success |
| [Test (ubuntu-latest)](https://github.com/relux-works/curator/actions/runs/38997554340/job/117293322669) | relux-rise1-curator-2 | success |
| [Gate self-test (ubuntu-latest)](https://github.com/relux-works/curator/actions/runs/38997554340/job/117293322670) | relux-rise1-curator-1 | success |
| [Naming gate](https://github.com/relux-works/curator/actions/runs/38997554340/job/117293322711) | relux-rise1-curator-1 | success |
| [Gate self-test (windows-latest)](https://github.com/relux-works/curator/actions/runs/38997554340/job/117293322717) | GitHub Actions 1000035628 | success |
| [Lint](https://github.com/relux-works/curator/actions/runs/38997554340/job/117293322737) | relux-rise1-curator-1 | success |
| [Go 1.26.0 driver (ubuntu-latest)](https://github.com/relux-works/curator/actions/runs/38997554340/job/117293322745) | relux-rise1-curator-2 | success |
| [Go 1.26.0 driver (macos-latest)](https://github.com/relux-works/curator/actions/runs/38997554340/job/117293322753) | GitHub Actions 1000035634 | success |
| [Test (windows-latest)](https://github.com/relux-works/curator/actions/runs/38997554340/job/117293322780) | GitHub Actions 1000035632 | success |
| [Go 1.27.1 driver (windows-latest)](https://github.com/relux-works/curator/actions/runs/38997554340/job/117293322793) | GitHub Actions 1000035629 | success |
| [Go 1.26.0 driver (windows-latest)](https://github.com/relux-works/curator/actions/runs/38997554340/job/117293322803) | GitHub Actions 1000035631 | success |
| [Go 1.25.5 driver (ubuntu-latest)](https://github.com/relux-works/curator/actions/runs/38997554340/job/117293322810) | relux-rise1-curator-1 | success |
| [Go 1.25.5 driver (windows-latest)](https://github.com/relux-works/curator/actions/runs/38997554340/job/117293322813) | GitHub Actions 1000035630 | success |
| [Race (ubuntu-latest)](https://github.com/relux-works/curator/actions/runs/38997554340/job/117293322826) | GitHub Actions 1000035633 | success |
| [Go 1.27.1 driver (macos-latest)](https://github.com/relux-works/curator/actions/runs/38997554340/job/117293322862) | GitHub Actions 1000035635 | success |
| [Go 1.27.1 driver (ubuntu-latest)](https://github.com/relux-works/curator/actions/runs/38997554340/job/117293322869) | relux-rise1-curator-2 | success |

The earlier run 38995896385 was deliberately cancelled after tightening the mutant harness to require exact diagnostics; its local gate process exited 137 and is not passing evidence. No checks were forged or replaced by a local mirror. Existing routing keeps Race (ubuntu-latest) on a hosted runner; the self-hosted Test and qualification jobs prove the Linux product behavior, without workflow or host changes. Two optional jobs (candidate suite and opt-in additional macOS lane) are skipped under the existing workflow.

## CI correction and validation history

Run 38996054029 exposed a repository-caused validation failure: the macOS Go suite exited 0 but platform-case gate exited 1 because the new off-Linux skip message was outside the existing closed vocabulary. Only the skip text in landlock_probe_other_test.go changed afterward; it now names the existing Linux-only script inventory control and explains process groups/subreaper. No workflow, classifier policy, or host changed. Its pending gate process was stopped with real exit 137 and remaining jobs cancelled, rather than presented as success.

The production classifier was run directly against the original downloaded macOS stream (exit 1, expected unrecognised skip) and a private byte-preserving fixture that replaces exactly that reason with the actual current local test output (exit 0). This is an instrument control, not a hosted check or a replayed Go suite. An initial reserialized fixture failed with exit 1 because it did not preserve the Go stream representation the checker parses; that failed attempt is not green evidence. The immutable downloaded hosted stream was never edited.

| Validation log / invocation | Real exit | Interpretation |
| --- | ---: | --- |
| regression-red-01 | 1 | Unused-import build failure, excluded from regression proof |
| regression-red-02 / regression-red-03 | 1 / 1 | Intended named baseline regression failures |
| regression-green-01 / 02 / 03 / 04 | 0 / 0 / 0 / 0 | Focused regressions at successive source revisions; 04 is current |
| production-probe-01 | 0 | Existing production dispatch/hidden surface test |
| linux-build-01 / 02 / 03 / 04 | 0 / 0 / 0 / 0 | Linux cross-compiles at successive test revisions; Linux compiled inputs are unchanged by the latest off-Linux skip text |
| build-01 + vet-01 combined shell | 0 shell | Independent build status not retained, so not used as build evidence; the following standalone build supplies it |
| build-02 | 0 | Standalone go build ./... on the unchanged production source set |
| vet-02 | 0 | Standalone go vet ./... |
| lint-01 | 2 | Deprecated new-test API plus stale cache; corrected, not passing evidence |
| lint-02 / 03 / 04 | 0 / 0 / 0 | Private-cache repository lint; 04 is current |
| lint-linux-01 / 02 | 0 / 0 | Linux scope lint at successive revisions; 02 matches the current Linux source/test set |
| skip-reason-01 | 0 | Current off-Linux test reports its classified skip; no Linux cleanup execution claimed locally |
| skip-vocab-red | 1 | Original stream rejected for the intended unknown skip reason |
| skip-vocab-green | 1 | Initial fixture representation failed; not passing evidence |
| skip-vocab-green-02 | 0 | Byte-preserving instrument fixture accepts actual current skip output |
| remote-gate-01 / 02 | 137 / 137 | Deliberately stopped; neither is full green evidence |
| remote-gate-03: sh scripts/remote-gate.sh | 0 | Exact current signed snapshot; all required remote jobs green |

The raw baseline regression excerpt below is sanitized only to remove local paths:

```
--- FAIL: TestLandlockProbeModeTerminatesAffectedTestBinaries (3.82s)
    --- FAIL: TestLandlockProbeModeTerminatesAffectedTestBinaries/internal/install (3.82s)
        --- FAIL: TestLandlockProbeModeTerminatesAffectedTestBinaries/internal/install/owned (3.00s)
            probe_reexec_test.go:47: probe ran tests or hung: context deadline exceeded; output --- FAIL: TestProjectManifestABAAroundTheReadRestartsClosure (0.00s)
                    aba_test.go:88: git [init -q]: exec: "git": executable file not found in $PATH
                        
                        gitfixture diagnostic (spawn failed, no git process ran):
                          argv0 "git" resolved: unresolved: exec: "git": executable file not found in $PATH
                          dir "<local-path>": lstat: mode=drwxr-xr-x; stat: mode=drwxr-xr-x size=64
                          cwd: <local-path>
FAIL
FAIL	github.com/relux-works/curator/internal/scriptworker	4.107s
FAIL
```

## Scope and handoff

No AC row is excluded. One related work item is outside the authorized scope: relocating the existing Linux Race job from its hosted runner to self-hosted runners. The task Scope explicitly says "no workflow or host changes". Linux Test and qualification jobs run on self-hosted runners, and the existing Linux Race runner is reported in the remote-gate table; no routing requirement is silently claimed as universal.


Production and tests remain inside internal/scriptworker. LOGBOOK.md is the only outside-module delta, required to persist the incident and reusable entry-point/process-group finding. No workflow or host configuration was changed. Existing README tool entry points remain accurate.

Evidence attached as BUG-261017-ygfdal_linux-evidence.json and this task-scoped handoff. Candidate identity was rechecked after the full gate: all 8 changed/untracked source files equal the signed remote snapshot; its tree is 103755136899324f1401da18113bd6cc5c17f3e7. Whitespace validation and the empty-index check both exited 0. The managed Story branch remains at its recorded checkpoint with the candidate uncommitted, ready for review through the native developer handoff.
