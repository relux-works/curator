# TASK-261013-g96cen — review verdict, revision 2 (reviewer B): CONTENT ACCEPTED; acceptance NOT recordable (board gate)

Reviewed base a015428 → candidate tree 0ac03793 (worktree equals candidate; `git diff` against it is empty). Docs only; no tests or builds run. Reviewer A's record (`gw-CS7r2-review-A.md`) was not attached when this verdict was written, so there was nothing to reconcile and no A-raised P0/P1 to check.

## Revision 1 items

| Item | Result |
|---|---|
| B1 (P1) | Fixed. Decisions 0010 and 0013 are absent from the changed-path list, so they are byte-identical to the base, and their bare "Decision 5" citations are unchanged. Decision 0019 now says "Decision 0013 Decision 5" at its two body sites, with a correct revision note. The Decision 0005 retitle keeps the original heading anchor. The gaps file no longer carries the false statement. |
| B2 | Fixed. CIP-0010 Phasing item 3 and CIP-0011 open question 2 link the launcher intake lane (Decision 0021 amendment; launcher SPEC §10). The disposition does the same at its three sites. The only `R-LR1`/`R-CS2` strings left are the historical audit-row link in Decision 0021 (a link label, not a normative use) and the disposition's "historical audit row R-CS2 (not a rule)". |
| B3 | Fixed. Decision 0021 Amendment v1 has a scoped cross-reference after the blanket sentence. The anchor `#wire-naming-rule` resolves. |
| B4 | Fixed. CIP-0003 and CIP-0010 (summary line and question 4) state: priority decided on 2026-10-13; channel (b) with fallback (c) proposed; qualification open. The API-key lane stays separate, and the other acceptance questions stay open. |
| B5 | Fixed. The disposition's D-WIRE-NAMES paragraph quotes the dispatcher §3.3 scope sentence and says the index lives in wiki `naming.md`, "Shared index: rules and wire names". The text says this is not a new owner decision. I confirmed in the local dispatcher repository that the quoted sentence is exact at 781a9dc, and that the section exists in the local wiki checkout. |
| A1 | Fixed. `donor-layout.puml:84` and `donor-lifecycle.puml:88` carry "except an admitted repair round". |
| A2 | Fixed. CIP-0009 D5a "Export outcomes and repair" and the transfer figure limit the contradiction comparison to the same unchanged quiesced tree, not across an admitted repair. |

## Checks 1–8

1. All 8 findings are applied in this module, with the sections listed in the gaps file (CIP-0009 D5a/D5c/test plan/open questions, the disposition, `protocol/remote-worker.md`, CIP-0003/0010/0011, the decisions, and the 10 figures plus the README and DSL). I found no finding that is claimed but missing.
2. Bounds hold. Standing discard is stated as unchosen and off. The `workspace-result/1` field set is unchanged. No qualification or implementation is claimed.
3. Shared names are unchanged. I found no invented wire fields, refusal codes or numeric values. Permission-failure classification, repair intake and the budget are stated as unpublished.
4. Security properties survive:
   - Unknown or interrupted exports stay pending and never discharge the obligation.
   - Terminal effects are not rerun; re-observation is a new linked effect.
   - Cleanup consent covers ordinary expiry only. Holds, ALLSTOP, revocation and restore need separate authority and are never bypassed.
   - Repair is forward work under current authority.
   - Waivers are per-act and widening.
5. The figures match the text: the transfer, retire, turn, layout, lifecycle and operator-custody sources agree with D5a and are cross-linked.
6. The gaps file is complete and its statements check out against the diff.
7. Hygiene: no F-number tokens were added (per-file counts of `F-<n>` outside words are equal in base and candidate; the existing ones were preserved, not stripped). I found no host names, local paths, event ids or seat names in the added lines.
8. No collateral edits; `UTF-8` and other passages outside the findings are untouched.

## Notes (P2, not blocking)

- Pin mix: the disposition cites the dispatcher at 74687e6 elsewhere but quotes the §3.3 sentence at 781a9dc. 74687e6 does not contain that sentence. Consider re-pinning to one revision when the dispatcher heads are next refreshed.
- Pre-existing `F-n` labels in CIP-0009 (D5 bullets, test plan) were retained as base text. They are outside this wave's bound; a future hygiene pass could remove them.
- The automatic handoff gate conflicts with the docs-only no-tests brief, as the gaps file records. Source-only validators were run by the developer; I did not rerun them.

## Recording status (important)

`accept_cr(TASK-261013-g96cen, revision=2, ...)` was refused with `change_request_state_conflict`: `CR-TASK-261013-g96cen-2` was created directly in state `changes_requested` (ledger sequence 3, run RUN-261013-27b50e) and has never been `ready`. The developer's gaps file explains why: the automatic handoff gate (`spec-gate.sh`) runs Python unittest discovery, `go vet` and Go test compilation, which the docs-only brief forbids, so the revision-2 handoff never completed cleanly. I did not run that gate myself or work around it. I also removed the obsolete "Tests green" checklist item (tests/builds are excluded by the brief), as the developer had already done once; I did not check it.

On content, I would accept revision 2: no P0/P1 remains, and the notes above are P2.

Needed to record acceptance: either a source-only documentation handoff gate for this task, or explicit owner authorization to run the existing test/compile gate. After that the developer re-runs `task-board handoff` so that a `ready` revision exists, and a reviewer records `accept_cr` against it. The candidate tree 0ac03793 is unchanged and still in the Story worktree.
