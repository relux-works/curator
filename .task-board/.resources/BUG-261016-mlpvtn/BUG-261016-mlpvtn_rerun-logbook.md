# BUG-261016-mlpvtn — self-hosted Linux CI rerun logbook

- Resumed the preserved producer candidate after the orchestrator reported the
  earlier Landlock incident cleared and runner task counts capped. No host,
  deadline, product flag or repository-variable change was made by this run.
- Preserved source/tests/workflow byte-for-byte. Updated only the obsolete
  unresolved-questions prose. The Story checkout remains uncommitted at
  checkpoint 612960ee90bedc7a4a113eda13aaa6cf2a2cbe9f.
- Fresh isolation/routing tests, actionlint and diff-check exited 0. Prior
  build/vet/shellcheck/selftest and eight narrowing mutants were reused only
  after source/tool identities and complete log digests were checked.
- Full remote gate 39004638406 ran signed snapshot
  97adb6bdbd8b6c3ead1d74cc2dbc8a768c2b9943 and exited 1:
  19 successful jobs, 1 failed, 2 explicit optional skips. No full retry or
  local mirror was used to hide the failure.
- Linux Test job 117314494625 ran 05:00:08–05:18:19 UTC on 2026-10-17.
  TestAuthoritativeDryRunCasesMutateNothingPersistent/compiled-cache-miss-is-read-only
  failed at 05:04:28 UTC with go-v1 process_timeout: Go probe exceeded its
  deadline. The reported build outcome toolchain-unavailable was outside the
  published case's admitted outcomes. Platform-case gate exit 0 did not mask
  Go test exit 1. Exact runner names remain in private operator evidence;
  the public job record and registration ID 22 bind the execution.
- Host load is unknown: the downloaded job log and failing-test output expose
  no load metric. Queue/busy state and elapsed time are not load evidence.
  All three Linux driver qualification jobs, Linux lint/selftest/interop/naming,
  macOS/Windows Test and both hosted Race lanes passed; this does not turn
  the aggregate green.
- Decision: follow the revised brief's blocked path. Required next input is
  owner diagnosis/remediation or a next-scope decision, correlating the failing
  subtest window with host/process/capacity telemetry and the separately owned
  Landlock recursion fix. Keep fixed product deadlines. A new timeout knob,
  hosted fallback or test skip is outside the brief.
- Original AC coverage remains 0 of 3 fully driven; the cache-isolation
  sub-contract covers 6 of 6 job definitions. Historical ABI0 reproduction
  and loaded-host qualification remain stated bounds. Three main successes
  belong to the orchestrator after integration, not to this handoff gate.

Evidence: BUG-261016-mlpvtn_rerun-blocked-evidence.md and
BUG-261016-mlpvtn_rerun-failure-witness.md; remote run:
https://github.com/relux-works/curator/actions/runs/39004638406
