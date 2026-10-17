# BUG-261016-mlpvtn \u2014 self-hosted Linux CI: logbook

2026-10-17, successor RUN-261017-a09694: recovered uncommitted workflow,
initializer and test changes from predecessor RUN-261017-609d74. Executed the
changed package tests, full project build, package vet, workflow lint,
shellcheck, gate self-tests and narrowing controls directly. Gate self-test:
267 passed, 0 failed, 2 explicitly reported host-dependent skips.

The isolation control does not reproduce the historical ABI0 linker failure.
Remote run 38991835985 has a failed Go 1.26 Linux qualification job:
TestRealTrustedGoProbe exceeded the fixed 2m fingerprint deadline (120.38s).
Other Linux jobs reported success, but the aggregate workflow was still active
and cannot be called green. A targeted rerun request was refused while active.
The old remote-gate shell log lacks a terminal exit and is not acceptance evidence.

Production inspection found no existing CI timeout override; oversized probe
and fingerprint Config values are clamped. No flag or timeout-policy change,
host mutation, cache deletion, repository-variable write or Story commit was
introduced by the successor. Host load/filesystem causation remains unknown.

Stop boundary: required green remote gate is unmet and neither pool-wide
capacity nor product deadline policy belongs to this workflow-only change.
Recommend owner-authorized pool capacity/filesystem investigation first; an
alternative is an explicit product deadline-policy/qualification decision.
Three self-hosted main successes remain rollout evidence for the orchestrator.
See BUG-261016-mlpvtn_blocked-evidence.md for AC ratio, mutant table, exact
validation exits and source/log identities. No developer review handoff issued.
