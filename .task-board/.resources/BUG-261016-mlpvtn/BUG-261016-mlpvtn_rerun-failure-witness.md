# BUG-261016-mlpvtn — self-hosted Linux CI: rerun failure witness

The Linux Test job has a terminal failing result. The aggregate remote-gate
command is still active waiting for Windows Test; its exit is not yet known.
This resource makes no aggregate passing claim.

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

Job log SHA256: `b33c77e63c44347dcb2f55187299e0e46f62939ab37ec14ba994ca32a57da397`.
Merged test stream SHA256: `beedc88142b38a14a7906b567e3088acafb73e78bea8b37a05520e5318e075eb`.
