# Review brief — TASK-260916-dv7xv5, review round 4 (rev4)

You are the independent reviewer of a **research** task, round 4. Round 3
(`TASK-260916-dv7xv5_review-verdict-rev3.md`) agreed with every verdict,
citation and sibling description and returned rework for exactly two
transcript defects: one missing leading tab on the E3 literal-output line for
`internal/skillspec/parse.go:697` and the wording "profile.go:260 is the
usage-error return" (return is `:259`, `:260` the closing brace). The
orchestrator applied both as **`verify-e-findings-rev4.md`** (outcome resource;
= rev3 + those two edits, nothing else — see the "Rev4 answers verdict rev3"
header). Review **rev4**; rev1–rev3 artifacts stay attached only as history.
Scope of this round: confirm the two corrections and that nothing else changed
against rev3 (`diff` the two resources); do not re-open settled agreements.

The findings E1–E6 (+ the "Minor" residuals the task calls E7) are defined in
the precondition resource `security-audit-2026-09-spec-supplement.md`.

## What to verify (all read-only; change no code, run no tests)

1. **Each rev3 correction is closed in rev4.** Check that the E3 transcript
   block matches your own capture byte for byte (three leading tabs on
   `parse.go:697`) and that the E1 introduction now reads `:259` return /
   `:260` closing brace; check that `diff verify-e-findings-rev3.md
   verify-e-findings-rev4.md` shows only the header paragraph and those two
   edits.
2. **Evidence check per finding (E1–E7).** Rev3 already agreed on every row;
   spot-check only that the rev4 rows are unchanged. For any row you re-open, open
   the cited `file:line` sites at the pinned revisions
   (`git -C ~/Developer/ReluxWorks/curator/curator show 80483355:<path>`,
   `git -C ~/Developer/ReluxWorks/curator/curator-agent-launcher show b34e1e27:<path>`;
   read through `git show`, do not check anything out and do not modify the
   trees). Confirm the cited code says what the row claims and that the verdict
   follows from it. Flag any verdict stronger or weaker than its evidence.
3. **Acceptance criteria.** (a) the outcome carries the per-finding table with
   both pins and `file:line` evidence; (b) the READMEs of
   `STORY-260916-ioemse` (E1), `-2d9coh` (E2), `-1i1gfo` (E3), `-2otjbn` (E4),
   `-73a5zg` (E5), `-wgt8vz` (E6), `-33vuzm` (E7) under `EPIC-260910-2hw1xb`
   now record the verified verdict, the pins and the narrowed scope (read them
   with `task-board q 'get(STORY-…) { overview }'` or the README files);
   (c) no code or test changes anywhere — research only.
4. **Scope discipline.** Do not start or suggest starting any remediation
   story; remediation is a separate goal.

## Checklist

The task checklist (13 items) is fully checked. Item 13 "Tests green" is a
role-baseline item the runtime re-adds on every spawn; the orchestrator checked
it on the recorded basis (task notes) that no test suite is part of this
read-only research deliverable and the required evidence re-verification (grep
re-runs at the pins, exit 0) passed. **Do not uncheck it and do not treat it as
a suite attestation** — state "Tests green: not applicable (orchestrator
basis)" in the verdict instead. Uncheck any other item only if you find it
unsatisfied and say why.

## Verdict

Record the verdict as a task-scoped outcome resource named
`TASK-260916-dv7xv5_review-verdict-rev4.md`: per-rev3-correction closure status,
per-finding evidence check (agree / disagree + why, with file:line), the
acceptance-criteria check, and the final verdict — `accept` or `rework` with
the concrete findings the producer must address. Then hand off with the
reviewer role's normal transition (`done` on accept; `to-dev`/`analysis` on
rework). Do not set `blocked` unless a human-only decision is genuinely
required.

Budget: a bounded read-only check of one 7-row table, the rev1 findings and
seven READMEs; finish in one pass.
