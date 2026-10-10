# THE ONLY CURRENT INSTRUCTION — rework TASK-261010-1ypla7 to rev3 (researcher, small)

The deciding delta review `TASK-261010-1ypla7_review-verdict-rev2.md` requests the following.

- **P1, row P8 (resume version rule).** The cited architecture section is marked proposed (AR §5.3 `[^предложено]`), not canon. Set its status to OWNER DECISION NEEDED and state the options: same version only, or a qualified version range. Do not present either one as canonical.
- **P2, rendering.** Remove the blank lines that break the tables before I5 and P7, so that every row renders as a table row.
- **P2, W4, P9, P11.** Apply the reviewer's notes. For P9: the owner's acceptance that pieces of board code may appear in public CI logs is a real owner decision, made in session on 2026-10-10. Cite it as such and keep the private-lane layout.

Add a "rev3 changes" line. No other edits, no first names, no `LOGBOOK.md` edits, no tests or builds on this host. Then run `task-board handoff TASK-261010-1ypla7 --role researcher` and END YOUR TURN.
