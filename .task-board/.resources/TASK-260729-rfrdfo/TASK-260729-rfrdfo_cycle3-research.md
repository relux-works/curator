# TASK-260729-rfrdfo — cycle 3: corrected arithmetic and measured routing

Date: 2026-07-29
Role: researcher
Cycle: 3 (response to `TASK-260729-rfrdfo_review-verdict-cycle-2.md`, finding **R2-1**)
Constraint in force: `TASK-260729-rfrdfo_evidence-only-constraint.md` — evidence-only, fail-closed no-Go scope.

---

## 0. Headline

The cycle-2 verdict was right and cycle-2 was wrong. Three things follow.

1. **R2-1 is upheld and then some.** The per-affected-context save removal is **exactly 8**,
   not 5, and not merely "at least 8". The `5C / (60 + 20C)` bound, its 25% asymptote, its
   "self-limiting" framing and the conclusion that *no assumption inside the model reaches the
   bar* are **retracted in full**. Corrected best case at the measured C=3 is **17.14%**, which
   sits **above** the 14.44% required against the best inherited run. The arithmetic could
   never have excluded the lever.
2. **The lever was then actually measured, and it still misses.** TASK-260729-2afulh built the
   14-file trim and ran it: focused count-one race **exit 0 in 493 s wall** (`ok … 492.231s`),
   against a strict `<=480 s` bar. Miss of **13 s wall / 12.231 s test-reported**, with
   negative margin.
3. **Routing is unchanged in destination but replaced in basis.** Fixture-only trim →
   **rejected**, on measurement rather than on a broken model. Production route stays
   **TASK-260729-365r5r**. Patch A stays accepted; Patch B stays as recorded; this task's
   13-file patch is not widened.

**No Go command of any kind was run this cycle** — no test, probe, benchmark, build, vet, lint,
Go-invoking helper script, or detached process. No product or prototype file was edited. See §7.

---

## 1. Retractions

### 1.1 Retracted: "5 saves per affected context"

Cycle-2 §2.2 modelled the `references/info.md` trim as removing 5 journal saves per installed
context. Source and measurement both say **8**.

Source, read this cycle in the jrrgw9 candidate (`.temp/TASK-260720-jrrgw9/worktree`):

| Fact | Location |
| --- | --- |
| `stageTarget` calls `saveJournal` **3×** per `StagingEntry` | `internal/transaction/staging.go:26`, `:33`, `:56` |
| `copyStagingFile` calls `saveJournal` **2×** per non-empty 32 KiB read chunk | `internal/transaction/staging.go:141`, `:161`; chunk size `:12` |
| `captureRemovalEntries` walks and records **directories as well as files** (`entry.Kind = "directory"`) | `internal/transaction/journal.go:506-539` |
| `references/info.md` is written once, and only once, in the package | `internal/install/atomicity/fixture_test.go:83` |

`"ref"` is 3 bytes → exactly one non-empty chunk. Removing the file removes its own staging
entry **and** the now-childless `references/` directory entry:

```
3 saves × 2 staging entries  +  2 saves × 1 chunk  =  8 saves per affected context
```

Cycle-2 counted the file and forgot the directory. That is the whole error, and it is exactly
what R2-1 diagnosed.

**Measurement confirms 8 as an equality, not a floor.** TASK-260729-2afulh's
`evidence/measurement-raw.txt` emits the staging manifest directly:

```
before: entries=directory:, directory:references, file:.csk-install.json, file:SKILL.md, file:references/info.md
after:  entries=directory:,                        file:.csk-install.json, file:SKILL.md
```

Two entries removed per context, one chunk removed per context. Its `ref_entries` counter reads
4 / 6 / 4 / 4 before and **0** after — i.e. 2 per affected context in every scenario phase.

### 1.1.1 Three independent derivations converge on 8 — and on `21 → 13`

