# Review brief SNAPPUB (deciding review B): SNAP-PUBLISH — Retained-handle destination publication (TASK-261016-w0yx2w)

Read gw-SNAPPUB-brief.md, the attached leaf file(s) and rc7-code-common.md. Review the candidate against the leaf:
1. Every acceptance criterion is met by code and tests, traced to the cited normative clauses (read them on curator-spec main).
2. Every paired test row exists, asserts the intended refusal and the unchanged forbidden state, and has its positive control; an implementation that refuses everything must fail the positive controls.
3. The plant/checker control was executed and failed on the planted violation.
4. No change outside the leaf's paths and seams without a justified note; no claim of qualification that was not executed.
5. Re-run the focused tests and the repository gate yourself.
Verdict ACCEPT or REJECT with findings P1/P2/P3. Follow the board's review contract.

## After the owner decision (2026-10-17)

Review against curator-spec main 9dfc557 §2 and CIP-0014: absent destination or intact prior only; an existing empty directory refuses. Attack the publish step yourself: create a competing destination between validation and publication, substitute a staging entry, interrupt staging; publication must refuse or reach exactly the intended object, never overwrite or delete anything but this invocation's staging. Check the no-replace syscalls per platform and that Windows and other unqualified platforms refuse.
