# Checkpoint run — TASK-260910-28kmef (R5 service, Change Request revision 2 accepted)

You are the tracked developer run bound to the accepted revision 2 of
`TASK-260910-28kmef` (story `STORY-260910-2xe3n2`). The reviewer accepted it
(`TASK-260910-28kmef_review-verdict-rev2.md`). This leaf is NOT the story's
final leaf (`TASK-260910-2rsajv`, `TASK-260910-3u9t1e`, `TASK-260910-2c7s0u` follow), so it is checkpointed, not integrated.

Do exactly this, nothing else:
1. `task-board worktree obligations` — confirm the row
   `TASK-260910-28kmef  2  accepted  checkpoint`.
2. `task-board worktree checkpoint TASK-260910-28kmef` — commits the accepted
   candidate tree as one internal commit on `task-board/story/STORY-260910-2xe3n2`.
   Quote the command output. If it refuses, quote the typed error verbatim in
   your outcome and stop; do not repair the tree, do not reset, do not commit
   by hand, do not touch the board status.
3. `git -C <story worktree> log --oneline -3` and `task-board worktree obligations`
   again — the leaf stays `integrating`, the obligation row is gone.
4. Attach `TASK-260910-28kmef_checkpoint-rev2.md` (task outcome) with the three
   transcripts and the checkpoint commit id, then exit. No handoff, no `done`,
   no code changes, no other board mutations.
