# SUPERSEDED — do not use

This document was re-derived from scratch in a later cycle of `TASK-260729-3dr6hw` **without first reading `LOGBOOK.md` or the task's prior outcomes**. It is the second time that has happened on this task (see LOGBOOK 2026-07-29 1935, `ANOMALY`).

Its root-cause analysis is correct and corroborates the earlier cycles independently. **Its projections are refuted by measurement.**

**The standing artifact is `.research/260729_install-race-timeouts.md` (cycle 7).** Read that, not this. This pointer was itself corrected in cycle 7 — see the P3 bullet below.

## What is actually true (measured, not projected)

Six focused race gates completed under `TASK-260729-rfrdfo`, each `.exit` = 0:

| Package | measured race times | verdict |
| --- | --- | --- |
| `internal/install` (Patch A) | 232.088 / 235.124 / 226.191 s | clears the 480 s bar by 245-254 s — **done** |
| `internal/install/atomicity` | 591.280 / 560.828 / 564.022 s | three greens, all above 480 s, worst margin **8.720 s** under the 600 s alarm; three-run spread 30.45 s (5.4 %) — inside its own noise |

This document's §6 projected ~400-450 s for atomicity. The real number is 560-591 s. **Ignore every projection in the superseded text.**

## Specifically wrong recommendations in the superseded text

- **P3 (drop `references/info.md`, `fixture_test.go:83`)** is **decided by measurement, not by the arithmetic quoted here.** That arithmetic — the 7.1-12.5 % ceiling — was withdrawn in cycle 7: it charged three journal saves per *target* where `internal/transaction/staging.go:16-61` charges three per *staging entry*, and it never counted the `references/` **directory** entry that disappears with its only child. `TASK-260729-2afulh` (`prototype-atomicity-fixture-trim`, `done`, review **accepted**) then measured the trim as a declared 14th file: staging-path saves `100→84`, `140→116`, `110→94`, `107→91` — a **14.55-17.14 %** reduction, *inside* the band this text called required — and one valid `-count=1 -race` gate at real exit **0**, **492.231 s**: 12.231 s over the 480 s bar, 107.769 s under the 600 s alarm. Repetition 2 produced no exit file and repetition 3 was never run. The trim is therefore **not promoted**, on wall clock rather than on work reduction. See §11 of the standing artifact. What does survive from this text: the trim leaves the target count `N` unchanged, so it cannot touch the `O((7N)²)` pairwise cost inside each surviving save.
- **P4 (partition the sweep)** is already spent. Each chain is `baseline + 1 injection + final` — the minimum the isolation invariants allow. Twelve per-class outer subtests span 196.25-247.78 s against a 78.02 s isolated chain (~3× self-contention) at ~9.6× effective concurrency on 12 performance cores. More partitioning is impossible; more parallelism oversubscribes.

## The standing conclusion

`internal/install/atomicity` does not reach the 480 s focused pass condition, so the task's test-only margin acceptance criterion is **unsatisfied** for that package. Cycle 6's "the test-only boundary is exhausted" claim was withdrawn in cycle 7 along with the invalid arithmetic supporting it; what replaces it is narrower and measured — **the best measured test-only configuration is 492.231 s on a single repetition, 2.5 % above the bar and inside the package's own 5.4 % run-to-run spread.** The package is contention-bound rather than chain-bound, which is why no *rearrangement* of the tests helps; only reducing total work does.

The higher-leverage work is production-side and **already open** as task **`TASK-260729-365r5r`** (`prototype-savejournal-namespace-validation`, parent `STORY-260720-3plyvy`, status `development`): hoist or memoise the target-namespace independence validation out of `Engine.saveJournal`. That also protects `cmd/curator`, at 557.779 s of a 600 s alarm and next in line to fail. No `-timeout` override, build tag, skip, or assertion cut.

## The one thing worth keeping

Two corrections this cycle made to its own earlier draft are factually sound and worth carrying into the production-side task:

1. Target counts are **8** for `internal/install` (race log line 82, `0x8`) and **19-20** for the atomicity scenarios (lines 225/267) — not 19-20 for both.
2. The namespace check runs **twice** per `saveJournal`: `(*Engine).validateJournal` (`journal.go:350-359`) calls `validateJournal` (which hits the check at `:344`) and then calls it again with `journalRoot` reserved.

## Process note

Before producing a "fresh" artifact on this task: read `LOGBOOK.md` and the task's own prior outcome resources first.
