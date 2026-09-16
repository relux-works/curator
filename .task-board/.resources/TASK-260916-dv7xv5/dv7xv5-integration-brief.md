# Integration brief — TASK-260916-dv7xv5 (accepted CR revision 2, empty repository delta)

Supersedes every earlier brief on this task for this run. Review round 4
accepted CR `CR-TASK-260916-dv7xv5-2` (revision 2) with `accept_cr`; the task
is `integrating`, it is the only (final) leaf of `STORY-260916-1nc5dc`, and the
repository delta is empty (`task-board worktree integrating` classifies it
`empty_delta_board_only`). You are the tracked integration run bound to that
accepted revision. Do not research, edit, or attach anything.

Your whole job, from the control root the runtime placed you in:

1. `task-board worktree integrating` — confirm `TASK-260916-dv7xv5` reads
   `empty_delta_board_only`, rev 2.
2. `task-board worktree integrate STORY-260916-1nc5dc --cr TASK-260916-dv7xv5 --revision 2`
   — this performs the Story's `done` transition and commits the board delta
   as a board-only commit on the local trunk. Do NOT run `task-board handoff`
   afterwards (the transaction owns `done`). Do not push; the orchestrator
   pushes.
3. If `integrate` refuses, record the exact refusal text with
   `task-board m 'set_notes(TASK-260916-dv7xv5, text="…")'` and stop; do not
   retry with other flags, do not run `checkpoint` (final leaf), do not touch
   git by hand.
