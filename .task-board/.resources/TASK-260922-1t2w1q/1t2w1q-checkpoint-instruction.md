# Checkpoint instruction — TASK-260922-1t2w1q (bound developer run)

Revision 4 is accepted but this leaf is NOT the Story's final leaf (TASK-260922-cww1ov remains in STORY-260922-1cenbr), so `integrate` refuses (task_delta). No board writes before/during. Run exactly, in the foreground, from the control root /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree checkpoint TASK-260922-1t2w1q 2>&1 | tee .temp/checkpoint-1t2w1q.log

Then attach `.temp/checkpoint-1t2w1q.log` as outcome resource `TASK-260922-1t2w1q_checkpoint-results.md` and stop. If it refuses, attach the exact refusal and stop.
