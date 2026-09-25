# TASK-260922-18ex37 — cleanup (THE ONLY CURRENT INSTRUCTION; bound developer run)

Revision 1 was CHANGES_REQUESTED only for F1 (review verdict rev1); the only required change is: delete `TASK-260922-3bbvrs_results.md` (it arrived with the 3bbvrs checkpoint) from
the repository root of the Story worktree (it is a board resource, not a repository file). Change nothing else.
1. `task-board m 'set_status(TASK-260922-18ex37, status=development)'`.
2. `git rm -q TASK-260922-3bbvrs_results.md` (it is committed on the Story branch by the checkpoint, so remove it as a tracked deletion) in the worktree; prove every other path is byte-identical to revision 1 (git status + per-file hashes).
3. Append "Revision 2 (cleanup)" to the board results resource (`task-board resource get … -o $TMPDIR/…`, edit there, `task-board resource
   update TASK-260922-18ex37 $TMPDIR/TASK-260922-18ex37_results.md --type outcome`) — never recreate it in the worktree.
4. `task-board handoff TASK-260922-18ex37 --role developer`; stay in the turn during the gate. A `run_wrote_outside_worktree … policy warn`
   block is a warning — verify status `to-review` and revision 2 published.
