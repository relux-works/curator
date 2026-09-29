# TASK-260928-2iu83q — carrier re-apply of TASK-260927-31gaka rev2 on current trunk (THE ONLY CURRENT INSTRUCTION)

Rev2 was ACCEPTED and checkpointed, but the Story could not land (base moved). The orchestrator split the revision-B flip out of this Story,
saved your accepted content as `refs/campaign/3qf8er-rev2-20260928` (parent 86552087, tree 31e7a36f) and discarded the Story branch. Your
Story worktree is fresh on trunk d8e87bac (E1, E5 nofollow, E6, global adopt, rose-air landed). You are the only leaf: your next revision is
the Story's final candidate.
1. `task-board m 'set_status(TASK-260928-2iu83q, status=development)'` (if the board refuses because a checkpointed revision exists, paste
   the refusal and stop).
2. `git diff 86552087 refs/campaign/3qf8er-rev2-20260928 -- . ':!.task-board' > $TMPDIR/31gaka.patch; git apply --3way $TMPDIR/31gaka.patch`;
   resolve KEEPING BOTH SIDES — your revision-A seed behaviour on top of E5's nofollow write helpers (every managed write through
   atomicManagedFile/atomicManagedLink/managedPath) and E1/E6 changes in status.go/managed.go; gap rows: keep trunk's, yours owned by
   TASK-260927-1e5qqm. Add nothing else.
3. VERIFY `git diff --name-only origin/main -- . ':!.task-board'` = the 7 rev2 paths.
4. Focused runs (real exit codes): `go test ./internal/envprofile -run 'Seed|Codex|Mcp|Status|Nofollow|Guarded'`, `go test ./cmd/curator -run
   'EnvResolve|Marker|EnvStatus|Seed|Mcp'`.
5. Append "Revision 1 — carrier re-apply of 31gaka rev2", `resource update`, handoff, END YOUR TURN. No CHANGELOG/LOGBOOK edit.
