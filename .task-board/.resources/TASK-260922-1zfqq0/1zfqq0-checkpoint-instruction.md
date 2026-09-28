# Checkpoint instruction — TASK-260922-1zfqq0 (bound developer run, launcher)

Revision 2 (F-L1a, launcher SPEC 0.5.0-draft) is ACCEPTED but this leaf is NOT the Story's final leaf
(TASK-260922-2u5jzw F-L1b remains in STORY-260922-39hxog), so it is checkpointed onto the Story branch.
No board writes before/during. Run exactly, in the foreground, from the control root
/Users/administrator/Developer/ReluxWorks/curator/curator-agent-launcher:

    task-board worktree checkpoint TASK-260922-1zfqq0 2>&1 | tee .temp/checkpoint-1zfqq0.log

Then attach `.temp/checkpoint-1zfqq0.log` as outcome resource `TASK-260922-1zfqq0_checkpoint-results.md`
and stop. If it refuses, attach the exact refusal and stop. Change no file.
