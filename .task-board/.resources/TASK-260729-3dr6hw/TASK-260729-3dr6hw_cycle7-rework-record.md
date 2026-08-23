# TASK-260729-3dr6hw — cycle-7 rework record

Date: 2026-07-29
Role: researcher
Input: `TASK-260729-3dr6hw_review-verdict-cycle-6.md` (changes requested → `analysis`)
Standing artifact after this cycle: `.research/260729_install-race-timeouts.md`, SHA-256
`3826d919da749b7c7bbc2539c9a3b90b33a772b828050517b9142af6cdc33539`, 1903 lines
(cycle-6 copy: `c4ba2c0ffb1c6206a1615a7fdf35242a84bd6ecaf8d325c9f4e1a38b47a09722`, 1690 lines)

**No Go command, build, vet, test, or tooling invocation was executed in this cycle.** The task
boundary is "diagnose without running Go tests and without edits"; the only writes were to
`.research/`, `.temp/TASK-260729-3dr6hw/` and the board. The candidate worktree
`.temp/TASK-260720-jrrgw9/worktree` and the prototype worktrees were read only.

---

## 1. Preservation proof for required item 1

The verdict requires the accepted verifier inventory, Patch A evidence, all six measured focused race
results, the candidate-integrity protocol, the assertion map, and the "no safe `./...` margin"
finding to be preserved.

They are, byte-for-byte. The slab from `## 1. Exact failing tests…` up to the `## 11.` heading — that
is, **§1 through §10 inclusive**, which contains every item on that list — was extracted from both
the cycle-6 board copy and the cycle-7 file and hashed:

| Copy | §1–§10 length | §1–§10 SHA-256 |
| --- | ---: | --- |
| cycle-6 board resource `TASK-260729-3dr6hw_install-race-timeout-diagnosis.md` | 85,713 bytes | `d3c0f86dda368d424d296542f4a0b073eae549d9ffe6988f85e61b4d80ea2105` |
| cycle-7 `.research/260729_install-race-timeouts.md` | 85,713 bytes | `d3c0f86dda368d424d296542f4a0b073eae549d9ffe6988f85e61b4d80ea2105` |

Identical. The whole-document diff (`TASK-260729-3dr6hw_cycle7-diagnosis.diff`, 506 lines) contains
exactly **4 hunks**, all in the header region and in §11–§12: the revision/verdict block, the new
"What changed in cycle 7" table, two rows of the *cycle-6* change table marked as withdrawn, and the
two rebuilt sections. Nothing in §1–§10 moved.

---

## 2. Required finding R6-1 — section 11's save-count bound is invalid

**Accepted in full. Every one of the verdict's four source claims was re-verified against the
candidate worktree before the rewrite, and every one is correct.**

| Verdict claim | Verification | Verdict |
| --- | --- | --- |
| `staging.go:18-56` iterates `target.StagingEntries`, not targets, and saves three times **per staging entry** (`:26`, `:33`, `:56`) | `staging.go:18-19` is `for target.StagingIndex < len(target.StagingEntries) { entry := target.StagingEntries[target.StagingIndex]`; the three `engine.saveJournal(journal)` calls sit at `:26`, `:33`, `:56` inside that loop | **correct** |
| A non-empty file adds two saves per 32 KiB chunk (`:141`, `:161`) | `copyStagingFile` (`staging.go:118-176`), `stagingCopyChunkSize = 32 * 1024` (`:12`), both saves inside `if count > 0` | **correct** — and a *zero-byte* file adds none, because the first `Read` returns `0, io.EOF` |
| `captureRemovalEntries` walks and records both directories and files (`journal.go:506-533`) | `filepath.Walk` at `journal.go:507`; `entry.Kind = "directory"` for `info.IsDir()`; the walk **root** is recorded too, its relative path normalised from `"."` to `""` at `:518-522` | **correct** |
| `fixture_test.go:83` creates `references/info.md`; removing that only child removes **two** staging entries, `references/` and `references/info.md` | `fixture_test.go:83` is `e.write(dir, "references/info.md", "ref")`. `e.write` (`:63-73`) `MkdirAll`s the parent, so `references/` exists only because of that line. `whitelist.CopyContext` skips an include root whose `os.Stat` fails (`whitelist.go:57-59`), so the directory does not reach the context tree either | **correct** |
| ⇒ at least `3 × 2 + 2 × 1 = 8` saves per affected context | reproduced independently: 21 → 13 staging saves per context target | **correct** |

