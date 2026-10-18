# THE ONLY CURRENT INSTRUCTION — BUG-261018-uoisfe: retry publication (developer)

Your previous runs blocked correctly. The orchestrator removed checklist item 24 (post-turn validation succeeds): it cannot hold before handoff. Earlier: the post-turn Change Request validation failed at snapshot signing because the orchestrator launched it without the approved host signing environment. That is fixed: this run and its validator are launched with `agent-git-env` sourced.

## Do this

1. Change no source, test or documentation file. Confirm the tree still equals your last signed candidate; if it does not, stop and report why.
2. If the board requires it, run `task-board worktree refresh-candidate BUG-261018-uoisfe`.
3. Add a short "Publication retry" section to the gaps file: the candidate identity and that nothing changed.

Do not add checklist items about the post-turn validation: it runs after your handoff and is the gate itself.

## Hand-off

Run `task-board handoff BUG-261018-uoisfe --role developer`, then END YOUR TURN. The configured validation reruns automatically.
