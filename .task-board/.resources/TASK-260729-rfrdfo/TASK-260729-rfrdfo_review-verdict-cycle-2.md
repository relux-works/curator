# TASK-260729-rfrdfo — review verdict cycle 2

Date: 2026-07-29
Role: reviewer
Verdict: **CHANGES REQUESTED → analysis**

No product or test code was modified. No Go, test, race, build, vet, format,
lint, install, stage, commit, publish, or pin command was run during this
review.

## Accepted implementation and integrity evidence

The 13-file prototype patch remains accepted as a faithful, applicable
test-only prototype:

- The patch contains exactly the 13 allowlisted test files. Its added
  `t.Parallel()` calls split as 88 in `internal/install` and 7 in
  `internal/install/atomicity`; the sanctioned 19 install tests remain
  sequential.
- A current read-only manifest check matched the cycle-2 handoff evidence for
  both the 391-file jrrgw9 candidate and the 391-file prototype. A fresh
  `patch -p1 --dry-run --forward` applied all 13 entries to the exact current
  candidate with exit 0, after which the candidate manifest remained
  byte-identical.
- The current immutable conformance root matches the cycle-2 448-file digest
  evidence. The accepted-worktree deltas reproduce the recorded 23-line
  candidate baseline and corrected 34-line prototype result.
- All 13 recorded focused gate exit files contain 0. Every test command carries
  `-count=1`, none carries a timeout override or `./...`, and no gate log
  contains `DATA RACE`.
- Patch A passes the 480-second condition at 232.088s, 235.124s, and 226.191s.
  Patch B is correctly reported as exit-0 but insufficient for that condition:
  591.280s, 560.828s, and 564.022s.
- `injectClasses` remains distinct from the full `classes` success-coverage
  list. The whole-state digest, committed-class cutoff, reverse rollback,
  journal cleanup, and full 7/5 success coverage remain; retirement of the 31
  ordered cross-class defense-in-depth pairs is explicit in source and in the
  outcome.

## Acceptance-blocking finding R2-1 — cycle-2 routing arithmetic is invalid

The cycle-2 response was required to route the next lever after Patch B missed
the 480-second bar. Instead, its routing artifact and logbook reject the
`references/info.md` fixture trim using an undercounted save model:

> `reduction_max(C) = 5C / (60 + 20C)`

That model treats the trim as five journal saves per affected context. The
current jrrgw9 source proves a larger staging contribution:

1. `internal/transaction/staging.go:18-56` iterates
   `target.StagingEntries` and calls `saveJournal` three times for every entry
   (`:26`, `:33`, and `:56`).
2. A non-empty file calls `saveJournal` twice per copied 32 KiB chunk
   (`staging.go:141` and `:161`).
3. `captureRemovalEntries` walks and records directories as well as files
   (`internal/transaction/journal.go:506-533`).
4. `internal/install/atomicity/fixture_test.go:83` creates
   `references/info.md`; removing that sole child removes both the
   `references/` directory entry and the file entry.

Therefore the trim removes at least
`3 × 2 staging entries + 2 × 1 non-empty file chunk = 8` journal saves per
affected context, not 5. For the three cited project contexts the numerator is
at least 24, not 15. Even against the response's own best-case denominator,
`24 / 120 = 20%`, which exceeds its stated required reduction range of
14.4–18.8%.

This does not prove the fixture trim will meet 480 seconds. It proves the
cycle-2 claim that no assumption can meet the bar, the “self-limiting” formula,
and the conclusion that the test-only boundary is exhausted are unsupported.
The task-scoped routing artifact and logbook therefore fail their
fact-checking obligation.

## Required bounded rework

1. Preserve the accepted 13-file patch, Patch A evidence, all six focused race
   timings, manifests, assertion-retention map, and cross-class retirement
   record unchanged.
2. Retract or correct the `5C / (60 + 20C)` bound and the derived claims in
   `TASK-260729-rfrdfo_cycle2-routing-and-evidence.md` and
   `TASK-260729-rfrdfo_logbook-entry.md`.
3. Inventory or measure the actual before/after `StagingEntries`, non-empty
   file chunks, and relevant total `saveJournal` calls. Re-adjudicate whether
   the separately scoped 14th-file fixture trim can defensibly reach the
   480-second focused condition.
4. If that lever remains insufficient, support the production-side route to
   `TASK-260729-365r5r` with the corrected evidence. Do not widen this
   prototype's 13-file patch, change a timeout, skip a case, weaken an
   assertion, or run the full-repository race gate.
5. Keep “Implementation matches AC”, “Solution fits project architecture”, and
   “Tests green” unchecked until the corrected routing evidence is reviewed.

This is recoverable research and decision rework. The correct branch is
`analysis`, not `to-dev` and not `blocked`.

## Evidence inspected

- `TASK-260729-rfrdfo_results.md`
- `TASK-260729-rfrdfo_install-race-timeout.patch`
- `TASK-260729-rfrdfo_evidence-bundle.tgz`
- `TASK-260729-rfrdfo_cycle2-routing-and-evidence.md`
- `TASK-260729-rfrdfo_cycle2-evidence.tgz`
- `TASK-260729-rfrdfo_logbook-entry.md`
- `TASK-260729-rfrdfo_review-verdict-cycle-1.md`
- Current jrrgw9 candidate, prototype, accepted comparison, conformance root,
  and transaction staging/journal source
- `TASK-260729-3dr6hw_review-verdict-cycle-6.md`

