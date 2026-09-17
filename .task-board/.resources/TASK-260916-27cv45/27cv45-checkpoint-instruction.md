# Checkpoint instruction — TASK-260916-27cv45 (bound developer run)

Revision 1 is accepted but this leaf is NOT the Story's final leaf (TASK-260916-hxr6qv remains), so `integrate` refuses (task_delta). No board writes before/during. Run exactly, in the foreground, from the control root /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree checkpoint TASK-260916-27cv45 2>&1 | tee .temp/checkpoint-27cv45.log

Then attach `.temp/checkpoint-27cv45.log` as outcome resource `TASK-260916-27cv45_checkpoint-results.md` and stop. If it refuses, attach the exact refusal and stop.
