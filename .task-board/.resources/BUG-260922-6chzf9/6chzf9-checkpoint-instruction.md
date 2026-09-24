# Checkpoint instruction — BUG-260922-6chzf9 (bound developer run)

The accepted revision is a task_delta; it is now the only leaf of its Story, so it is checkpointed first and the Story then lands through the checkpointed-integrate path. No board writes before/during. Run exactly, in the foreground, from the control root /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree checkpoint BUG-260922-6chzf9 2>&1 | tee .temp/checkpoint-6chzf9.log

Then attach `.temp/checkpoint-6chzf9.log` as outcome resource `BUG-260922-6chzf9_checkpoint-results.md` and stop. If it refuses, attach the exact refusal and stop.
