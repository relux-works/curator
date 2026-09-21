# Checkpoint instruction — TASK-260920-3ccq6b (bound developer run)

Revision 2 is accepted but this leaf is NOT the Story's final leaf (BUG-260920-2d9gfv, TASK-260919-2cmg0y remain in STORY-260919-37szes), so `integrate` refuses (task_delta). No board writes before/during. Run exactly, in the foreground, from the control root /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree checkpoint TASK-260920-3ccq6b 2>&1 | tee .temp/checkpoint-3ccq6b.log

Then attach `.temp/checkpoint-3ccq6b.log` as outcome resource `TASK-260920-3ccq6b_checkpoint-results.md` and stop. If it refuses, attach the exact refusal and stop.
