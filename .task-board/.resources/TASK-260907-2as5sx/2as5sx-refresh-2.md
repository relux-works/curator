# TASK-260907-2as5sx — refresh onto trunk a48f584c (THE ONLY CURRENT INSTRUCTION)

Revision 7 was ACCEPTED on content; it went stale only because trunk a48f584c (TASK-260924-1kpw4w, manifest dependency directory) changed
`internal/marker/marker.go`, which rev7 also changes. Safety ref of your worktree: refs/campaign/2as5sx-rev7-20260924.
1. `task-board m 'set_status(TASK-260907-2as5sx, status=development)'`.
2. Combine trunk's incoming content: `git diff 948ae7c9 a48f584c -- . ':!.task-board' ':!CHANGELOG.md' | git apply --3way` into the worktree;
   in `internal/marker/marker.go` keep BOTH sides (1kpw4w's directory fields/validation + your stateread seam); if 1kpw4w added any new
   manager-state reader, migrate it onto the seam or allowlist it with a reason so the deny-by-default guard stays green. Leave nothing staged.
3. `task-board worktree refresh-candidate TASK-260907-2as5sx` (replay conflicts via its template only).
4. Bounded: the guard test, marker package, your focused rows, 1kpw4w's marker/directory rows. No CHANGELOG edit; no stray files.
Append "Revision 8 — refresh onto a48f584c", `resource update`, `task-board handoff TASK-260907-2as5sx --role developer`. A
`run_wrote_outside_worktree … policy warn` block is a warning.
