# Review brief — TASK-260916-dv7xv5 (verify-e1-e6-against-curator-and-launcher-main)

You are the independent reviewer of a **research** task. The producer's outcome
is the resource `verify-e-findings.md` on this task: a static, read-only
verification of findings E1–E7 against `relux-works/curator` `main` @
`80483355` and `relux-works/curator-agent-launcher` `main` @ `b34e1e27`.
The findings themselves are defined in the precondition resource
`security-audit-2026-09-spec-supplement.md` (E1–E6 plus the "Minor" residuals
that the task calls E7).

## What to verify (all read-only; change no code, run no tests unless a claim cannot be checked otherwise)

1. **Evidence check per finding (E1–E7).** For each row of the table in
   `verify-e-findings.md`, open the cited `file:line` sites at the pinned
   revisions (`git -C ~/Developer/ReluxWorks/curator/curator show 80483355:<path>` for
   curator, `git -C ~/Developer/ReluxWorks/curator/curator-agent-launcher show b34e1e27:<path>`
   for the launcher; both revisions are present in those checkouts — read them
   through `git show`, do not check anything out and do not modify the trees). Confirm the cited code says what
   the row claims and that the verdict (`confirmed` / `mitigated` /
   `partially confirmed` / `not applicable`) follows from it. Grep claims
   ("returns nothing") must be re-run as stated.
2. **Verdict soundness.** Flag any verdict that is stronger than the evidence
   (e.g. "mitigated" where the mitigation is partial) or weaker than it.
3. **Acceptance criteria of the task.** (a) the outcome resource carries the
   per-finding table — check; (b) *sibling story descriptions updated with the
   verdicts*: read the READMEs of `STORY-260916-ioemse` (E1), `-2d9coh` (E2),
   `-1i1gfo` (E3), `-2otjbn` (E4), `-73a5zg` (E5), `-wgt8vz` (E6),
   `-33vuzm` (E7) under `EPIC-260910-2hw1xb` and confirm each reflects the
   verdict recorded in the "Consequences for the sibling stories" section. If
   a story description does not carry its verdict, that is a rework finding,
   not an accept.
4. **Scope discipline.** The task is research: no code or test changes are
   expected anywhere. Do not start or suggest starting any remediation story;
   remediation is a separate goal.

## Checklist

The task checklist has three items mirroring the acceptance criteria. The
producer handed off before the checklist existed, so YOU check each item
(`task-board m 'check_item(TASK-260916-dv7xv5, item=N)'`) only once you have
verified it yourself; leave an item unchecked if it fails and record why.

## Verdict

Record your verdict as a task-scoped outcome resource named
`TASK-260916-dv7xv5_review-verdict-rev1.md` with: per-finding check result
(agree / disagree + why, with file:line), the acceptance-criteria check,
and the final verdict — `accept` or `rework` with the concrete findings the
producer must address. Then hand off with the reviewer role's normal status
transition (`done` on accept, `to-dev`/`analysis` on rework). Do not set
`blocked` unless a human-only decision is genuinely required.

Budget: this is a bounded read-only check of seven rows and seven READMEs;
finish in one pass.
