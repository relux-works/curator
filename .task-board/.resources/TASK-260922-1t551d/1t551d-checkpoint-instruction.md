# Checkpoint instruction — TASK-260922-1t551d (bound developer run)

Revision 4 is accepted but this leaf is NOT the Story's final leaf (TASK-260922-1t2w1q, TASK-260922-cww1ov remain in STORY-260922-1cenbr), so `integrate` refuses (task_delta). No board writes before/during. Run exactly, in the foreground, from the control root /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree checkpoint TASK-260922-1t551d 2>&1 | tee .temp/checkpoint-1t551d.log

Then attach `.temp/checkpoint-1t551d.log` as outcome resource `TASK-260922-1t551d_checkpoint-results.md` and stop. If it refuses, attach the exact refusal and stop.
