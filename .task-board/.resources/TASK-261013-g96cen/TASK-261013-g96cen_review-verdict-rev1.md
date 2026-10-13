# TASK-261013-g96cen — reviewer B verdict, CR revision 1: CHANGES REQUESTED

Base `a01542857678108a77d5e1b1f051989b00384f1d`, candidate tree `d71ef82972d8c5d40d681e26f13c974faec55ce6`. Docs only, read-only review; no tests or builds were run. My checks 1–8 were done and this verdict was formed before I read `gw-CS7-review-A.md`; reconciliation with A follows.

## Verdict

The transfer, export, cleanup, base-binding, mode, exec-input and Muse changes are applied and meet the findings. They stay inside the brief's bounds. Reading the diff, I found these things:
- Standing discard stays off.
- No wire field, refusal code or value is invented.
- Unknown is never treated as absent.
- There is no automatic retry after an ambiguous effect.
- The per-act waiver is classified as widening and added to the disposition's operator authority boundary.
- ALLSTOP, revocation and restore never reopen implicitly.
- The diff touches only 20 in-scope files.
- Added lines carry no private identifiers, and `UTF-8` is not mangled.

One defect is a P1: the edits to Decisions 0010, 0013 and 0019 replace correct citations with wrong ones. Because of it I request changes (`to-dev`).

## Defects that must be fixed

### B1 — P1 (regression): the "Decision 5" qualification mis-attributes in-document citations in Decisions 0010, 0013 and 0019

- **What the diff did.** It rewrote about 20 bare "Decision 5" citations in these three files into "agents-management Decision 5 (profile composition and interactive launch)" or "agents-management Decision 5". It added a revision note to each file saying "No behavior changes".
- **What is actually true.**
  - In `decisions/0010-agent-environment-profiles.md`, "Decision 5" is that document's own §5, "Current profile, composition, and switching". Its mentions of "Decision 5 composition", "scoped switching" and "the operator warning in Decision 5" all point to that own section.
  - In `decisions/0013-execution-ownership-and-launch-plans.md`, "Decision 5" is that document's own §5, "`LaunchModeInteractive` in `agents-management`". Examples are "(Decision 5)" in the acceptance list next to "(Decision 6)" and "Decisions 3, 4, and 7", and "Decision 5 forbids it in the interactive plan". Its cross-document form is already explicit: "Decision 0010 Decision 6".
  - In `decisions/0019-fragment-consumers-and-one-construction-site.md`, "Decision 5" means Decision 0013's §5. The text says "amends Decision 0013 Decisions 1, 5, 6.3 and 6.5", and "What changes in Decision 0013" lists "Decision 5".
- **Why it is a regression.** The base citations were correct in-document references. The new label turns three different decisions into one invented "agents-management Decision 5 (profile composition and interactive launch)". That label matches neither §5 title and points to a decision record that does not exist in this repository. The revision notes also state "no behavior changes" and describe the references as being to a separate decision, which is false. The collision it was meant to fix is only between "Decision 5" and "Decision 0005", and the Decision 0005 retitle already handles that.
- **Requested change.**
  - Revert the in-body substitutions in 0010, 0013 and 0019.
  - Where qualification is wanted, qualify by the true owner:
    - "this document's Decision 5" in 0010 and 0013;
    - "Decision 0013 Decision 5" in 0019.
  - Alternatively, report in the gaps file that the bare citations are unambiguous self-references and needed no change, and keep only the Decision 0005 retitle.
  - Correct or remove the three revision notes accordingly.
  - Correct the gaps statement ("qualified all bare Decision 5 references ... as the separate agents-management composition/interactive-launch decision"), which is untrue as written.

## Notes (P2/P3), to be folded into the rework where cheap

- **B2 (P2).** Two normative uses of the audit-row label remain in this repository. These are CIP-0010 "The intake release/grammar belongs to R-LR1 and Decision 0021's amendment" and CIP-0011 "Intake schema/version publication and qualification remain R-LR1 work". The disposition now says audit labels are not rules, so these two copies contradict it. Replace each with the launcher-intake-lane link, as the disposition already does. The gaps file should report the CIP-0011 change if made.
- **B3 (P2).** Decision 0021 still states the blanket form of the naming rule ("versioned platform wire identifiers carry the repository name"). The disposition's scoped sentence now contradicts it. Either add a scoped cross-reference there or list it truthfully in the gaps file as a residual copy.
- **B4 (P2).**
  - Stale Muse status remains in CIP-0003 ("Muse subscription support remain owner questions").
  - In CIP-0010's summary line ("Muse subscription channel ... retain the explicit options below"), the options "prioritise or defer" no longer exist. Update both sentences in the same three-way wording, or report them.
- **B5 (P3).** In the disposition, the D-WIRE-NAMES scope sentence is written as the decision itself. The finding asks for an owner or contract clarification of its scope. State in the gaps file that the sentence is the proposed scoping pending the external naming index, so it is not read as a new owner decision.
- **Reconciled with reviewer A.** I confirm A1 and A2 and add them to the required fixes.
  - A1 (P2): `donor-layout.puml:84` ("fresh materialization each round") and `donor-lifecycle.puml:88` ("fresh B/C materialization") contradict same-workspace refusal repair. Qualify both summaries as "except an admitted repair round".
  - A2 (P3): state that the contradiction comparison applies to the same unchanged quiesced tree, not across an admitted repair. This is in CIP-0009 D5a's Export outcomes paragraph and in the transfer figure.
  - A swept the Decision-5 edits and judged them acceptable; I disagree, per B1. A raised no P0/P1.

## Check summary

| Check | Result |
|---|---|
| 1 Findings applied | 8 of 8 are addressed in the sections named in the gaps file, except as noted in B1–B4 and A1–A2. |
| 2 Bounds | Standing discard unchosen and off. No claim of implementation or qualification. Result schema fields unchanged. |
| 3 Names | Shared names are unchanged. No invented fields, codes or values. Unpublished formats are listed as open questions (CIP-0009 open questions 11, 14, 16). |
| 4 Security properties | Preserved. Unknown stays pending. A lost reply reconciles and never relaunches. Cleanup consent covers ordinary expiry only. There is no launch of a retiring generation. A waiver is a widening act. Holds, ALLSTOP and restore are explicit. |
| 5 Figures | The transfer, turn, retire and summary figures match the text, apart from A1. All are linked in the README. |
| 6 Gaps file | Accurate, except the B1 claim and the figure-consistency claim (A1). |
| 7 Hygiene | Clean. The added lines carry no F-numbers (the one `F-16` is baseline text), seat names, event ids, local paths or board links. |
| 8 Collateral | None beyond B1. `UTF-8` is untouched. |
