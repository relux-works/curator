# TASK-260729-3dr6hw review verdict — cycle 6

Verdict: **changes requested → `analysis`**

Review boundary: read-only. No Go, test, build, vet, format, or worktree-mutating
command was run. The candidate and prototype worktrees were not modified.

## Accepted evidence

- Verifier-3 inventory remains correct: `internal/install` timed out at
  603.306s with `TestStrictRegistryPolicyFailsUnknown` newly active, and
  `internal/install/atomicity` timed out at 603.701s with the project/global
  rollback sweeps active. No data-race marker appears in the verifier log.
- The six sibling-prototype race results were independently re-read from their
  `.exit`, `.seconds`, and package logs. `internal/install` passed at
  232.088s, 235.124s, and 226.191s. Atomicity passed at 591.280s, 560.828s,
  and 564.022s. Every real exit is 0 and no log contains a data-race marker.
- Patch A and the 88/19 concurrency split are supported by those measurements.
  The measured atomicity result does not support the required contention
  margin: all three focused runs exceed the diagnosis's 480s bar, and the
  smallest margin below the unchanged 600s alarm is only 8.720s.
- The task-scoped outcome on the board is byte-identical to
  `.research/260729_install-race-timeouts.md`, SHA-256
  `c4ba2c0ffb1c6206a1615a7fdf35242a84bd6ecaf8d325c9f4e1a38b47a09722`.

## Required finding R6-1 — section 11's save-count bound is invalid

Section 11 concludes that removing `references/info.md` can save at most
7.1–12.5%, using 15 removed saves over a claimed 120–210 total. The source does
not implement that model.

1. `internal/transaction/staging.go:18-56` iterates
   `target.StagingEntries`, not transaction targets, and calls `saveJournal`
   three times for **every staging entry** (`:26`, `:33`, `:56`).
2. A non-empty file also calls `saveJournal` twice per 32 KiB chunk
   (`staging.go:141`, `:161`).
3. `captureRemovalEntries` walks and records both directories and files
   (`internal/transaction/journal.go:506-533`).
4. `atomicity/fixture_test.go:83` creates `references/info.md`. Removing that
   only child removes two staging entries from each affected context:
   `references/` and `references/info.md`, not one file entry alone.

For each affected context the staging-path reduction is therefore at least
`3 × 2 entries + 2 × 1 file chunk = 8` journal saves. The three project
contexts cited by section 11 yield **24**, not 15. More importantly, the
section's denominator `3 × targets + 5 × files` is not the source formula:
the staging contribution is `3 × staging entries + 2 × non-empty file
chunks`, and a claim about total saves must also account for non-staging
`saveJournal` calls.

The stated 120–210 denominator and 7.1–12.5% ceiling are therefore not valid
bounds. Even substituting the corrected numerator into the document's own
best-case denominator gives 24/120 = 20%, which crosses the stated
14.4–18.8% required reduction. That does not prove the fixture trim succeeds;
it proves the current arithmetic cannot rule it out. Consequently sections
11–12 do not establish that the test-only boundary is exhausted.

## Required finding R6-2 — successor routing is misstated

The production successor already exists as **task**
`TASK-260729-365r5r`, `prototype-savejournal-namespace-validation`, under
`STORY-260720-3plyvy`; it is currently in `development`. The outcome calls it
a "new story" and does not identify it. Replace that wording with the exact
task ID and current relationship.

## Required rework

1. Preserve the accepted verifier inventory, Patch A evidence, all six
   measured focused race results, candidate-integrity protocol, assertion map,
   and the finding that the current 13-file Patch B lacks a safe `./...`
   margin.
2. Rebuild section 11 from staging-entry semantics. If claiming a percentage,
   inventory or measure the actual before/after `StagingEntries`, file chunks,
   and relevant total `saveJournal` calls. Count the disappearing
   `references/` directory as well as its file.
3. Until that evidence exists, keep the fixture trim explicitly unmeasured and
   withdraw the claims that it is bounded below the required saving and that
   the test-only boundary is exhausted.
4. If the final disposition remains production-side work, link
   `TASK-260729-365r5r` exactly and state that the original test-only margin AC
   remains unsatisfied rather than presenting the successor as a newly opened
   story.
5. Keep the "Tests green" checklist item unchecked: the focused atomicity
   package is exit-0 but misses the task's own margin, and the contended
   verifier race gate has not passed.

This is bounded evidence repair and decision work, so the correct route is
`analysis`, not `to-dev` and not `blocked`.
