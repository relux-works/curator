# THE ONLY CURRENT INSTRUCTION — delta reviewer A (R138), RECORD-ONLY, for Change Request rev2 of TASK-261010-2iqn63 (read-only)

Rev1 was sent back by `TASK-261010-2iqn63_review-B-deciding-verdict-rev1.md`. Its section "Rows to fix in rev2" is the checklist. The producer's brief was `docs-audit-rework-brief.md`.

Check:
1. Every listed row is fixed as required:
   - P1: A09.9; A06.1, A08.4, A11.7 and A11.8; the sweep of all other CIP-0009 D-block citations.
   - P2: A09.4; A02.2, A07.6 and A08.3; the D1 OWNER DECISION tags and the R-CS2/R-CB1 order; the row IDs in R-SH1, R-CB1 and B-F1; the removal of the first names; A06.2 and A08.1.
   For the P1 rows, verify the corrected locators against CIP-0009 at 1604d402 (`gh api`, read-only).
2. The rev1→rev2 diff of the study touches only those rows, the "rev2 changes" section, and text that directly depends on them. Name anything else that changed.
3. No first names, secrets, personal or local paths, or session links; no `LOGBOOK.md` change.

RECORD-ONLY: attach `docs-audit-delta-A.md` on TASK-261010-2iqn63, with a findings block where `severity` is one of `bypass|regression|robustness|note` and the P-level goes in `severity_reason`. Do NOT call accept_cr or reject_cr. No tests or builds on this host. END YOUR TURN.