### 2.1 What was withdrawn

- The `3 × targets + 5 × files` save model. It charged per target, not per entry, and had no term for
  directory entries.
- The **15**-save numerator.
- The **120–210** denominator.
- The **7.1–12.5 %** ceiling and the derived 490.7–517.4 s table.
- The conclusion that the fixture trim is *bounded below* the required saving.
- The conclusion that the test-only boundary is **exhausted** (also removed from §12). It is replaced
  by a narrower, measured statement: the best measured test-only configuration is **492.231s** on one
  repetition, 2.5 % above the bar and inside the package’s own 5.4 % spread.

### 2.2 What replaced it

§11 is rebuilt from the source formula
`saves(install) = O(phases) + O(targets) + Σ_t (3·E_t + 2·C_t)`, with `E_t = len(StagingEntries)` and
`C_t` the 32 KiB chunk count over `t`'s file entries. Concretely established:

- **Per context target:** entries 5 → 3, chunks 3 → 2, staging saves **21 → 13**, **−8**, **−38.1 %**.
- **Per chain (patched Patch B shape, `baseline → 1 injection → final`):** project-hybrid touches 8
  context targets ⇒ **−64 saves**; global touches 6 ⇒ **−48 saves**.
- **Across the sweep:** `7 × 64 + 5 × 48 = ` **−688 saves**, before the four non-sweep tests and
  `activation_test.go`.
- **Direction of the correction:** on cycle 6's own single-install basis the numerator goes 15 → 24,
  **×1.6**, which on cycle 6's own best-case denominator is `24/120 = 20 %` — above the 14.4–18.8 %
  it called the requirement. This is the verdict's own arithmetic, reproduced.

### 2.3 Two further errors found while rebuilding, both in the *unfavourable* direction for cycle 6

1. **A context target does not carry four fixture files.** `csk-skill.json` is not in
   `whitelist.IncludeRoots` (`whitelist.go:20-23`) and `scripts/` is excluded twice over
   (`Commands` non-empty ⇒ `includeScripts=false` at `targets.go:59-62`; `runtime_roots:["scripts"]`
   ⇒ `excludeRoots` at `:65`). The tree is `SKILL.md`, `references/info.md`, `.csk-install.json`
   (`marker.go:27`, written at `targets.go:85`). `locale.Render` (`targets.go:70`) adds nothing —
   the fixture has no `locales/`, so `analysis.LocaleToRender == ""` and it returns at
   `locale.go:123-125`.
2. **The trim does not remove "three files once".** It removes one file *and one directory* from
   **every** context target of **every** installation in **every** chain — which is where the ×1.6
   and the 688 come from.

### 2.4 Required item 3 — the evidence exists; the trim is measured, not unmeasured

This cycle's first pass concluded the trim was **unmeasured** and specified an A/B measurement for a
future producer. That was wrong, and it was wrong for the reason this task has now been caught on
three times: it was written before reading `LOGBOOK.md`.

`LOGBOOK.md` entry **2026-07-29 1947** records **`TASK-260729-2afulh`**
(`prototype-atomicity-fixture-trim`, parent `STORY-260720-3plyvy`, status **`done`**, review
**ACCEPTED**), which applied exactly the §4.3 trim as a declared 14th allowlisted path, instrumented
the staging inventory in both arms, and ran the focused gates. Read this cycle from
`.temp/TASK-260729-2afulh/evidence/measurement-raw.txt`, `gates/*.exit`, `gates/*.seconds`,
`gates/DRIVER-STOP-REASON` and its own review verdict:

