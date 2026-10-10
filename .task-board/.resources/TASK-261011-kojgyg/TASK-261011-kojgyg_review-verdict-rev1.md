# TASK-261011-kojgyg — deciding review (reviewer B), CR rev1: ACCEPT

Delta reviewed: a60fd341 -> tree 1486fc81, one path (`decisions/0021-sessions-enter-through-curator-run.md`, +117/-15). The working tree equals the candidate tree. Checks 1-5 were done and this verdict drafted before reading reviewer A's record; A was then reconciled (below).

## Checks
1. Rows. R-LR1 in D1 assigns four A-rows (A09.1, A09.6, A09.10, A10.9); all four are fixed to the table's values or to a recorded owner decision. I additionally compared 14 table rows against the changed text: N6, P1, P2, P6, L1, L2, I2, I3, C3, C9, X1 (via D-PATH), W2 (via D-WIRE-NAMES), W3, I4, plus C4. Sizes (48/64 KiB, 80 KiB/5 s, 96 KiB/5 s, 128 KiB, 64 KiB broker), FD3-only plan, receipt-before-drop, setup-token via CLAUDE_CODE_OAUTH_TOKEN with no file/fallback/refresh, and the D-PATH order all match. W3, I4 and C4 stay open with the table's own two options each.
2. Owner decisions. D-EXECUTOR rails (no compose/profile resolution under the agent account, verify-and-execute only, `curator-run exec --plan-fd N` as an example entry, no network except the broker, one shared library, separate binary only a later option, first consumer = curator-run) are applied as stated. D-PATH, D-WIRE-NAMES and the first-consumer decision are applied exactly. No recommendation became a decision. The activation gate is stated as a gate, not a claim of qualification.
3. Diagrams. Decision 0021 has none; R-LR1 names none in this repository (cips/diagrams hold only donor-side figures, which belong to R-CS1). The alignment says so. Nothing to correct.
4. Alignment outcome. 4/4 rows mapped; the "8/8 amendment boundary rows" claim is true (the table has 8 rows); the stated scope (one file, uncommitted) is true; the validation claims are limited to syntax/packaging.
5. Docs only; no first names, local paths, session links or board-resource links in the text. I reran `git diff --check` on the candidate file: exit 0. No tests or builds run.

## Reconciliation with reviewer A
A reports 0 blocking findings and one informational note (A-N1, the amendment authorizes a mode but does not publish or qualify the intake). I agree; no P0/P1 from A remains to be checked.

## Findings (no P0, no P1; notes under tb-R226)
```yaml
findings:
  - id: B-N1
    severity: robustness
    severity_reason: "P2 note. The 'Network and provider channel' row adds that the launched harness has a separately admitted egress policy. D-EXECUTOR says only 'no network except the broker'; the clarification is the author's reading (reasonable, since the harness must reach its provider) but it is not sourced and not marked as an interpretation."
    location: "decisions/0021-sessions-enter-through-curator-run.md:231"
    action: "In the launcher intake amendment, state the executor-versus-harness egress boundary explicitly and have the owner confirm it."
  - id: B-N2
    severity: note
    severity_reason: "P2 note. The text cites D-EXECUTOR/D-PATH/D-WIRE-NAMES and 'the supplied owner decisions of 2026-10-10/11' by task labels; the source file is a board precondition with no durable citable location, and the table/audit links point to the curator research repository. A reader of this repository cannot resolve the labels."
    location: "decisions/0021-sessions-enter-through-curator-run.md:7-8,186-192"
    action: "Fold the decision text, or a stable pointer, into the contract-owner documents the next time they are published."
  - id: B-N3
    severity: note
    severity_reason: "P2 note. The shared-library paragraph calls the old launcher boundary 'single-module/no-shared-library'. The launcher SPEC §1 wording is 'No Curator or ax imports ... No shared libraries, no other import edges'. The amendment's intent is right; the label is loose."
    location: "decisions/0021-sessions-enter-through-curator-run.md:205-208"
    action: "Quote the SPEC §1 clause when the launcher amendment lands."
  - id: B-N4
    severity: note
    severity_reason: "P3 scope note. W3, I4 and C4 are not R-LR1 rows; they are carried as open questions in a decision about the executor. They are faithful to the table's options and do not enable anything, but their owners are the user-manager, dispatcher/session-host and broker documents."
    location: "decisions/0021-sessions-enter-through-curator-run.md:247-268"
    action: "Keep one owner per question when those documents are amended; avoid divergent copies."
  - id: B-N5
    severity: note
    severity_reason: "P3 style. The launcher link points to unpinned `main`; one reflowed line in the Status section is longer than the surrounding wrap."
    location: "decisions/0021-sessions-enter-through-curator-run.md:22,214"
    action: "Optional."
```

## Verdict
No P0/P1 remains. Accepting revision 1 via accept_cr; the producer role routes the integration.
