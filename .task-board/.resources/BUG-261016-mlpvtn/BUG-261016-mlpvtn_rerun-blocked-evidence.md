# BUG-261016-mlpvtn — self-hosted Linux CI isolation: blocked rerun evidence

The required remote gate is RED. Per the revised brief, this attempt is blocked; no review handoff is admitted.

## Scope and preserved candidate

The workflow initializes job-private GOCACHE, GOMODCACHE, GOPATH and lint
state before any Go consumer in all six Linux-routable job definitions.
setup-go and golangci action caches are disabled on self-hosted Linux.
Concurrency groups use workflow/ref; only pull requests cancel prior runs.
The fork guard, hosted macOS/Windows routing, Race routing, SPEC_PIN, PNPM_PIN,
GOTOOLCHAIN=local and GOENV=off retain their pre-fix values.

Files: `.github/workflows/ci.yml`, `.github/ci/isolate-go-state.sh`,
`tools/ciworkflow/isolation_test.go`, `docs/self-hosted-runner-setup.md`,
`README.md`, and `UNRESOLVED_QUESTIONS.md`.
The resumed run preserved the implementation and tests byte-for-byte; its
only repository edit replaces the obsolete blocker in UNRESOLVED_QUESTIONS.md
with the incident resolution and outstanding rollout obligation.
The source checkout remains uncommitted at checkpoint
`612960ee90bedc7a4a113eda13aaa6cf2a2cbe9f`. The normal remote-gate entry point
created and verified signed detached snapshot
`97adb6bdbd8b6c3ead1d74cc2dbc8a768c2b9943`; it did not move the Story branch.

This is resumed producer work, not reviewer rework. Previously committed tests
and gate scripts are retained under their original names. The initializer,
Go tests, README tools entry and unresolved-questions entry support the workflow
change; there is no unrelated module delta.

## Coverage and stated bounds

**0 of 3 original AC rows fully driven; 6 of 6 Linux-routable job definitions
checked.** The revised brief explicitly permits candidate review after a green
full remote gate and assigns the three-main-run obligation to the orchestrator.

| AC | Production entry point and named tests | Evidence / bound |
| --- | --- | --- |
| 1 | ci.yml -> isolate-go-state.sh; TestLinuxCIWorkflow, TestGoStateInitialization, TestGoStateMissingRunnerContext | Pre-fix workflow control exits 1; fixed gate exits 0. Real shell initialization publishes separate paths consumed by go env, including poisoned ambient-state and missing-context cases. Historical ABI0 linker-race reproduction remains unestablished: the control demonstrates a contract violation, not that link failure. |
| 2 | ci.yml concurrency; TestLinuxCIWorkflow, TestLinuxCIWorkflowRejectsNarrowingPlants | Workflow/ref serialization and PR-only cancellation are covered. The revised brief reports that the runaway Landlock tree was cleared and runner task counts capped. This producer neither changes nor independently qualifies host-wide scheduling; a passing candidate is not arbitrary-loaded-host qualification. |
| 3 | main push workflow | 0 of 3 consecutive green self-hosted main runs established by this producer; post-integration obligation belongs to the orchestrator under the revised brief. |

Out of producer handoff scope: AC3's post-integration main-run sequence, per
“three consecutive green main runs are a post-integration obligation of the
orchestrator, not a handoff prerequisite.” AC1's historical race reproduction
and AC2's host-wide qualification remain stated bounds, not silently excluded
requirements. No other original AC row is declared out of contract.

Brief gap: no formal surface table was supplied. The missing table is reported
explicitly; no catalog-derived coverage map is claimed. Structural checking
covers the current workflow shape, not every future Actions syntax or action.
TestLinuxRoutingKeepsForksHosted executes the real routing expression subset
with Node for forks/trusted PRs/push/dispatch and all matrix operating systems;
actual GitHub scheduling is independently recorded by the remote jobs.

## Deadline finding and incident provenance

