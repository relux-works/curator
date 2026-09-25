# Checkpoint instruction — TASK-260922-3bbvrs (bound developer run)

THE ONLY CURRENT INSTRUCTION. Revision 4 is accepted; TASK-260922-18ex37 remains open in STORY-260922-2goxjs, so this leaf is checkpointed onto the Story branch (not integrated). No board writes before/during. Run exactly, in the foreground, from the control root /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree checkpoint TASK-260922-3bbvrs 2>&1 | tee .temp/checkpoint-3bbvrs.log

Then attach `.temp/checkpoint-3bbvrs.log` as outcome resource `TASK-260922-3bbvrs_checkpoint-results.md` and stop. If it refuses, attach the exact refusal and stop.
