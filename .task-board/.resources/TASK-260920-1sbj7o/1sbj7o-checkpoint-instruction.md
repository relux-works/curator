# Checkpoint instruction — TASK-260920-1sbj7o (bound developer run)

Revision 3 is accepted but this leaf is NOT the Story's final leaf (TASK-260920-3ccq6b, BUG-260920-2d9gfv, TASK-260919-2cmg0y remain in STORY-260919-37szes), so `integrate` refuses (task_delta). No board writes before/during. Run exactly, in the foreground, from the control root /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree checkpoint TASK-260920-1sbj7o 2>&1 | tee .temp/checkpoint-1sbj7o.log

Then attach `.temp/checkpoint-1sbj7o.log` as outcome resource `TASK-260920-1sbj7o_checkpoint-results.md` and stop. If it refuses, attach the exact refusal and stop.