`internal/install/builddeps.go` calls ConfigFromEnvironment and godriver.Probe.
`internal/godriver/session.go` fixes ProbeTimeout at at most 15s and
FingerprintTimeout at at most 2m. ConfigFromEnvironment reads trusted toolchain
selectors only; there is no documented environment override to raise deadlines.
No deadline, product flag, timeout knob or test skip was added.

The revised brief attributes the earlier Go 1.26 fingerprint timeout to the
runaway Landlock probe tree during the incident window and reports remediation.
That is orchestrator-provided incident evidence, not a new host inspection by
this producer. The recursion fix remains a separate task.

## Narrowing mutants (reused exact evidence)


All listed red commands exited 1 with the intended named failure. No survivor
was observed among these eight executed mutants; this is not an exhaustive
mutation claim. The pre-fix control is additional to the mutants.

| Mutant | What it narrows the gate to | Named failing test | Result / survival bound |
| --- | --- | --- | --- |
| Echo initialization command in one job | Five jobs execute initialization; one only retains the command token | TestLinuxCIWorkflow | Killed |
| One setup-go cache:true with original condition retained in comment | Allow shared cache restore in one job | TestLinuxCIWorkflow | Killed |
| Cancel on pushes as well as PRs | Protect only dispatches from cancellation | TestLinuxCIWorkflow | Killed |
| Add run_id to concurrency group | Serialize only a single run, allowing same-ref overlap | TestLinuxCIWorkflow | Killed |
| Invoke isolation test only on PR | Gate remains but pushes bypass it | TestLinuxCIWorkflow | Killed |
| Checker accepts cache:true | Cache guard remains but admits one unsafe cache value | TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache | Killed |
| Initializer shares only GOMODCACHE under RUNNER_TEMP | Other state stays private; same-runner jobs share module cache | TestGoStateInitialization | Killed by executing the real shell entry point |
| Add PR bypass OR while preserving original fork-guard token | Fork PR reaches self-hosted labels in one job | TestLinuxRoutingKeepsForksHosted | Killed by executing the actual expression with Node |

The last two behavioral mutants preserve the source tokens used by structural
checks and execute the behavioral suite, rather than only the static checker.


## Validation provenance

Fresh commands in this run, each a standalone process with its real exit:

| Command | Exit | Evidence |
| --- | ---: | --- |
| `go test -count=1 -timeout=3m -v ./tools/ciworkflow/` | 0 | ciworkflow-tests-05.log; all five top-level tests and narrowing subtests green. |
| `actionlint -config-file .temp/BUG-261016-mlpvtn/actionlint.yaml .github/workflows/ci.yml` | 0 | actionlint-05.log; actionlint 1.7.12, unchanged label configuration. |
| `git diff --check` | 0 | diff-check-06.log. |

Reused from the previously attached BUG-261016-mlpvtn_blocked-evidence.md:
`go build ./...`, `go vet ./tools/ciworkflow/`, shellcheck and local gate-selftest
(all exit 0; selftest 267 passed, 2 explicitly host-dependent skips), the
pre-fix control (expected exit 1), and all eight narrowing mutants (expected
exit 1 with named failures). No missing/truncated aggregate gate log is reused
as passing evidence. Source/test/workflow hashes and every referenced log digest
were rechecked. Go 1.25.5 darwin/arm64, Node v25.6.1, actionlint 1.7.12 and the
workflow-lint configuration match the prior evidence identity. Only the
unresolved-questions prose changed. The full remote suite below is a new run,
not reused evidence from the earlier failed aggregate.

Private raw logs remain outside public board resources; public evidence omits
host identifiers, personal paths and credentials. Failed log-download attempts
are recorded separately from successful downloads and from validation exits.

## Fresh remote failure and stop boundary

