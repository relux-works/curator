## 12. Cycle-7 conclusion — the test-only margin AC is unsatisfied; the decision is production-side

This is the cycle-5 verdict's required item 4, retaken on the cycle-6 verdict's corrected basis and
on evidence that did not exist when cycle 6 was written.

**Patch A is done.** `internal/install` under `-race` is **226.191–235.124s** across three runs, all
exit 0, against a 600s alarm. §4.1's 88/19 split is confirmed correct and should be kept as the
recommended patch for that package. Nothing further is needed there.

**Patch B is a real improvement that does not reach the bar.** It takes
`internal/install/atomicity` from a hard timeout (`FAIL 603.701s`) to three green runs at
**560.828–591.280s**. But every run misses §7's 480s pass condition, the worst-case margin against
the 600s alarm (**8.72s**) is inside the measurement's own 5.4 % spread, and §7's bar exists precisely
because focused single-package numbers read optimistically against
`go test -count=1 -race ./...`, where `cmd/curator` (**557.779s** under race) overlaps for nearly the
whole run. **The current 13-file Patch B has no safe `./...` margin, and on this evidence the patched
package should still be expected to time out under the real gate.**

**Patch B plus the §4.3 fixture trim is better, and still misses.** `TASK-260729-2afulh` measured it:
one valid `-count=1 -race` gate, real exit **0**, **492.231s**, no `DATA RACE`, with a measured
14.55–17.14 % reduction in staging-path journal saves (§11.2, §11.4). It misses §7's bar by
**12.231s** and clears the 600s alarm by **107.769s**. That task is `done` with an **accepted**
review verdict which reproduced the measurement, rejected the fixture-only patch as the timeout
solution, and routed onward.

**Therefore this task's own acceptance criterion — "expected margin below 10 minutes for both
packages under race", operationalised as §7's 480s focused pass condition — remains UNSATISFIED for
`internal/install/atomicity`.** It is satisfied for `internal/install`.

**Cycle 6's "the test-only boundary is exhausted" claim is withdrawn**, together with the arithmetic
that supported it (§11.1, §11.5). What replaces it is narrower and evidential: **the best measured
test-only configuration is 492.231s on a single repetition, 2.5 % above the bar and inside the
package's own 5.4 % run-to-run spread.** That is not a demonstration that no test-only path exists;
it is a demonstration that none of the identified ones has been shown to clear the bar. §11.5 names
the cheapest evidence that would settle it — repetitions 2 and 3 of the gate `TASK-260729-2afulh`
already built, on the tree it already produced.

**The higher-leverage work is already open.** The production-side fix described in §6 is **not** a
new story: it exists as task **`TASK-260729-365r5r`**,
`prototype-savejournal-namespace-validation`, parent **`STORY-260720-3plyvy`**, status
**`development`** at the time of writing. Its acceptance criteria already require that every
`saveJournal` still revalidates current filesystem namespace facts but performs at most `O(P)`
identity/resolution reads per pass instead of repeating them for `O(P²)` pairs, that fail-closed
validation is preserved before any mutation, and that focused non-race and race evidence either
demonstrates a defensible atomicity margin at or below 480s **or explicitly rejects the prototype**.
It is working in `.temp/TASK-260729-365r5r/worktree`, built from the rfrdfo prototype state so the
480s bar is quoted against the same tree that produced the 560.828–591.280s baseline. Expected effect
is on the same hot path that also protects `cmd/curator`, currently at **557.779s** of a 600s alarm
and the next package in line to fail.

**The two levers are independent and compose.** The trim reduces the *number* of `saveJournal` calls
by 14.6–17.1 %; the production fix reduces the *cost* of each surviving call from `O(P²)` filesystem
reads to `O(P)`. If the production fix lands but needs help, the trim is the obvious partner and its
evidence is already built.

