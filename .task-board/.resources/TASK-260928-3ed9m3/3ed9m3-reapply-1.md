# TASK-260928-3ed9m3 — re-apply accepted rev1 on trunk d8e87bac (THE ONLY CURRENT INSTRUCTION)

Revision 1 was ACCEPTED (global add/install lock publication ordering). Trunk moved (E5 managed-write nofollow, E6 path-source boundary)
and the merge conflicts on internal/envprofile/managed.go and switch.go. Accepted content: `refs/campaign/lpnvkn-rev1-20260928` (parent
97e85642, tree 83a0c43a). Your Story worktree is fresh on trunk.
1. `task-board m 'set_status(TASK-260928-3ed9m3, status=development)'`.
2. `git diff 97e85642 refs/campaign/lpnvkn-rev1-20260928 -- . ':!.task-board' > $TMPDIR/3ed9m3.patch; git apply --3way $TMPDIR/3ed9m3.patch`;
   resolve KEEPING BOTH SIDES: your lock-first ordering on top of E5's nofollow write helpers (every managed write stays through
   atomicManagedFile/atomicManagedLink/managedPath) and E6's path-source preflight. Add nothing else.
3. VERIFY `git diff --name-only origin/main -- . ':!.task-board'` = the 7 rev1 paths.
4. Focused runs (real exit codes): `go test ./internal/envprofile -run 'Global|Takeover|Lock|Nofollow|Path|Guarded'`,
   `go test ./cmd/curator -run 'Global|Takeover|Profile'`.
5. Append "Revision 2 — re-apply on d8e87bac", `resource update`, handoff, END YOUR TURN. No CHANGELOG/LOGBOOK edit.
