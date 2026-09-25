# Checkpoint instruction — TASK-260924-1aa9wb (bound developer run)

THE ONLY CURRENT INSTRUCTION. Revision 1 is accepted; TASK-260924-m28s6b (lock replay) remains open in STORY-260924-3eywt2, so this leaf is checkpointed onto the Story branch (not integrated). No board writes before/during. Run exactly, in the foreground, from the control root /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree checkpoint TASK-260924-1aa9wb 2>&1 | tee .temp/checkpoint-1aa9wb.log

Then attach `.temp/checkpoint-1aa9wb.log` as outcome resource `TASK-260924-1aa9wb_checkpoint-results.md` and stop. If it refuses, attach the exact refusal and stop.
