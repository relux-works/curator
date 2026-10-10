# THE ONLY CURRENT INSTRUCTION — rework TASK-261010-1ypla7 to rev2 (researcher)

The deciding review `TASK-261010-1ypla7_review-verdict-rev1.md` requests changes. Reviewer A's record is `TASK-261010-1ypla7_contract-table-review-A.md`. Everything else in rev1 stays as it is.

Required:
- **F1 (P1).** Add a canonical row for the user manager's retire fate (audit A03.9). Cover the allowed set, `archive` / `delete` / `keep`, who may select each, and where `keep` is decided. The helper request admits archive and delete, while the archive policy adds an operator-only keep (UM L121–150). Mark it decided or OWNER DECISION NEEDED from the sources only.
- **F2 (P1).** Add rows for audit A07.5, A07.8 and A07.10:
  - A07.5: same-version versus qualified-version-range resume;
  - A07.8: which SH1 gate lanes are public module CI and which are private board-repository lanes;
  - A07.10: public tagged-module consumption versus the older private-fork rule.

Recommended (P2):
- Define `LR` in the source index, or replace it by `LS`.
- Add the R-CB1 items (helper ledger locator, key classes) and the R-DP1 sweep entry, if they fit.

Size: up to 80 KiB is allowed. Compress the repeated prose in the "Source and propagation index" before cutting any content. Add a short "rev2 changes" section. Cite every new value to file and line at the pinned commits. No first names, no local paths, no `LOGBOOK.md` edits, no tests or builds on this host. Then run `task-board handoff TASK-261010-1ypla7 --role researcher` and END YOUR TURN.
