# Checkpoint instruction — TASK-260910-1a75qd (bound developer run)

Revision 3 is accepted but this leaf is NOT the Story's final leaf (TASK-260910-19w2aj remains), so `integrate` refuses (task_delta). No board writes before/during. Run exactly, in the foreground, from the control root /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree checkpoint TASK-260910-1a75qd 2>&1 | tee .temp/checkpoint-1a75qd.log

Then attach `.temp/checkpoint-1a75qd.log` as outcome resource `TASK-260910-1a75qd_checkpoint-results.md` and stop. If it refuses, attach the exact refusal and stop.
