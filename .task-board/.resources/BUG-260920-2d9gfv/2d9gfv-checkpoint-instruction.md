# Checkpoint instruction — BUG-260920-2d9gfv (bound developer run)

Revision 2 is accepted but this leaf is NOT the Story's final leaf (TASK-260919-2cmg0y remains in STORY-260919-37szes), so `integrate` refuses (task_delta). No board writes before/during. Run exactly, in the foreground, from the control root /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree checkpoint BUG-260920-2d9gfv 2>&1 | tee .temp/checkpoint-2d9gfv.log

Then attach `.temp/checkpoint-2d9gfv.log` as outcome resource `BUG-260920-2d9gfv_checkpoint-results.md` and stop. If it refuses, attach the exact refusal and stop.
