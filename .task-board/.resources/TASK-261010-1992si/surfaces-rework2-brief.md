# THE ONLY CURRENT INSTRUCTION — finish rev2 of TASK-261010-1992si: refresh the stale Story checkpoint, then republish (researcher, mechanical)

The two previous runs prepared rev2 correctly. The delta from Story checkpoint 1d7eb18c is only `.research/261010_project-surfaces-coverage.md`, byte-identical to rev1 (resources `TASK-261010-1992si_rev2-scope.md` and `TASK-261010-1992si_rev2-identity.md`). Change Request construction then refused with stale-anchor `change_request_base_authority_mismatch`: checkpoint 1d7eb18c does not descend from current trunk, because trunk moved by board-state commits.

Do exactly this:
1. Run `task-board worktree refresh-candidate TASK-261010-1992si`. It replays the Story checkpoint onto fresh trunk and carries your uncommitted delta. Expect no conflict, because trunk moved only under `.task-board/`. If it refuses or reports a conflict, attach the exact output as a resource and stop. Do not resolve anything by hand.
2. Confirm that the workspace delta from the new checkpoint is still only `.research/261010_project-surfaces-coverage.md`, byte-identical to rev1.
3. Run `task-board handoff TASK-261010-1992si --role researcher` and END YOUR TURN.

No content changes, no `LOGBOOK.md` edits, no tests or builds on this host, and no hand git operations on the branch (no commit, rebase, reset or checkout).
