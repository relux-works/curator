# Checkpoint run — TASK-260917-2ecpjv (rc.12 qualification, research, Change Request revision 1 accepted)

You are the tracked researcher run bound to the accepted revision 1 of
`TASK-260917-2ecpjv` (story `STORY-260917-3w3lvj`). The reviewer accepted it
(`TASK-260917-2ecpjv_review-verdict-rev1.md`). This leaf is NOT the story's
final leaf (`TASK-260918-11f9l1` follows), so it is checkpointed, not integrated.

Do exactly this, nothing else:
1. `task-board worktree obligations` — confirm the row
   `TASK-260917-2ecpjv  1  accepted  checkpoint`.
2. `task-board worktree checkpoint TASK-260917-2ecpjv` — records the accepted empty-delta revision (an empty delta creates no
   commit on `task-board/story/STORY-260917-3w3lvj`; the leaf stays integrating).
   Quote the command output. If it refuses, quote the typed error verbatim in
   your outcome and stop; do not repair the tree, do not reset, do not commit
   by hand, do not touch the board status.
3. `git -C <story worktree> log --oneline -3` and `task-board worktree obligations`
   again — the leaf stays `integrating`, the obligation row is gone.
4. Attach `TASK-260917-2ecpjv_checkpoint-rev1.md` (task outcome) with the three
   transcripts and the checkpoint commit id, then exit. No handoff, no `done`,
   no code changes, no other board mutations.
