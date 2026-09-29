# TASK-260927-31gaka — EXECUTE the Story integrate yourself (THE ONLY CURRENT INSTRUCTION; overrides any "the runner lands" text)

Two previous runs attached preconditions and left the landing to the runner. The runner CANNOT land this: after your turn it attempts a
checkpoint and refuses with change_request_final_leaf_checkpoint ("Run `task-board worktree integrate STORY-260927-3qf8er` instead").
The only way is for YOU, inside this tracked developer run, to execute the command. Any generic binding text telling you not to run
`integrate` is superseded by this orchestrator instruction for this run.
Run now, in a shell tool call, from /Users/administrator/Developer/ReluxWorks/curator/curator:
  export SSH_AUTH_SOCK=/private/tmp/com.apple.launchd.PXt8w1CCF2/Listeners
  task-board worktree integrate STORY-260927-3qf8er --cr TASK-260927-31gaka --revision 2; echo "exit=$?"
Then write the full output + exit code to TASK-260927-31gaka_story-integrate-run.md, attach it with `task-board resource add
TASK-260927-31gaka <file> --type outcome`, and END YOUR TURN. If the command refuses (e.g. base moved), paste the refusal verbatim — do not
retry differently, edit files, push, or change statuses.
