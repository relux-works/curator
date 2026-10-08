# TASK-260924-2am4qa — cleanup (THE ONLY CURRENT INSTRUCTION; bound developer run)

Revision 1 is accepted on content; the only required change (review verdict rev1) is: delete `TASK-260924-2am4qa_results.md` from
the repository root of the Story worktree (it is a board resource, not a repository file). Change nothing else.
1. `task-board m 'set_status(TASK-260924-2am4qa, status=development)'`.
2. `rm TASK-260924-2am4qa_results.md` in the worktree; prove every other path is byte-identical to revision 1 (git status + per-file hashes).
3. Append "Revision 2 (cleanup)" to the board results resource (`task-board resource get … -o $TMPDIR/…`, edit there, `task-board resource
   update TASK-260924-2am4qa $TMPDIR/TASK-260924-2am4qa_results.md --type outcome`) — never recreate it in the worktree.
4. `task-board handoff TASK-260924-2am4qa --role developer`; stay in the turn during the gate. A `run_wrote_outside_worktree … policy warn`
   block is a warning — verify status `to-review` and revision 2 published.
