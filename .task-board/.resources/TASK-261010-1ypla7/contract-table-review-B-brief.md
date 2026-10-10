# THE ONLY CURRENT INSTRUCTION — reviewer B of two (R138, cross-provider), the deciding review of Change Request rev1 of TASK-261010-1ypla7 (read-only)

The study `.research/261010_platform-contract-table.md` is the single contract table that every P0 documentation fix will follow. The brief is the precondition `contract-table-brief.md`; owner decisions are in `owner-decisions-20261010-audit.md`.

Check:
1. **Citations.** Take at least 12 rows across names, signatures, identifiers, bounds, ownership, credentials, goals, identity and donor. Each canonical value matches its cited text at the pinned commit (`gh api`, read-only).
2. **Decided versus open.** No row invents a value or turns a recommendation into a decision. Every "decided" row cites a real decision. D-GOALS (variant b for now, task-board optional, no coupling) and D-IDENTITY are reflected exactly. Every OWNER DECISION NEEDED row states real options.
3. **Coverage.** Every P0 item of the audit's D1 (R-CS1, R-LR1, R-DP1, R-UM1, R-SH1, R-CB1, R-CS2) can be executed from this table: name any contradiction the table leaves without a canonical row.
4. **Hygiene.** Only the research file changes; no `LOGBOOK.md` change; no first names, secrets, local paths or session links.

No tests or builds on this host. Do checks 1–4 yourself and write your verdict BEFORE you read `contract-table-review-A.md`. Then reconcile and check every P0/P1 that A raised. Accept (`accept_cr`) if no P0/P1 remains; P2 issues are notes under tb-R226. Otherwise request changes, naming each row. END YOUR TURN.
