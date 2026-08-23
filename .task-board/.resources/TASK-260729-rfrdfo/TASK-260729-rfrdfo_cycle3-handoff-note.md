# TASK-260729-rfrdfo — cycle-3 handoff note

## Status route

`task-board handoff TASK-260729-rfrdfo --role researcher` was run and **refused**:

    cannot hand off TASK-260729-rfrdfo: unchecked checklist items [3 12 13 14]
    (Approved install and atomicity focused commands pass with count=1 and no timeout override;
     Implementation matches AC; Solution fits project architecture; Tests green):
    handoff evidence missing

Items 12, 13 and 14 are unchecked **because review verdict cycle 2 item 5 ordered exactly
that**: keep them unchecked until the corrected routing evidence is reviewed. Producing that
corrected evidence is this cycle role. Checking them to satisfy the verifier would invert
the instruction and manufacture handoff evidence that does not exist.

Item 3 is also left unchecked. The recorded commands did run green (13/13 gate `.exit` files
contain 0, every `go test` carries `-count=1`, none carries a timeout override or `./...`, no
`DATA RACE`), and the cycle-2 reviewer independently confirmed that. But the same reviewer has
deliberately left this item unchecked across two cycles, evidently reading it against the
sanctioned `<=480 s` condition that the atomicity command misses. That reading is the
reviewers to make, not mine to overturn in order to unblock a gate.

The task was therefore routed with an explicit `set_status(TASK-260729-rfrdfo,
status=to-review)` — the documented end status for the researcher role — rather than by
checking items to clear the verifier. Items 3, 12, 13 and 14 are the reviewers to resolve.

Items **18** (fact-checking performed — claims verified, sources cited) and **20** (all
questions from the task description answered) were checked this cycle and are backed by
`TASK-260729-rfrdfo_cycle3-corrected-routing.md` sections 1, 2 and 6.

## Cycle-3 scope compliance

- **No Go command of any kind was run** — no test, probe, benchmark, build, vet, gofmt, lint,
  Go-invoking helper script, or detached process. Per
  `TASK-260729-rfrdfo_evidence-only-constraint.md`.
- **No product, prototype, or candidate file was edited.** Both worktrees verified untouched
  after the cycle (source and prototype `fixture_test.go` unchanged, 14045 bytes, mtime
  2026-07-28 16:53).
- Two-scan process barrier run for the record at start and end: both empty, `BARRIER_OK`.
  Not required, since no Go command was issued.
- Partial probe artifacts from cancelled `RUN-260729-5e854d`
  (`.temp/TASK-260729-rfrdfo/measure/`) are excluded; no figure is drawn from them.

## Artifacts this cycle

| Artifact | Action |
| --- | --- |
| `TASK-260729-rfrdfo_cycle3-corrected-routing.md` | **new** — corrected arithmetic, measured rejection, routing |
| `TASK-260729-rfrdfo_cycle2-routing-and-evidence.md` | **updated** — retraction notice + inline retraction markers at sections 2.2, 2.3, 3, 5 |
| `TASK-260729-rfrdfo_logbook-entry.md` | **updated** — corrected entry superseding the cycle-2 version |
| `TASK-260729-rfrdfo_cycle3-research.md` | duplicate of the corrected-routing doc; `resource delete` is blocked by a pre-existing board validation error (`unknown status: todo`) |
| `.research/260729_install-race-timeout-corrected-routing.md` | researcher deliverable on disk |
| `.research/260729_rfrdfo-cycle2-routing-and-evidence-preservation.md` | resolution banner added under the existing 3dr6hw cycle-7 correction |
| `LOGBOOK.md` entry 2015 | corrected finding recorded |

## What the reviewer is being asked to adjudicate

1. That the `5C / (60 + 20C)` bound and its derived claims are retracted, and that the
   corrected per-affected-context figure of **8** (equivalently `21 -> 13` staging saves per
   context target) is right.
2. That the fixture-only trim is now rejected **on measurement** — count-one race exit 0 at
   **493 s** against a strict `<=480 s` bar, 13 s over with negative margin — and not on the
   withdrawn arithmetic.
3. That the stated limit of that rejection is acceptable: it proves failure to demonstrate
   margin, not impossibility, and the single valid repetition is sufficient because the
   acceptance predicate is universally quantified over repetitions.
4. That routing to `TASK-260729-365r5r` is supported, without this task claiming that
   successors acceptance.
