# Checkpoint instruction — TASK-260907-187z6x (bound developer run)

THE ONLY CURRENT INSTRUCTION. Revision 4 (identity-reviewed republish) is accepted; TASK-260907-2as5sx remains open in STORY-260906-1a2i5a, so this leaf is checkpointed onto the Story branch (not integrated). No board writes before/during. Run exactly, in the foreground, from the control root /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree checkpoint TASK-260907-187z6x 2>&1 | tee .temp/checkpoint-187z6x.log

Then attach `.temp/checkpoint-187z6x.log` as outcome resource `TASK-260907-187z6x_checkpoint-results.md` and stop. If it refuses, attach the exact refusal and stop.
