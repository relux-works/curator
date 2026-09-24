# Checkpoint instruction — TASK-260916-1xib1x (bound developer run)

Revision 5 is accepted but this leaf is NOT the Story's final leaf (TASK-260916-2ok97n remains in STORY-260822-2h0v9j), so `integrate` refuses (task_delta). No board writes before/during. Run exactly, in the foreground, from the control root /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree checkpoint TASK-260916-1xib1x 2>&1 | tee .temp/checkpoint-1xib1x.log

Then attach `.temp/checkpoint-1xib1x.log` as outcome resource `TASK-260916-1xib1x_checkpoint-results.md` and stop. If it refuses, attach the exact refusal and stop.
