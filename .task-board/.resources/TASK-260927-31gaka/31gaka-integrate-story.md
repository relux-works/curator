# TASK-260927-31gaka — YOU must run the Story integrate (THE ONLY CURRENT INSTRUCTION; bound developer run)

The previous bound run only confirmed preconditions and did not run the command; the runner then attempted a checkpoint again and refused
with change_request_final_leaf_checkpoint: "Run `task-board worktree integrate STORY-260927-3qf8er` instead, which lands the branch".
The board says an untracked caller cannot do it ("requires a tracked developer (implementer) run") — YOU are that tracked run.
1. In a shell, run EXACTLY (control root /Users/administrator/Developer/ReluxWorks/curator/curator, with
   `export SSH_AUTH_SOCK=/private/tmp/com.apple.launchd.PXt8w1CCF2/Listeners` and your run's environment unchanged):
   `task-board worktree integrate STORY-260927-3qf8er --cr TASK-260927-31gaka --revision 2`
2. Paste its full output and exit code into `TASK-260927-31gaka_story-integrate-2.md` and attach it (`task-board resource add … --type
   outcome`). If it refuses, paste the refusal and stop. Do not edit files, push, or set statuses by hand. END YOUR TURN.
