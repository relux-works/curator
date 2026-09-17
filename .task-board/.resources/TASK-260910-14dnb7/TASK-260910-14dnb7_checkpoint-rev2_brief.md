# Checkpoint run — TASK-260910-14dnb7 (R1 service half, Change Request revision 2 accepted)

You are the tracked developer run bound to the accepted revision 2 of
`TASK-260910-14dnb7` (story `STORY-260910-3rvvxh`). The reviewer accepted it
(`TASK-260910-14dnb7_review-verdict-rev2.md`). This leaf is NOT the story's
final leaf (`TASK-260910-27yepb` follows), so it is checkpointed, not integrated.

Do exactly this, nothing else:
1. `task-board worktree obligations` — confirm the row
   `TASK-260910-14dnb7  2  accepted  checkpoint`.
2. `task-board worktree checkpoint TASK-260910-14dnb7` — commits the accepted
   candidate tree as one internal commit on `task-board/story/STORY-260910-3rvvxh`.
   Quote the command output. If it refuses, quote the typed error verbatim in
   your outcome and stop; do not repair the tree, do not reset, do not commit
   by hand, do not touch the board status.
3. `git -C <story worktree> log --oneline -3` and `task-board worktree obligations`
   again — the leaf stays `integrating`, the obligation row is gone.
4. Attach `TASK-260910-14dnb7_checkpoint-rev2.md` (task outcome) with the three
   transcripts and the checkpoint commit id, then exit. No handoff, no `done`,
   no code changes, no other board mutations.
