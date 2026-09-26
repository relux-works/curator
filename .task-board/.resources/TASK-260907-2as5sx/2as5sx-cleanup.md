# TASK-260907-2as5sx — cleanup (THE ONLY CURRENT INSTRUCTION; bound developer run)

Revision 4 content is complete (its gate failed only on the known Windows internal/snapshot flake, fixed separately by BUG-260923-11jgkt). Required change before republishing: delete `TASK-260907-2as5sx_results.md` from
the repository root of the Story worktree (it is a board resource, not a repository file). Change nothing else.
1. `task-board m 'set_status(TASK-260907-2as5sx, status=development)'`.
2. `rm TASK-260907-2as5sx_results.md` in the worktree; prove every other path is byte-identical to revision 4 (git status + per-file hashes).
3. Append "Revision 5 (cleanup + republish)" to the board results resource (`task-board resource get … -o $TMPDIR/…`, edit there, `task-board resource
   update TASK-260907-2as5sx $TMPDIR/TASK-260907-2as5sx_results.md --type outcome`) — never recreate it in the worktree.
4. `task-board handoff TASK-260907-2as5sx --role developer`; stay in the turn during the gate. A `run_wrote_outside_worktree … policy warn`
   block is a warning — verify status `to-review` and revision 5 published.