This cycle's source reading, review finding **R2-1**, and `TASK-260729-3dr6hw` cycle 7 (finding
**R6-1**) all reached 8 independently. The measured staging manifest closes it as an equality
at the per-context-target level:

| | staged entries | staging saves |
| --- | --- | ---: |
| before | 2 directory (`.`, `references`) + 3 file (`SKILL.md`, `.csk-install.json`, `references/info.md`) | `3×5 + 2×3` = **21** |
| after | 1 directory (`.`) + 2 file (`SKILL.md`, `.csk-install.json`) | `3×3 + 2×2` = **13** |

`21 − 13 = 8`, matching 3dr6hw R6-1's corrected model exactly.

This also reconciles a premise that reads like a contradiction. Cycle-2 §2.1's "each fixture
skill contributes 4 files" is **correct about what the fixture writes** (`SKILL.md`,
`references/info.md`, `scripts/<name>-tool`, `csk-skill.json`,
`internal/install/atomicity/fixture_test.go:75-100`). Only **3** of them are ever *staged*:
`whitelist.IncludeRoots` (`internal/whitelist/whitelist.go:20-23`) is
`SKILL.md, agents, references, .skill_triggers, assets, templates, examples, data` — it carries
`references/` (which is why `references/info.md` is copied at all) but **not** `csk-skill.json`;
and `scripts/` is only appended when `includeScripts` is set (`whitelist.go:48-51`), which
`runtime_roots: ["scripts"]` suppresses. `.csk-install.json` is not a fixture file at all — it
is the install marker written by the installer (`internal/marker/marker.go:27`,
`internal/hashing/hashing.go:22`). The staged set is therefore `SKILL.md`, `.csk-install.json`,
`references/info.md`. Both statements are true about different sets; the error was using the
*written* count as a *staging* floor.

### 1.2 Retracted: the denominator `saves = 3N + 5F`, N = 20, F ∈ [12, 30]

Wrong in both terms.

- **`3N` with N = 20.** The 20 is the alarm-frame target count, inherited from the diagnosis
  and — as cycle-2 §7 itself admitted — never independently re-derived. Measured
  `staged_targets` per transaction is **12 / 18 / 14 / 15**, not 20.
- **`5F` per staged file.** 5 is the cost of a *non-empty file* entry (3 + 2). It silently
  charges nothing for directory entries, which cost 3 each and which
  `captureRemovalEntries` demonstrably creates.

The correct staging-path count is `3·entries + 2·chunk_syncs`. It reproduces **all eight**
measured rows exactly, with no residual:

| scenario | phase | entries | chunks | staging saves | affected ctx | Δ saves | Δ% |
| --- | --- | --- | --- | --- | ---: | ---: | ---: |
| project | baseline | 24 → 20 | 14 → 12 | 100 → 84 | 2 | −16 | **16.00%** |
| project | upgrade | 34 → 28 | 19 → 16 | 140 → 116 | 3 | −24 | **17.14%** |
| global | baseline | 26 → 22 | 16 → 14 | 110 → 94 | 2 | −16 | **14.55%** |
| global | upgrade | 25 → 21 | 16 → 14 | 107 → 91 | 2 | −16 | **14.95%** |

Every Δ is `8 × (affected contexts)`. Corrected staging-save reduction range: **14.55–17.14%**.

### 1.3 Retracted: `reduction_max(C) = 5C / (60 + 20C)` and everything derived from it

Retracted: the formula, the 25% asymptote, the "self-limiting" characterisation, the six-cell
miss table, the claim that the best reachable cell is 490.72 s, and the conclusion that *"at
the actual C=3, no assumption inside the model reaches the bar."*

At the measured C=3 the corrected best case is **24 / 140 = 17.14%**, versus a required
**14.44%** against the best inherited run (561 s). The lever sat *inside* the required band.
R2-1's independent estimate of `24 / 120 = 20%` used cycle-2's own optimistic denominator; the
measured denominator is 140, so the true figure is 17.14% — lower than R2-1's, still above the
requirement, and the finding stands unchanged.

