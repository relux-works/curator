## 2026-07-29 — curator: atomicity race timeout — fixture trim measured and rejected, production route confirmed

> **Supersedes the cycle-2 version of this entry.** The earlier entry rejected the fixture-trim
> lever on an arithmetic bound that review finding R2-1 proved invalid. The bound is retracted;
> the routing conclusion survives, now carried by measurement. Details below.

**Decision:** the `internal/install/atomicity` race-gate overrun is **not** fixable by the
remaining test-only lever. Routed to the production-side `saveJournal` fix
(TASK-260729-365r5r). The basis is a measured gate, not a model.

**What changed since the cycle-2 entry — the retraction:**
- The cycle-2 entry claimed the `references/info.md` trim removes **5** journal saves per
  installed context and is bounded by a *self-limiting* `reduction_max(C) = 5C/(60+20C)`
  (asymptote 25%, 12.5% at C=3), concluding "no value of the free parameter reaches the bar."
- **All of that is retracted.** The real figure is **8 saves per affected context**: removing
  the file removes its own staging entry *and* the now-childless `references/` directory entry
  (`captureRemovalEntries` walks directories too — `internal/transaction/journal.go:506-539`),
  so `3 saves × 2 entries + 2 saves × 1 chunk = 8`. The denominator `3N + 5F` was also wrong:
  `N = 20` was an unverified alarm-frame figure (measured `staged_targets` is 12–18 per
  transaction) and `5F` charged nothing for directory entries. Correct staging count is
  `3·entries + 2·chunks`.
- Corrected best case at the measured C=3 is **17.14%**, *above* the 14.44% required against
  the best inherited run. **The arithmetic could never have excluded the lever.** Cycle 2
  reached a correct destination by an argument that did not carry it.

**What actually settled it — measurement (TASK-260729-2afulh):**
- Measured staging inventory, four scenario phases: entries 24→20 / 34→28 / 26→22 / 25→21,
  chunks 14→12 / 19→16 / 16→14 / 16→14, staging saves 100→84 / 140→116 / 110→94 / 107→91.
  Every delta is exactly `8 × affected contexts`. Save reduction **14.55–17.14%**.
- Focused count-one race gate: **exit 0 in 493 s wall** (`ok … 492.231s`), no `DATA RACE`,
  against a strict `<=480 s` bar. **Miss of 13 s wall / 12.231 s test-reported, negative
  margin.** Non-race 286 → 273 s.
- Realized wall-clock reduction 12.12% off the best inherited run (561 s), 13.91% off the mean
  (572.67 s) — versus 14.44% / 16.18% required.

**Non-obvious result worth keeping — the calibration.** The corrected staging-save model
predicts 14.55–17.14%; the race arm delivered 13.91% against the mean. So journal saves are a
large but not exclusive share of race-instrumented runtime, and a save-count model runs ~1–3
points optimistic as a wall-clock predictor. That constant is the durable output of this whole
detour and belongs to TASK-260729-365r5r.

**Honest limit on the rejection.** The acceptance predicate is universally quantified over
repetitions, so one run at 493 s falsifies it and repetitions 2/3 were correctly not completed.
But that proves the trim *fails to demonstrate the required margin* — not that it can never
land under 480 s. The inherited arm's own spread was 32 s, larger than the 13 s miss. A bar
with negative margin is refused rather than retried. Both halves recorded so nobody re-opens
the lever on "you only ran it once", and nobody repeats cycle-2's over-claim of impossibility.

**Root cause (mechanism unchanged, magnitudes now flagged unverified):** `Engine.saveJournal`
(`internal/transaction/journal.go:71`) validates the target-namespace independence graph
**twice per journal write** (`journal.go:344` and `:354`) across **23 production call sites**
(16 `engine.go`, 7 `staging.go`), two of which sit inside the 32 KiB staging copy loop and fire
per chunk. `validateIndependentTargetNamespaces` (`namespace.go:26`) allocates `7×targets`
paths and runs a full pairwise loop (`:100-101`); on Darwin the per-path `Pathconf`/`Statfs`
helpers are **not memoised**. The quadratic shape is confirmed; the previously logged
"~2.4M–4.1M pairwise comparisons per install" inherited the unverified `N = 20` and should be
re-derived, not quoted.

**Also:** `cmd/curator` sits at 557.779 s of the same 600 s alarm (inherited, unverified) and is
next in line to fail, so the production fix pays twice.

**Two gotchas that look like violations but aren't** (still true): the cycle-1 conformance
manifest stores absolute paths while `manifest.sh` stores relative ones, so a raw diff reports
all 448 lines changed (format artifact — digests identical); and `bin/run-gates.sh` greps
positive for `-timeout` and `./...` because line 14 is a comment documenting the prohibition.

**Method lesson:** when a lever is cheap to measure, measure it. Cycle 2 dressed an unmeasured
bound as a structural proof and the reviewer was right to refuse it. The measured answer landed
in the same place — but only by 13 seconds, and the model said it should have cleared.
