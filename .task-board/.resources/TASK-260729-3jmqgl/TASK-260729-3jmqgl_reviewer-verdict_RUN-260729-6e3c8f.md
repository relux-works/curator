# TASK-260729-3jmqgl independent reviewer verdict

## Verdict

Changes requested; route to `to-dev`.

The prototype is truthful about macOS remaining unqualified, its attached source archive is byte-identical to the task worktree, and the captured machine-readable artifacts are internally consistent. The independent reviewer gate `go test -count=1 ./...` ran only in `prototypes/macos-hardened-probes` after a stable clean process barrier and exited 0 for all six packages. No probe process remained afterward.

## Acceptance evidence that passed

- All 11 capability classes appear exactly once and each report entry has a positive check plus a negative control or adversarial check.
- `evidence.json` is byte-identical to captured stdout and validates as a closed `hardened-capability-evidence-v1` record.
- The forced-unavailable sweep covers all 11 classes; every case rejects at `capability-probe` with `hardened_capability_unavailable` and exit 1.
- The report distinguishes supported, conditional/unavailable, private, and deprecated mechanisms and repeatedly disclaims production enforcement and qualification.
- The outcome documents Curator/csk reuse boundaries, exact host/tool versions, commands, and exit codes.
- `capture-evidence.sh` parses with `sh -n`; the final evidence packet contains no work directories or probe binary; no hardened-probe descendant remained.

## Blocking acceptance gap

The task explicitly requires executable evidence for process-tree/resource/disk/time limits. The `aggregate-resource-bounds` implementation at `internal/probe/classes.go:806-895` only declares and exercises an RLIMIT_NOFILE cap and a directory byte budget. It does not execute capability probes or matched controls for CPU, memory/address-space, process count, or wall-clock deadline enforcement across descendants. The only timeout at `cmd/hardened-probe/main.go:49,93-98` bounds the harness run; it is not evidence that a build domain enforces the time guarantee.

The platform inventory nevertheless groups `RLIMIT_AS`, `RLIMIT_NPROC`, and `RLIMIT_CPU` with the measured RLIMIT_NOFILE mechanism at `internal/probe/mechanisms.go:69-72`, so the report currently makes conclusions wider than its executable observations. In addition, `supervisor-side-accounting-is-unescapable` is hard-coded to `observed: escapable` / `pass: false` at `classes.go:886-893` instead of being reduced from this run's measured membership/termination result. That conflicts with the harness's stated property that a host capability change should change the observation.

A secondary fail-closed test issue exists at `internal/probe/run_test.go:31-59`: failure to build the real probe binary causes all end-to-end probe tests to skip rather than fail, allowing the test suite to pass without exercising the host probes.

## Required rework

1. Add executable, machine-reported probes and matched negative/adversarial controls for CPU, memory/address-space, process-count, and wall-clock limits over a descendant tree, retaining the existing descriptor and disk-byte probes.
2. Derive supervisor-side accounting/termination conclusions from the actual domain-membership and atomic-termination observations from the same run; do not hard-code the platform verdict.
3. Add cancellation/deadline cleanup coverage proving detached descendants cannot remain after a timed-out probe, and make an unavailable test probe binary fail the end-to-end suite rather than skip it.
4. Regenerate the task-scoped source/evidence/outcome artifacts on the macOS primary host with exact commands and exit codes, then run another independent reviewer cycle.

No code or repository documentation was modified by the reviewer.