### 1.4 Retracted as unsupported: "the test-only boundary is exhausted"

Cycle-2 §3 reached the right routing destination by an argument that did not carry it. The
destination is re-established below **on measurement**. The claim of structural impossibility is
withdrawn and is not re-asserted anywhere in this document.

### 1.5 Downgraded to unverified: the N = 20 syscall magnitudes

Cycle-2 §2.4's "~140 paths, ~9,800 pairwise comparisons per pass, ~2.4M–4.1M per install"
inherits the same unverified N = 20. Measured `staged_targets` is 12–18 per transaction.

The **kind** of the result is unaffected and re-confirmed against source this cycle:
`validateIndependentTargetNamespaces` (`internal/transaction/namespace.go:26`) allocates
`len(targets)*7 + len(reserved)` paths and runs a full pairwise loop at
`namespace.go:100-101`; `saveJournal` (`journal.go:71`) reaches it **twice per write**
(`journal.go:344` inside `validateJournal`, and `journal.go:354` inside
`engine.validateJournal` with the journal root as reserved path); the Darwin per-path helpers
are not memoised. At N = 18 that is 126 paths and ~7,900 unordered pairs per pass — same order,
different constant. **The specific counts should be re-derived by TASK-260729-365r5r, not
inherited from here.**

### 1.6 What is *not* retracted

Everything in cycle-2 §1 (preservation manifests, patch dry-run, conformance-root immutability,
the path-format artifact), §1.1 (semantic parity, 88/19 split, assertion counts), §1.2 (13/13
gate exits, comment-line grep hits), §2.1 (the §4.3 premises table), and §2.4's mechanism
*shape* stands. Independently re-confirmed this cycle by source reading:
**23 production `saveJournal` call sites** — 7 in `staging.go` (`:26 :33 :56 :141 :161 :218
:244`) and 16 in `engine.go` (`:64 :105 :259 :325 :332 :339 :359 :401 :413 :443 :538 :547 :588
:658 :772 :823`), plus the declaration at `journal.go:71`. Exactly as recorded.

---

## 2. The measurement that actually decides it

All figures below are read from stored gate artifacts. Nothing was re-run.

### 2.1 Inherited Patch B baseline (rfrdfo cycle-1, preserved)

`TASK-260729-rfrdfo_evidence-bundle.tgz`, `gates/`:

| gate | exit | wall | test-reported |
| --- | ---: | ---: | ---: |
| `atomicity` non-race | 0 | 286 s | 285.434 s |
| `atomicity` race rep 1 | 0 | 593 s | 591.280 s |
| `atomicity` race rep 2 | 0 | 561 s | 560.828 s |
| `atomicity` race rep 3 | 0 | 564 s | 564.022 s |

No `DATA RACE` marker. All three race repetitions clear exit 0 and all three miss `<=480 s`.

### 2.2 14-file trim arm (TASK-260729-2afulh, preserved)

`TASK-260729-2afulh_tester-evidence.tgz`, `gates/`:

| command | exit | wall | evidence |
| --- | ---: | ---: | --- |
| `gofmt -l internal/install internal/install/atomicity` | 0 | 0 s | empty |
| `go build ./...` | 0 | 1 s | empty |
| `go vet ./internal/install/...` | 0 | 0 s | empty |
| `go test -count=1 -v ./internal/install/atomicity` | 0 | **273 s** | `ok … 272.580s` |
| `go test -count=1 -race ./internal/install/atomicity` | 0 | **493 s** | `ok … 492.231s`, no `DATA RACE` |

