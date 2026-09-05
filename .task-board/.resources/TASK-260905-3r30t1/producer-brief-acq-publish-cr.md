# Producer brief: publish a fresh Change Request for the byte-exact acquisition task (no edits)

The deliverable is in the curator repository (`/Users/iv/Developer/ReluxWorks/.worktrees/curator-acquisition-byte-exact`,
branch `feat/byte-exact-acquisition`, head `a46abc80`, PR #58 green) — not in this curator-spec story workspace, which
is expected to carry an empty delta. The previous managed workspace was discarded after a base refresh made its
Change Request stale; your run has a fresh one. Do exactly this: `git status --short` in your managed story worktree
must be empty (remove `tools/__pycache__/` if present); do NOT edit, commit, or touch the curator worktree; then
`task-board handoff TASK-260905-3r30t1 --role developer` to publish the Change Request revision with an empty delta.
Attach `TASK-260905-3r30t1_cr-publish-report.md` with the handoff output (revision, base) or the exact refusal.
Never write LOGBOOK.md or anything into the control root or any repository.
