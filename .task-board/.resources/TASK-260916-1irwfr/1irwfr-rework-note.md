# Rework note — TASK-260916-1irwfr (revision 2: story_final)

Revision 1 was published while sibling leaves were still open, so the board derived kind `task_delta`. Both siblings (TASK-260908-1bpra2, TASK-260916-bn5kvb) are now `done` (closed as landed: relux-mcp 027f55b, relux-root-context abaadf43). You are the Story's last open leaf, so the next publication derives `story_final`.

Do exactly what 1irwfr-brief.md says, nothing more: in the managed Story worktree verify `git status --porcelain` empty, HEAD tree == main tree (9eaad1ee), `bash scripts/validate.sh` exit 0; refresh the evidence resource (update TASK-260916-1irwfr_evidence.md, add a "revision 2" section with the outputs and exit codes); tick the checklist; `task-board handoff TASK-260916-1irwfr --role developer`. If the handoff refuses (e.g. a refresh/base-authority error), attach the exact output and stop — no workaround, no file edits, no manual commits.