Race repetition 2 took SIGTERM in flight: **0-byte log, no `.exit` file** — correctly excluded
from evidence. Repetition 3 was never started. `gates/DRIVER-STOP-REASON` records the
cooperative stop under orchestrator directive `RUN-260729-b2a441:nudge:3afa0e`, and the
`DRIVER-DONE` marker is annotated as manual (`2026-07-29T19:47:35+0400
cooperative-stop-after-race1`) — it does not claim natural driver completion.

### 2.3 Realized versus required

| baseline | realized reduction to 493 s | required to reach 480 s | verdict |
| --- | ---: | ---: | --- |
| best inherited run, 561 s | **12.12%** | 14.44% | short |
| inherited mean, 572.67 s | **13.91%** | 16.18% | short |
| worst inherited run, 593 s | **16.86%** | 19.06% | short |

Non-race arm: 286 → 273 s = **4.55%** wall (285.434 → 272.580 = 4.50%).

**Miss: 13 s wall, 12.231 s by test-reported duration.**

### 2.4 Model calibration — the first measured anchor this line of reasoning has had

The corrected staging-save model predicts **14.55–17.14%**. The race arm realized **13.91%**
against the inherited mean. So the corrected model is directionally right and **mildly
optimistic**: journal saves are a large but not exclusive share of the race-instrumented
runtime, and the shortfall between predicted save reduction and realized wall-clock reduction
is roughly 1–3 points.

That is a genuinely useful number for TASK-260729-365r5r, and it did not exist before this
measurement. Cycle-2 asserted a bound with `F` *statically bounded, not measured*; that is
precisely the gap the reviewer refused to accept, and it is now closed with real inventory.

---

## 3. Why one repetition settles it — stated with its limit

The acceptance predicate is **universally quantified** over the count-one race repetitions:
*every* focused repetition must land `<=480 s`. A single repetition at 493 s falsifies the
conjunction regardless of what repetitions 2 and 3 would have produced. Stopping was sound
evidence discipline, not truncated evidence, and it released the shared Go slot for the
successor task.

**The limit, stated plainly so nobody has to re-derive it:** this proves the trim **fails to
demonstrate the required margin**. It does *not* prove the trim can never land under 480 s in
an individual run. The inherited arm's own spread was 593 − 561 = **32 s**, larger than the
13 s miss. A second repetition could plausibly have come in under the bar.

That is exactly why a bar with negative margin is refused rather than retried. Recording both
halves so a future reader neither re-opens the lever on *"but you only ran it once"* nor
repeats cycle-2's error of over-claiming impossibility.

---

## 4. Routing decision

| Item | Disposition | Basis |
| --- | --- | --- |
| Patch A — `internal/install`, 88 × `t.Parallel()`, 11 files | **Keep, accepted** | 226.191–235.124 s test-reported (wall 234/235/227 s) against the 480 s bar. Untouched by this cycle. |
| Patch B — atomicity partition, 2 files | **Keep as recorded, insufficient** | Exit 0 at 591.280/560.828/564.022 s. Replaces a hard `FAIL 603.701s` timeout. Does not meet 480 s. Not over-claimed. |
| 14th-file `references/info.md` fixture trim | **REJECTED as the timeout solution** | **Measured**, not modelled: count-one race exit 0 at 493 s, 13 s above a strict `<=480 s` bar, negative margin. TASK-260729-2afulh is `done` with an ACCEPTED verdict recording exactly this. |
| Production `saveJournal` / namespace revalidation | **Active route** | **TASK-260729-365r5r** — `prototype-savejournal-namespace-validation`, status `development`, same story `STORY-260720-3plyvy`. Its own AC already carries the O(P)-per-pass requirement and the `<=480 s`-or-explicitly-reject bar. |

**This task's 13-file patch is not widened.** No timeout changed, no case skipped, no assertion
weakened, no full-repository race gate run, no candidate mutated.

### 4.1 The route is now corroborated by the successor's own stored evidence

Read this cycle from `TASK-260729-365r5r_gate-evidence.tgz` (that task's artifacts, not
re-measured here):

