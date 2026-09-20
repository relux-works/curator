# Checkpoint instruction — TASK-260910-stbg4d (bound developer run)

Revision 2 is accepted but this leaf is NOT the Story's final leaf (TASK-260910-1xya7x remains in STORY-260910-1cnwwp), so `integrate` refuses (task_delta). No board writes before/during. Run exactly, in the foreground, from the control root /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree checkpoint TASK-260910-stbg4d 2>&1 | tee .temp/checkpoint-stbg4d.log

Then attach `.temp/checkpoint-stbg4d.log` as outcome resource `TASK-260910-stbg4d_checkpoint-results.md` and stop. If it refuses, attach the exact refusal and stop.
