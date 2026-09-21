# Checkpoint instruction — BUG-260920-2wyzde (bound developer run)

Revision 1 is accepted but this leaf is NOT the Story's final leaf (BUG-260920-3v7x6j, BUG-260920-2eg8nv, TASK-260920-1sbj7o, TASK-260920-3ccq6b, BUG-260920-2d9gfv, TASK-260919-2cmg0y remain in STORY-260919-37szes), so `integrate` refuses (task_delta). No board writes before/during. Run exactly, in the foreground, from the control root /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree checkpoint BUG-260920-2wyzde 2>&1 | tee .temp/checkpoint-2wyzde.log

Then attach `.temp/checkpoint-2wyzde.log` as outcome resource `BUG-260920-2wyzde_checkpoint-results.md` and stop. If it refuses, attach the exact refusal and stop.