| gate | exit | wall |
| --- | ---: | ---: |
| `atomicity` non-race, same-session baseline | 0 | **306 s** |
| `atomicity` non-race, prototype | 0 | **66 s** |
| `atomicity` race rep 1 / 2 / 3, prototype | 0 / 0 / 0 | **84 / 76 / 75 s** |
| `install` race, prototype | 0 | 72 s |

Against the same 480 s bar the production fix lands at **~5.7× headroom** on its worst race
run, where the test-only trim landed 13 s *over*. That is the difference in kind this cycle's
corrected arithmetic could not have predicted and the measurement makes obvious: every
test-side lever moves the **save count**; only the production fix moves the per-save
**O(P²) → O(P)** cost.

**State it accurately though — the successor is not accepted yet.** TASK-260729-365r5r remains
`development`: its `gate-lint-abs` exits **1** with three introduced `revive unused-parameter`
issues in `internal/transaction/namespace_pass_test.go`, and `gate-lint` exits **127** (bare
binary name absent from the driver's `PATH` — a missing-binary code, never a pass). Integration
is blocked on that rename plus a re-run of the four gates that compile the file, and on its own
reviewer cycle. The **routing** is validated; the successor's **acceptance** is not this task's
to claim.

The fixture trim retains standalone value as *evidence* for the successor — it demonstrates a
real, measured 14.55–17.14% staging-save reduction and a 12–17% wall-clock improvement — but it
is archived, not integrated.

---

## 5. Key aspects

1. **The reviewer caught a real undercount, and the corrected number is worse than the reviewer
   assumed.** R2-1 said "at least 8 per context"; it is exactly 8, and the denominator is 140
   rather than the 120 R2-1 charitably reused, so the corrected best case is 17.14% rather than
   20%. Still inside the required band. The finding is upheld on its own terms.
2. **Cycle-2's failure mode was arguing impossibility from an unmeasured model.** It reached a
   correct destination and dressed a guess as a structural proof. The destination survived;
   the reasoning did not. When a lever is cheap to measure, measure it.
3. **The corrected model is now calibrated against reality** (predicts 14.55–17.14% saves,
   delivered 13.91% wall-clock against the mean). That calibration is the durable output of
   this whole detour and belongs to TASK-260729-365r5r.
4. **The 480 s bar did its job.** A 493 s result with negative margin, inside a 32 s
   run-to-run spread, is not a fixed race gate. The bar refuses a result that would likely fail
   the real `./...` gate, where `cmd/curator` (557.779 s under race, unverified inheritance)
   overlaps for nearly the whole run.
5. **N = 20 is still unverified and is now contradicted by measurement** (`staged_targets` 12–18
   per transaction). Flagged so the successor re-derives rather than inherits — the quadratic
   conclusion is robust to N ∈ [12, 20], the specific syscall magnitudes are not.
6. **The successor makes the scale of the misjudgement visible.** The trim bought 12–17% and
   missed by 13 s; the production fix buys ~5.7× headroom (§4.1). Cycle-2's model was arguing
   about the wrong order of magnitude entirely — not because its asymptote was too low, but
   because it was modelling the only quantity a test-side change can move.
7. **Presentation reconciliation.** The 2afulh review verdict quotes the reduction as
   "12.5–17.6%". That range is over entry- and chunk-*count* reductions (chunks 16 → 14 =
   12.5%; entries 34 → 28 = 17.6%), not over saves. The save reduction is 14.55–17.14%. Both
   are correct statements about different quantities; noted so the two documents reconcile
   rather than read as a discrepancy.

---

## 6. Questions from the task description — answered

| Question | Answer |
| --- | --- |
| Is the `5C/(60+20C)` bound correct? | **No.** Retracted in full (§1.1–§1.3). Per-context removal is 8, not 5; the denominator model is wrong in both terms. |
| What are the actual before/after `StagingEntries`, chunks and save counts? | Measured and tabulated in §1.2: entries 24→20, 34→28, 26→22, 25→21; chunks 14→12, 19→16, 16→14, 16→14; staging saves 100→84, 140→116, 110→94, 107→91. |
| Can the separately scoped 14th-file trim defensibly reach 480 s? | **No — on measurement.** Focused count-one race exit 0 at 493 s, 13 s over, negative margin (§2). Not excluded by arithmetic; excluded by the gate. |
| If insufficient, is the production route supported? | **Yes.** TASK-260729-365r5r, `development`, carries the O(P)-per-pass requirement and the `<=480 s`-or-reject bar in its own AC (§4). No new task needed. |
| Do the 23 `saveJournal` call sites hold? | **Yes**, re-confirmed against source: 7 in `staging.go`, 16 in `engine.go` (§1.6). |
| Does the accepted 13-file patch change? | **No.** Not widened, not edited, not re-run this cycle. |

---

## 7. Evidence-honesty statement

- **No Go command of any kind was run this cycle.** No `go test`, probe, benchmark, `go build`,
  `go vet`, lint, Go-invoking helper script, or detached process. No `gofmt`. Fully compliant
  with `TASK-260729-rfrdfo_evidence-only-constraint.md`.
- **No product, prototype, or candidate file was edited.** The only writes this cycle are this
  document, the board artifacts it accompanies, the logbook entry, and read-only extraction of
  two stored evidence tarballs into `.temp/TASK-260729-rfrdfo/cycle3*/`.
- Process barrier, run twice as standalone scans for the record: both empty
  (`grep` exit 1, no match) → `BARRIER_OK`. It was not required, since no Go command was issued.
- **Partial probe artifacts from the cancelled `RUN-260729-5e854d` are excluded.**
  `.temp/TASK-260729-rfrdfo/measure/` (31-byte `probe.log`, 0-byte `vet.log`, scratch `tree/`,
  `gotmp/`, empty `dump/`) is invalid and **no figure in this document is drawn from it**. Its
  spawn log on the board is 0 bytes.
- All timings are quoted from stored gate artifacts, re-read this cycle:
  `TASK-260729-rfrdfo_evidence-bundle.tgz` `gates/*.exit` + `*.seconds` + `*.log` for the
  inherited arm, and `TASK-260729-2afulh_tester-evidence.tgz` `gates/` for the trim arm.
  Nothing was re-measured.
- The 493 s result is reported as **exit 0 and missing the bar** — both facts together. It is
  not presented as a failure of the test, nor as a pass of the gate.
- Race repetitions 2 and 3 of the trim arm **do not exist as evidence** and are not
  extrapolated from. §3 states the resulting statistical limit explicitly rather than papering
  over it.
- The staging inventory in §1.2 is TASK-260729-2afulh's measurement, taken with a
  measurement-only `zz_measure_test.go` present in scratch trees only and **absent from the
  14-file deliverable**; both A/B arms passed as standalone processes (61.271 s / 54.448 s).
  Independently verified by that task's reviewer. The save counts are exact arithmetic
  (`3·entries + 2·chunks`) over directly observed entry and chunk counts, not a claimed direct
  hook — because `saveJournal` is unexported and the literal file allowlist forbids a
  production seam.
- §4.1's successor figures are read from `TASK-260729-365r5r_gate-evidence.tgz` `gates/` and
  `gates-baseline/` `.exit`/`.seconds` files. That task owns them; this task did not produce or
  re-run them, and does not claim its acceptance.
- Source citations in §1.1, §1.5 and §1.6 were verified by reading
  `.temp/TASK-260720-jrrgw9/worktree` this cycle. The candidate was opened read-only and not
  written to.
- One figure differs from cycle-2 by rounding basis: the inherited spread is 32 s, which is
  5.4% of the worst run and 5.7% of the best. Cycle-2 quoted 5.4%. No conclusion moves.
