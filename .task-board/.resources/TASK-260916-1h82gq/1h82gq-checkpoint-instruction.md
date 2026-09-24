# Checkpoint instruction — TASK-260916-1h82gq (bound developer run)

Revision 3 is accepted but this leaf is NOT the Story's final leaf (TASK-260916-1l44nd, TASK-260916-1xib1x, TASK-260916-2ok97n remain in STORY-260822-2h0v9j), so `integrate` refuses (task_delta). No board writes before/during. Run exactly, in the foreground, from the control root /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree checkpoint TASK-260916-1h82gq 2>&1 | tee .temp/checkpoint-1h82gq.log

Then attach `.temp/checkpoint-1h82gq.log` as outcome resource `TASK-260916-1h82gq_checkpoint-results.md` and stop. If it refuses, attach the exact refusal and stop.
