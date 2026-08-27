# TASK-260729-rfrdfo — review verdict cycle 3

Date: 2026-07-29
Role: reviewer
Verdict branch: **CHANGES REQUESTED → analysis**

No product, prototype, candidate, or accepted-comparison file was modified. No
Go, test, race, build, vet, format, lint, helper-that-invokes-Go, install,
stage, commit, publish, or pin command was run during this review.

## Accepted evidence

The implementation and integrity portions of the handoff remain accepted:

- The stored patch names exactly the 13 sanctioned test files: 11
  `internal/install` files plus
  `internal/install/atomicity/activation_test.go` and
  `internal/install/atomicity/commit_atomicity_test.go`.
- Stored manifests and the prior independent reviews establish
  `modified=13 added=0 deleted=0 unexpected=0 forbidden_touched=0`; the current
  jrrgw9 candidate and immutable conformance root remained byte-identical.
- The patch dry-run evidence applies all 13 paths cleanly to the exact current
  jrrgw9 candidate without fuzz, offset, reject, or candidate mutation.
- Patch A is accepted: the three `internal/install` race runs exited 0 at
  232.088s, 235.124s, and 226.191s, comfortably below the authoritative 480s
  focused pass condition.
- Assertion retention remains supported: 88 sanctioned install tests are
  parallelized; the 19 exclusions remain sequential; `injectClasses` is
  separate from full `scenario.classes` success coverage; the whole-state
  digest, committed-class cutoff, reverse rollback, journal cleanup, and full
  7/5 success coverage remain. Retirement of the 31 ordered cross-class
  defense-in-depth pairs is explicit.
- Cycle 3 correctly retracts the rejected `5C/(60+20C)` arithmetic. Stored
  TASK-260729-2afulh evidence measures staging saves as 100→84, 140→116,
  110→94, and 107→91, consistent with exactly eight removed saves per affected
  context.
- Partial artifacts from cancelled RUN-260729-5e854d are excluded and were not
  used.

## Acceptance-blocking finding R3-1

The authoritative diagnosis defines the focused pass condition as both real
exit 0 and printed package duration **≤480s** for each package. Patch B does
not satisfy it:

- `internal/install/atomicity` race runs exited 0 at 591.280s, 560.828s, and
  564.022s, missing the bar by 111.280s, 80.828s, and 84.022s.
- The separately scoped 14th-file fixture trim was measured by
  TASK-260729-2afulh. Its non-race run exited 0 in 273s and its only valid
  count-one race run exited 0 in 493s (`ok ... 492.231s`), still 13s wall /
  12.231s test-reported above the strict bar.
- Race repetition 2 of that trim has a zero-byte partial log and no exit file;
  repetition 3 was never started. They are correctly excluded. Because the
  criterion requires every repetition to be ≤480s, the valid 493s result
  already falsifies acceptance; stopping did not weaken the decision.

Therefore the corrected evidence proves that the fixture trim fails to
demonstrate the required margin. It does not prove that an individual future
run could never fall below 480s, and this verdict makes no such claim.

The task checklist is correspondingly honest: the focused performance item,
“Implementation matches AC”, “Solution fits project architecture”, and “Tests
green” remain unchecked. Exit 0 under the 600s alarm is not a pass of the
authoritative 480s condition.

## Required bounded route

1. Preserve the accepted 13-file patch, manifests, applicability evidence,
   Patch A result, assertion-retention map, and explicit cross-class retirement
   record unchanged.
2. Keep the 14th-file fixture-only option archived as a measured rejected
   timeout solution; do not integrate or rerun it merely to seek a favorable
   sample.
3. Continue the production-side namespace-validation optimization in
   TASK-260729-365r5r and subject that task to its own lint, focused gates, and
   reviewer cycle.
4. Do not widen this patch, change a timeout, skip a case, weaken an assertion,
   cite cancelled RUN-260729-5e854d artifacts, or run the prohibited full
   repository/full-race gate here.

This is recoverable analysis/routing work with an active successor, not a
stop-the-line boundary. The correct branch is `analysis`, not `to-dev` and not
`blocked`.

## Evidence inspected

- `TASK-260729-rfrdfo_cycle3-corrected-routing.md`
- `TASK-260729-rfrdfo_results.md`
- `TASK-260729-rfrdfo_install-race-timeout.patch`
- `TASK-260729-rfrdfo_review-verdict-cycle-2.md`
- `TASK-260729-2afulh_tester-evidence.md`
- `TASK-260729-2afulh_review-verdict.md`
- `TASK-260729-3dr6hw_install-race-timeout-diagnosis.md` revision 7 standing
  wording, including the unchanged revision-3 section 7 pass condition
- Narrow board state for TASK-260729-rfrdfo, TASK-260729-2afulh, and
  TASK-260729-365r5r