Remote run [39004638406](https://github.com/relux-works/curator/actions/runs/39004638406)
ran signed snapshot `97adb6bdbd8b6c3ead1d74cc2dbc8a768c2b9943` with
LINUX_SELF_HOSTED=true already configured. This producer made no repository
variable or host changes and did not rerun the red test seeking a lucky green.

[Linux Test job 117314494625](https://github.com/relux-works/curator/actions/runs/39004638406/job/117314494625)
failed with test-gate exit 1: Go test exit 1, platform-case gate exit 0.
The job ran 2026-10-17 05:00:08–05:18:19 UTC; the failing step ran
05:00:26–05:18:14 UTC. The failure witness is
`TestAuthoritativeDryRunCasesMutateNothingPersistent/compiled-cache-miss-is-read-only`
in `internal/install/dryrun_conformance_test.go:297`. The subtest started at
05:02:47.223740471 UTC and failed at 05:04:28.708632158 UTC (Go-reported elapsed
88.89s). Its assertion reported build outcome `toolchain-unavailable` with
`go-v1 process_timeout: Go probe exceeded its deadline`, diagnostic
`process_timeout`, outside the published case's admitted outcomes.

The job log and uploaded test JSON were downloaded successfully. Artifact
[12090278181](https://github.com/relux-works/curator/actions/runs/39004638406/artifacts/12090278181)
preserves the full test evidence. Host load is UNKNOWN: no load-average,
/proc/loadavg, uptime or CPU-pressure metric was found in the downloaded job
log or the failing-test output. Neither queue/busy state nor test duration is
used as a substitute load measurement. The producer did not inspect or alter
the host. The exact runner registration name is retained in private operator
evidence; the public job link and registration ID 22 identify the execution
without copying a hostname into this public board.

Private Go initialization and Go identity checks passed in the failed job;
setup-go post-job caching was skipped. The two historical witness tests,
TestReviewProbeLegacyLaneIdentityInvalid and TestInstallGitignoreEntryPoint,
passed in this run. Their successes do not override the new failing witness.
All three Linux Go-family qualification jobs also passed, including Go 1.26.

**External blocker:** the required green self-hosted aggregate gate is unmet.
The assumption that incident remediation made the unchanged fixed-deadline
suite qualify again did not hold for this run; the cause of the new timeout
is unestablished. The revised brief explicitly directs a blocked handoff when
a Linux job fails again. No product deadline, timeout knob, fixture outcome,
platform exclusion or skip was changed to make the failure green.

Recommended next step: the pool/incident owner should correlate the failing
subtest window with host/process/capacity telemetry and the separately owned
Landlock recursion fix, establish the cause and remediate outside this
workflow-only rerun, then authorize a fresh qualification run. An alternative
is to wait for the separately owned recursion fix to land before requalification;
that delays evidence and alone does not prove the timeout cause. Hosted fallback
and raising product deadlines are excluded by the brief. Required external
input is owner diagnosis/remediation or an explicit next-scope decision,
with evidence that supports another self-hosted qualification attempt.

No reviewer Change Request is claimed while this gate is red. Candidate changes
remain uncommitted and preserved for the owning orchestrator.

## Terminal aggregate verdict

`sh scripts/remote-gate.sh` ran directly under the configured agent Git
environment and exited **1**. GitHub records the run as `completed/failure` at
the exact signed snapshot above: **19 jobs succeeded, 1 failed, 2 explicitly
skipped**. Windows Test and both hosted Race lanes succeeded. The two optional
workflow-skipped jobs provide no passing evidence for those optional paths.
This is an unexpected red qualification result, not an expected-red control.
The normal gate was not cancelled, shortened, retried after failure, or replaced
by a local mirror. The local observer returned before this lifecycle ends.

The earlier BUG-261016-mlpvtn_rerun-failure-witness.md was attached while the
aggregate was active; this terminal verdict supersedes that in-progress note.
No prior truncated aggregate log is used as success evidence.

| Self-hosted Linux job | Result | Runner registration ID | Authoritative job record |
| --- | --- | ---: | --- |
| Go 1.25.5 driver (ubuntu-latest) | success | 21 | [117314494499](https://github.com/relux-works/curator/actions/runs/39004638406/job/117314494499) |
| Gate self-test (ubuntu-latest) | success | 22 | [117314494589](https://github.com/relux-works/curator/actions/runs/39004638406/job/117314494589) |
| Test (ubuntu-latest) | failure | 22 | [117314494625](https://github.com/relux-works/curator/actions/runs/39004638406/job/117314494625) |
| Lint | success | 21 | [117314494640](https://github.com/relux-works/curator/actions/runs/39004638406/job/117314494640) |
| Interop conformance gate | success | 22 | [117314494653](https://github.com/relux-works/curator/actions/runs/39004638406/job/117314494653) |
| Naming gate | success | 22 | [117314494721](https://github.com/relux-works/curator/actions/runs/39004638406/job/117314494721) |
| Go 1.26.0 driver (ubuntu-latest) | success | 21 | [117314494748](https://github.com/relux-works/curator/actions/runs/39004638406/job/117314494748) |
| Go 1.27.1 driver (ubuntu-latest) | success | 22 | [117314494754](https://github.com/relux-works/curator/actions/runs/39004638406/job/117314494754) |

Exact runner registration names are retained in private operator evidence,
`BUG-261016-mlpvtn_runner-map-04.tsv`; this public copy uses registration IDs
and job links to avoid committing hostnames. All eight Linux jobs ran on the
self-hosted pool. The fork guard passed the fresh routing suite. The producer
observed LINUX_SELF_HOSTED=true before the gate and did not change that variable.

The worktree reports 0 commits behind its local main reference; that is not a
fresh remote-main qualification claim. All evidence is bound to the candidate
snapshot, not to current trunk. Three consecutive green self-hosted main runs
remain the orchestrator's post-integration obligation.

## Auditable identities

Candidate source identities (SHA256):

```json
{
  ".github/workflows/ci.yml": "f2d5daededfda50de425aa790338e8118334ac9002790df1b3fef45940c0378e",
  ".github/ci/isolate-go-state.sh": "1647682f0de005c2cc2def56f8ee683d2d58ce070ce20b314cf48287c6637cf0",
  "tools/ciworkflow/isolation_test.go": "89406e206e340dd7746e3d5b1f9fab2103067e713e50c5295d95188ada9b8dc6",
  "README.md": "9aa2153c622311bff029c1c556b8ca15224f321c6579ebd7cb95f5d3bc347bf9",
  "docs/self-hosted-runner-setup.md": "d68f0c3b525112022d34dfe1bbc2b6b4569268e2bbc2db4826bbaa8d97f40f28",
  "UNRESOLVED_QUESTIONS.md": "8625d747d18fdec998e7d2c4585c2654c9f471c9946c362660f8358a37c09711"
}
```

Fresh evidence digests (SHA256):

```json
{
  "ciworkflow-tests-05.log": "76f95db1df2a7ca05b358512282f59b33707a1c1d5d5bfe91e70b259fa4e6b2e",
  "actionlint-05.log": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  "diff-check-06.log": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  "remote-gate-04.log": "5558c0d0090a068688e5a93f0a6b37da1876db36b10f01bd6924646c556fad5b",
  "remote-linux-failed-04.log": "b33c77e63c44347dcb2f55187299e0e46f62939ab37ec14ba994ca32a57da397",
  "remote-linux-evidence-04/test/go-test.json": "beedc88142b38a14a7906b567e3088acafb73e78bea8b37a05520e5318e075eb",
  "remote-terminal-04.json": "801e4365404afeea8d219f852a95e08adb0bd1d6cdb9dcd5975bdddf7290c8d0",
  "remote-jobs-43.json": "1ef3fafd7f4c8fe5a3dad9bd9830627dbadfdd9feaa67a60ee8a99f032fb05d5"
}
```
