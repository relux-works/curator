# THE ONLY CURRENT INSTRUCTION — rework TASK-261010-2iqn63 to rev2 (researcher)

The deciding review, `TASK-261010-2iqn63_review-B-deciding-verdict-rev1.md`, requested changes. Fix exactly the rows it lists under "Rows to fix in rev2". Leave every other row and section as it is.

P1:
- **A09.9.** Restate the book clause: capacity waits, quota refuses. Compare it with v0 (refusal), not v1 (queue). Adjust the R-DP1 basis text to match.
- **A06.1, A08.4, A11.7, A11.8.** Correct the CIP-0009 locators at 1604d402: D9 is L160, D10 is L162, D11 is L164–169, and the carrier clause is L145. Then check every other CIP-0009 D-block citation (A04.4, A04.5, A11.1–A11.11) against the text it names, and fix any drift.

P2, fixed in the same revision:
- **A09.4.** Relabel it as an ambiguity (a default registry versus a shipped trust bundle), not a CONFLICTING row. Keep the clarification action.
- **A02.2, A07.6, A08.3.** Relabel these as scope or qualifier gaps.
- **D1.** Tag R-LR1 and the first-consumer item in R-CS2 as OWNER DECISION. Put R-CS2 before R-CB1, or state the dependency explicitly.
- **D1 R-SH1, R-CB1 and D2 B-F1.** List the concrete A-row IDs instead of ranges.
- **Hygiene.** Remove the first names quoted from a private document's heading in the AR:L1412 link titles (five occurrences) and use "§7.8" instead. This file is public.
- **A06.2, A08.1.** Mention the public bridge README stub as the only bridge source.

Re-run your citation verifier on the changed rows. Add a short "rev2 changes" section that lists each row ID with one line on what changed. No `LOGBOOK.md` edits, no tests or builds on this host. Then run `task-board handoff TASK-261010-2iqn63 --role researcher` and END YOUR TURN.
