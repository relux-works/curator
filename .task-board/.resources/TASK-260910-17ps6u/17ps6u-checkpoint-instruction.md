# Checkpoint instruction — TASK-260910-17ps6u (bound developer run)

Revision 4 is accepted but this leaf is NOT the Story's final leaf (TASK-260910-dufdai remains in STORY-260910-20sx61), so `integrate` refuses (task_delta). No board writes before/during. Run exactly, in the foreground, from the control root /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree checkpoint TASK-260910-17ps6u 2>&1 | tee .temp/checkpoint-17ps6u.log

Then attach `.temp/checkpoint-17ps6u.log` as outcome resource `TASK-260910-17ps6u_checkpoint-results.md` and stop. If it refuses, attach the exact refusal and stop.
