# TASK-260729-3dr6hw — review verdict cycle 5

Date: 2026-07-29
Role: reviewer
Verdict: **CHANGES REQUESTED → analysis**

No candidate file was modified and no Go command was run during this review.

## Accepted cycle-5 repair

The bounded cycle-4 evidence correction is correct:

- `603.306 / 341.415 = 1.767075…`, so the strict censored floor is correctly
  written as `> ×1.7670`, not `> ×1.77`.
- `603.701 / 441.122 = 1.368557…`, so the strict censored floor is correctly
  written as `> ×1.3685`, not `> ×1.37`.
- The atomicity row is now consistently described as a valid but severely
  censored lower bound that cannot estimate the completed factor or support a
  cross-package comparison.
- The corrected `internal/install` inventory is also right: test 73 was in
  flight, hence 72 finished, 1 in flight, and 34 never started.
- The cycle-5 SHA-256 preservation evidence keeps the previously accepted
  sections and allowlists unchanged.

The original verifier-3 facts remain correct:

- `cmd/curator` passed under race at 557.779s.
- `internal/install` timed out at 603.306s with
  `TestStrictRegistryPolicyFailsUnknown` active.
- `internal/install/atomicity` timed out at 603.701s with
  `TestFailureAtEveryTargetClassRestoresPriorStateInReverseOrder` active in
  `project-hybrid-auto` and `global-auto`.
- The verifier-3 race log contains no `DATA RACE` marker.

## Acceptance-blocking finding R5-1

The latest outcome is stale against completed evidence and, more importantly,
its required 13-file test-only patch no longer satisfies the task's runtime
acceptance claim.

The task-owned gate files for sibling producer `TASK-260729-rfrdfo` now contain
all six focused race results, each with exit 0:

| Package | repetition 1 | repetition 2 | repetition 3 | Plan's 480s bar |
| --- | ---: | ---: | ---: | --- |
| `internal/install` | 232.088s | 235.124s | 226.191s | passes by at least 244.876s |
| `internal/install/atomicity` | 591.280s | 560.828s | 564.022s | fails by 111.280s, 80.828s, and 84.022s |

This confirms Patch A and rejects the current sufficiency claim for Patch B.
The worst atomicity run leaves only 8.720s below the unchanged 600s alarm in a
focused single-package run. Section 7 deliberately requires `≤480s` because
focused runs are optimistic relative to the original concurrent
`go test -count=1 -race ./...` gate; sections 9 and 10 themselves state that
the 591.280s result should be expected to time out under that gate.

Therefore the document cannot simultaneously:

1. keep the 13-file Patch A + Patch B plan as the recommended smallest required
   patch;
2. keep the `≤480s` safety bar as necessary for the real gate; and
3. satisfy the acceptance criterion requiring a smallest test-only plan with
   expected margin below 10 minutes for both packages under race.

The additive section also remains factually stale: it says atomicity repetition
3 is in flight and all install repetitions are unstarted, although all four
have now completed successfully.

## Required bounded rework

1. Update section 10 with all six completed package lines and real exit codes.
   Record Patch A as confirmed at 226.191–235.124s and Patch B as confirmed at
   560.828–591.280s.
2. Preserve the accepted verifier inventory, censored-bound repair, hot-path
   evidence, Patch A 88/19 split, assertion map, candidate-integrity protocol,
   and non-overlapping validation commands.
3. Stop presenting the current 13-file Patch B as sufficient for the original
   race gate. Promote the `references/info.md` fixture reduction from optional
   contingency to a literal required test-only follow-up only if it can be
   supported with:
   - the exact added file/function allowlist (including
     `internal/install/atomicity/fixture_test.go` if adopted);
   - a static map showing which setup repetition is removed and why every
     assertion, fixture meaning, per-scenario isolation rule, and candidate
     integrity invariant remains preserved;
   - quantified expected savings against the measured 560.828–591.280s
     baseline and a defensible expected `≤480s` focused result.
4. If the fixture reduction cannot defensibly supply that margin, say so
   explicitly. The diagnosis must then conclude that the test-only boundary is
   exhausted and recommend the already identified production-side
   `saveJournal` namespace-validation work as the next separate story. Do not
   call a focused 8.720s margin a solution and do not change the timeout, skip
   cases, or weaken assertions.
5. Keep “Tests green” unchecked until the diagnosis's own pass condition is
   met; exit 0 under the raw timeout is not equivalent to the required margin.

This is ordinary research/decision rework, not a stop-the-line boundary.

## Evidence checked

- `.temp/TASK-260720-jrrgw9/verifier3/go-test-race-all.log`
- `.temp/TASK-260720-jrrgw9/verifier3/go-test-all.log`
- `.temp/TASK-260729-rfrdfo/gates/gate-race-atomicity-{1,2,3}.{log,exit,seconds}`
- `.temp/TASK-260729-rfrdfo/gates/gate-race-install-{1,2,3}.{log,exit,seconds}`
- Board outcome `TASK-260729-3dr6hw_install-race-timeout-diagnosis.md`
- Board outcome `TASK-260729-3dr6hw_cycle5-rework-record.md`
- Board outcome `TASK-260729-3dr6hw_cycle5-diagnosis.diff`

