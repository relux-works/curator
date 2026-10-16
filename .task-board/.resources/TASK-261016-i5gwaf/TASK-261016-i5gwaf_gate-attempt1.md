# TASK-261016-i5gwaf — full gate attempt 1 (red)

Task: implement-rc7-muse-148-exclusion-replay. Convergence verification remains in development; no review handoff yet.

`sh scripts/remote-gate.sh` exited **1** on signed candidate snapshot `a88ce7997152b97e99389b6bbf9303e5e210e75f` (tree equal to the current uncommitted candidate, checkpoint `dccac3199aa5c543c589c129f5ef2bc437523a2a`). Hosted run: https://github.com/relux-works/curator/actions/runs/38869688515/attempts/1. Nineteen jobs passed; Test (windows-latest) failed; two optional jobs skipped.

## Logbook entry

The uploaded Windows Go JSON stream names only `internal/registry` as failing. Test `TestSnapshotFutureBoundToleratesCheckerLatency/a_slow_sibling's_latency_is_inside_the_bound` failed at registry_test.go:1071 after 8.48s: `skew 0s: a timestamp ahead of the post-fetch clock must refuse: map[]`. Its parent also fails. Platform-case gate exited 0 with 322 declared skips. The remote gate's failure-log tail omitted the actual assertion; the complete uploaded Go JSON stream was downloaded and inspected before diagnosing this.

This unchanged registry test uses 1.5s sleeps and a 3s future timestamp. Scheduling sensitivity is plausible; an external CI outage is not established and no local-mirror exception is claimed. Registry source is outside the leaf and has not been edited. Fresh M2 opt-in suite passed all 44 vendor cases, physical closure/reference regressions and offline boundary, without skips; build, vet, lint and links also passed. Interactive vendor exclusion remains FAILURE, with section 7.9 pending.

A single diagnostic rerun of only the failed Windows job was dispatched with `gh run rerun 38869688515 --job 116870600742` (dispatch exit 0, not test evidence) on the same candidate SHA. Attempt 2 remains running. The other 19 green jobs are retained by the hosting platform. No status was forged, no gate relaxed and no test result waived. The full-gate and Tests green checklist items were unchecked pending a legitimate green rerun. This initial red result remains part of the handoff evidence even if the retry passes.