**Explicitly not sanctioned, at any point on this ladder:** a `-timeout` override, a build tag that
excludes the sweep, skipping cases, weakening assertions, or reporting an 8.72s — or a single-run
492.231s — focused margin as a solved race gate.

### 12.1 Recommended routing

| Item | Where it goes |
| --- | --- |
| Patch A (`internal/install`, 88 × `t.Parallel()`, 11 files) | keep in the current test-only task; measured green with 244.9–253.8s of margin under §7's bar |
| Patch B (atomicity sweep partition, 2 files) | keep — it is strictly better than the timeout it replaces — but **do not claim it satisfies the race gate**; this task's margin AC stays unsatisfied for this package |
| §6 production `saveJournal` namespace-revalidation fix | **already open**: task **`TASK-260729-365r5r`** (`prototype-savejournal-namespace-validation`), parent `STORY-260720-3plyvy`, status `development`. Not a new story. It is the blocking work for the `./...` race gate. |
| §4.3 `references/info.md` fixture trim (14th file) | **measured and not promoted**: `TASK-260729-2afulh`, `done`, review **accepted**, one valid race gate at **492.231s** — 12.231s over §7's bar, 107.769s under the alarm. Archived, not discarded. Revisit as a *partner* to `TASK-260729-365r5r`, or settle it with repetitions 2 and 3 per §11.5, only if that task is rejected or lands short. |

### 12.2 Evidence-honesty statement for §11–§12

**No Go command, build, vet, test, or tooling invocation was executed by this task in any cycle,
including cycle 7.** Every measured number in §11.4 and §12 is read from files another task wrote.

- §11's save-count model is read from `internal/transaction/staging.go:16-61` and `:118-176`,
  `internal/transaction/journal.go:71`, `:344-357` and `:507-533`, and the `saveJournal` call-site
  inventory already verified in §2.2. Each line reference was re-read from the candidate worktree
  this cycle.
- §11.2's context-tree derivation is read from `internal/install/targets.go:39-90`,
  `internal/whitelist/whitelist.go:20-23`, `:37-99` and `:106-120`, `internal/locale/locale.go:114`
  and `:123-125`, `internal/marker/marker.go:27`, and
  `internal/install/atomicity/fixture_test.go:63-100`. Its chain and context-target counts are read
  from `internal/install/atomicity/commit_atomicity_test.go:41-93` (the two scenario constructors and
  their class lists) and `:131-215` (the sweep body).
- §11.2's **measured** rows and both `MANIFEST` lines are quoted verbatim from
  `.temp/TASK-260729-2afulh/evidence/measurement-raw.txt`. The derivation and the measurement were
  produced independently and agree on every row; the derivation was written before the measurement
  was located.
- §11.4's gate table is read from `.temp/TASK-260729-2afulh/gates/*.exit`, `*.seconds` and `*.log`,
  and from `gates/DRIVER-STOP-REASON`. **`gate-race-atomicity-2` has no `.exit` file and is reported
  as having produced no evidence; `gate-race-atomicity-3` is reported as not run.** Neither is
  presented as a result. The single valid repetition is stated as a single repetition, with the
  package's own 5.4 % spread attached.
- The per-chain aggregation in §11.2 (380 → 316 and 324 → 276) is **derived** from the measured
  per-phase rows on the assumption that the injection and final installations share the upgrade
  shape; it is labelled as derived and no conclusion rests on it alone.
- All six Patch A / Patch B gate results in §10 are read from `.temp/TASK-260729-rfrdfo/gates/*.exit`
  and `*.seconds`.
- `TASK-260729-2afulh`'s and `TASK-260729-365r5r`'s identity, parent and status were read from the
  board this cycle. `TASK-260729-365r5r`'s in-flight gate results are deliberately **not** quoted,
  because they were still being written while this section was authored.
- Checklist item "Tests green" stays **unchecked**: no Go command was run here, and §12 records that
  the downstream producers' atomicity gates, while exiting 0, do not meet this diagnosis's own pass
  condition.
