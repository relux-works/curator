# Checkpoint instruction — TASK-260910-5nrmtt (bound developer run)

Revision 6 is accepted but this leaf is NOT the Story's final leaf (TASK-260916-2bwfli remains), so `integrate` refuses (task_delta). No board writes before/during. Run exactly, in the foreground, from the control root /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree checkpoint TASK-260910-5nrmtt 2>&1 | tee .temp/checkpoint-5nrmtt.log

Then attach `.temp/checkpoint-5nrmtt.log` as outcome resource `TASK-260910-5nrmtt_checkpoint-results.md` and stop. If it refuses, attach the exact refusal and stop.
