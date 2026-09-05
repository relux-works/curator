# Producer brief: publish Change Request revision 2 for environments.md 1.1 (bring the reviewed commit into the fresh workspace)

The rework is committed on `draft/environments-revision-1-1` at `db642b1` (one signed commit past
curator-spec main `ec695ba`, tree = the reviewed rework + cycle-1 fixes + sprint evidence). The
previous managed Story workspace was discarded after a client-identity change; your run has a fresh
workspace at trunk. Do exactly this:

1. In your managed story worktree: `git status --short` must be empty; `git log --oneline -1` shows
   the trunk tip (`ec695ba` or later main). Run
   `git cherry-pick -S db642b1` (the commit is in the same repository; verify with
   `git cat-file -t db642b1`). If main advanced past `ec695ba`, the cherry-pick still yields exactly one
   commit past the workspace base; if it conflicts, stop and report — do not resolve by hand.
2. Verify: `git log --oneline -2` shows your new commit then the base; `git log --show-signature -1`
   verifies; `git diff db642b1 HEAD --stat` is empty (same tree unless main moved); `git status --short`
   empty (remove `tools/__pycache__/` if present; do not run the validator).
3. Do NOT edit any file otherwise. `task-board handoff TASK-260905-jb6rvg --role developer` publishes
   CR revision 2. Attach `TASK-260905-jb6rvg_cr2-publish-report.md` with the handoff output (revision,
   candidate commit and tree) or the exact refusal. Never write into the control root.
