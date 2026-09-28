# TASK-260906-1xbrz6 rework 1 → revision 2 (orchestrator brief, binding)

Verdict on revision 1: CHANGES_REQUESTED (`TASK-260906-1xbrz6_review-verdict-rev1.md`). The four
normative sentences themselves are fine — the reviewer confirmed they agree on diagnostic,
exclusion and recovery, and that no conformance family enumerates the carriers. What is wrong is
the PIN that guards them, plus two recovery claims that need substantiating. Keep the five-member
carrier set; do not widen it.

## F1 (major) — the pin checks token presence, not meaning
`tools/validate.py:4209` and `:4284` only require certain tokens to appear. Replacing "they fail
closed" with "they fail open", or "operations are outside this closed set" with "operations are
inside this closed set", both PASS `validate_takeover_closed_set_text`. Contradictory normative text
passes the gate. Fix: pin the actual exclusion + fail-closed sentence — including its subjects
(import activation / the §9.4 global operations) — in all four locations as a
**whitespace-normalised exact sentence**. That is simpler and stronger than semantic parsing, and it
removes the sentence-splitting ambiguity of F2 at the same time. Add negative tests for PREDICATE
changes (fail closed → fail open; outside → inside; subject swapped), not just token deletion.

## F2 (moderate) — section and sentence boundaries are not enforced
`:4245` and `:4253` match headings by prefix split, so `### 9.5 Onboarding` → `### 9.5
Onboarding-renamed` still passes; sentence splitting recognises only uppercase/backtick successors,
so splitting the exclusion at "closed set. on" also passes. Anchor COMPLETE heading lines, delimit a
section at the next same-or-higher heading, and test sentence movement and splitting independently.

The reviewer's read-only reproduction (all four mutations SURVIVED on revision 1) is in the verdict —
turn each of those four into a committed negative test that now FAILS the function.

## Recovery claims — substantiate or record a decision packet
Two sequences in your gap analysis are asserted, not established:
1. **Import retry.** `profile sync/use --takeover` only write paths represented by an installed
   lock, and re-running `profile import` with an already-installed name meets
   `profile_import_name_taken` (§9.7). So the recovery can complete ACTIVATION, but a literal rerun
   of `profile import` is not shown to succeed. Distinguish "retry the activation" from "repeat the
   import" in the text, and state which one the recovery sentence means.
2. **Global add ordering.** The universal global-add recovery claim depends on the newly extended
   lock being published BEFORE a surface conflict is raised; §9.4 alone does not spell out that
   ordering. Either cite the transaction contract that fixes it, or record the open question.

For each: if the normative text (transaction contract, §9.1/§9.4/§9.6/§9.7) settles it, cite the
exact clause and make the sentences precise. If it does not, write the brief-required **decision
packet** (constraint, evidence, alternatives, recommendation, exact question) as a resource — do
not invent behaviour and do not widen the carrier set.

## Rules
Continue from the revision-1 tree (no checkout/clean/stash). Run the validator's own unit rows and
the four new negatives with real exit codes; the board validation command runs once at handoff.
Append a "Revision 2" section to `TASK-260906-1xbrz6_results.md`, then
`task-board handoff TASK-260906-1xbrz6 --role developer`.
