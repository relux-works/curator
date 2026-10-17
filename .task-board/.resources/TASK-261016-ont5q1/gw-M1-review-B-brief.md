# Review brief M1 (deciding review B): M1 — Emit Muse v4 from verified Curator state (TASK-261016-ont5q1)

Read gw-M1-brief.md, the attached leaf file(s) and rc7-code-common.md. Review the candidate against the leaf:
1. Every acceptance criterion is met by code and tests, traced to the cited normative clauses (read them on curator-spec main).
2. Every paired test row exists, asserts the intended refusal and the unchanged forbidden state, and has its positive control; an implementation that refuses everything must fail the positive controls.
3. The plant/checker control was executed and failed on the planted violation.
4. No change outside the leaf's paths and seams without a justified note; no claim of qualification that was not executed.
5. Re-run the focused tests and the repository gate yourself.
Verdict ACCEPT or REJECT with findings P1/P2/P3. Follow the board's review contract.

## Overlap with the landed L15-1 Muse slice

Much of this leaf landed with TASK-261016-wh6v9a (3113bad). Check the producer's gap analysis row by row against main: every claimed "already satisfied" row must point at real code and a test that would fail on the forbidden effect; every new row must be needed. An empty or test-only delta is acceptable only with that evidence.