| scenario / phase | entries | chunks | staging saves `3E+2C` | Δ | reduction |
| --- | ---: | ---: | ---: | ---: | ---: |
| project baseline | 24 → 20 | 14 → 12 | 100 → 84 | −16 | 16.00 % |
| project upgrade | 34 → 28 | 19 → 16 | 140 → 116 | −24 | 17.14 % |
| global baseline | 26 → 22 | 16 → 14 | 110 → 94 | −16 | 14.55 % |
| global upgrade | 25 → 21 | 16 → 14 | 107 → 91 | −16 | 14.95 % |

Every Δ is exactly `8 × (context targets in that phase)` — **the derivation in §2.2 above, produced
before the measurement was located, is confirmed row for row**. So is the entry list itself; the
measured `MANIFEST` dump is
`directory:,directory:references,file:.csk-install.json,file:SKILL.md,file:references/info.md`
before and `directory:,file:.csk-install.json,file:SKILL.md` after — the derived 5-entry and 3-entry
trees element for element, including the absence of `csk-skill.json` and `scripts/`.

Gates: `gofmt` 0, `build` 0, `vet` 0, focused non-race **0 / 272.580s** (control 285.434s, −4.50 %),
one valid focused race gate **real exit 0 / 492.231s**, no `DATA RACE`. Repetition 2 was SIGTERM'd
with no `.exit` file and is correctly excluded; repetition 3 was never run.

**Disposition, restated on evidence rather than arithmetic:** 492.231s misses §7's 480s bar by
**12.231s** and clears the 600s alarm by **107.769s** (Patch B alone: 8.72s). One repetition against
a 5.4 % spread is not a demonstration either way, which is why §11.5 now asks for repetitions 2 and 3
**on the tree `TASK-260729-2afulh` already produced** rather than for a new prototype.

**What this changes about cycle 6:** its conclusion ("not enough") survives; its arithmetic does not.
The measured 14.55–17.14 % reduction is *inside* the 14.4–18.8 % band cycle 6 called required, so the
trim fails on wall clock, not on work reduction. The 7.1–12.5 % ceiling understated it.

---

## 3. Required finding R6-2 — successor routing

Read from the board this cycle:

```
{"id":"TASK-260729-365r5r","name":"prototype-savejournal-namespace-validation",
 "parent":"STORY-260720-3plyvy","status":"development",
 "title":"TASK-260729-365r5r: prototype-savejournal-namespace-validation"}
```

"New story" is withdrawn everywhere it appeared. §12 and the §12.1 routing table now name the task,
its slug, its parent story and its `development` status, and note it is working in
`.temp/TASK-260729-365r5r/worktree` off the rfrdfo prototype state so the 480s bar is quoted against
the same tree that produced the 560.828–591.280s baseline. Its own acceptance criteria — `O(P)`
identity/resolution reads per pass instead of `O(P²)` pairs, fail-closed validation preserved before
any mutation, and focused evidence that either demonstrates a ≤480s atomicity margin **or explicitly
rejects the prototype** — are quoted rather than paraphrased.

Its in-flight gate results are deliberately **not** quoted in the diagnosis: `.temp/TASK-260729-365r5r/gates/`
was still being written while this cycle was authored, and quoting a partial run is the failure mode
cycle 5 was already corrected for.

---

## 4. Required item 4 — the margin AC

Stated in the document header and as its own paragraph in §12:

> this task's own acceptance criterion — "expected margin below 10 minutes for both packages under
> race", operationalised as §7's 480s focused pass condition — remains **UNSATISFIED** for
> `internal/install/atomicity`. It is satisfied for `internal/install`.

---

## 5. Required item 5 — checklist

