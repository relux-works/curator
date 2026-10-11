# THE ONLY CURRENT INSTRUCTION — hand off rev3 of TASK-261011-2xdx9u (developer)

Rev2 was withdrawn because its base was superseded: it edited `decisions/0021…`, which belongs to TASK-261011-kojgyg and has landed. The orchestrator restored that one file to its base version in your worktree. Your CIP changes are still in place.

Do this:
1. If the workspace is not on current trunk, converge it with the standard flow. If you are a rework leaf, run `task-board worktree refresh-candidate TASK-261011-2xdx9u`.
2. Make sure `decisions/0021…` is NOT changed by your candidate.
3. Remove the untracked `LOGBOOK.md` from the repository. Producers never write LOGBOOK; put any notes into a board resource instead.
4. Keep `cips/platform-contract-disposition.md` and the CIP edits.
5. Hand off normally. The configured validation at handoff is sanctioned (compile-only Go); do not run tests or builds yourself.

If a board command refuses, attach the exact output and stop. END YOUR TURN.