**Item 15, "Tests green", stays unchecked.** No Go command was run in this or any prior cycle of this
task; the focused atomicity package is exit-0 in three downstream tasks but above this diagnosis's own
bar in all of them; and the contended `./...` race gate has not passed.

**Item 13, "Implementation matches AC", is unchecked this cycle.** It was checked. The task AC asks
the outcome to "recommend the smallest test-only patch with expected margin below 10 minutes for both
packages under race", and §7 operationalises that as the 480s focused pass condition. §12 now states
plainly that the criterion is **unsatisfied** for `internal/install/atomicity`, so leaving item 13
checked would contradict the document's own conclusion. It is left for the reviewer to adjudicate
rather than asserted either way: the *diagnosis* clauses of the AC (exact failing tests and timings,
assertion/invariant mapping, non-overlapping validation commands, independent review) are all met;
the *margin* clause is not, and cannot be by any recommendation this task is scoped to make.

---

## 6. Repository consistency

A grep of `.research/` found the withdrawn arithmetic in **three** documents, not one. The cycle-4
verdict (R4-1 item 3) already established that a correction has to reach every copy, so all three are
handled:

| Document | What it carried | Action |
| --- | --- | --- |
| `.research/260729_install-race-timeouts.md` (standing artifact) | §11's whole model; and two rows of the *cycle-6* change table restating it as current | §11 rebuilt around the measured evidence; both cycle-6 rows struck through and marked **WITHDRAWN IN CYCLE 7** in place, so the historical record survives without reading as current |
| `.research/260729_install-race-timeout-diagnosis.md` (superseded-duplicate pointer) | the 7.1–12.5 % rejection of P3 and "the test-only boundary is exhausted" | corrected in place; P3 is now **decided by measurement** (492.231s, one repetition, 12.231s over the bar) and the exhaustion claim is replaced by the measured 2.5 %-above-bar statement |
| `.research/260729_rfrdfo-cycle2-routing-and-evidence-preservation.md` (**another task's** artifact) | §2.2 reproduces the withdrawn table "to the decimal", and §2.3 builds a `5C / (60 + 20C)` "stronger bound" on a *four files per context* premise that is also wrong | **non-destructive correction banner** added at the top, attributed to this task and cycle, naming both errors, quoting the `TASK-260729-2afulh` measurement that supersedes them, and listing the surviving parts. Its text is not rewritten — it belongs to `TASK-260729-rfrdfo` |

Two cross-task board notes were written so the affected owners see the correction without having to
re-read `.research/`:

- **`TASK-260729-rfrdfo`** (`analysis`) — the two source errors, the corrected numerator, the
  withdrawal of its §3 conclusion, and the explicit list of its findings that stand.
- **`TASK-260729-365r5r`** (`development`) — that it is now named exactly in §12/§12.1 as the open
  successor rather than "a new story"; that the fallback lever is measured at 492.231s rather than
  closed; and that per-save cost is linear in path depth (`EvalSymlinks` plus two
  `existingNamespaceAncestor` walks per path per pass, none memoised), which bears on its design.

After the edits, the strings `7.1-12.5` / `7.1–12.5` and `boundary is exhausted` appear in `.research/`
only inside passages that record them **as withdrawn**, plus the rfrdfo document's own now-bannered
§2.2–§2.3 and §3.

---

## 7. Board artifacts written this cycle

| Name | Type | Content |
| --- | --- | --- |
| `TASK-260729-3dr6hw_install-race-timeout-diagnosis.md` | outcome (updated) | the cycle-7 standing diagnosis, byte-identical to `.research/260729_install-race-timeouts.md` |
| `TASK-260729-3dr6hw_cycle7-rework-record.md` | outcome (new) | this record |
| `TASK-260729-3dr6hw_cycle7-diagnosis.diff` | outcome (new) | the exact 4-hunk diff of cycle 7 against the cycle-6 board copy |
| `TASK-260729-3dr6hw_race-timeout-diagnosis.md` | outcome (updated) | superseded pointer, withdrawn arithmetic corrected |
