# TASK-260910-32gki6 — re-apply on trunk, second pass (THE ONLY CURRENT INSTRUCTION)

Your rev6, the re-apply of accepted rev5 on 6a7deb11, was green on every lane except the Naming gate. That failure came from a trunk
board resource, now scrubbed; nothing in your change caused it. Since then trunk has also landed BUG-260928-uyak0e, the pathwalk
vanishing-entry fix. It changes `internal/pathboundary/pathboundary.go`, and your delta conflicts with it there.

Your rev6 content is saved as `refs/campaign/148pj1-rev6-20260929` (parent 6a7deb11, tree 858cd916). Your Story worktree is fresh on the
current trunk.
1. `task-board m 'set_status(TASK-260910-32gki6, status=development)'`.
2. Apply the saved delta:
   `git diff 6a7deb11 refs/campaign/148pj1-rev6-20260929 -- . ':!.task-board' > $TMPDIR/32gki6.patch; git apply --3way $TMPDIR/32gki6.patch`.
   Resolve pathboundary.go KEEPING BOTH SIDES:
   - uyak0e's rule: an entry that vanishes between readdir and lstat, i.e. ENOENT on a child that is NOT the root or a path named by the
     lock, is skipped, not refused. Read its commit and tests first: `git log -1 -p origin/main -- internal/pathboundary`;
   - your S5 boundary checks.

   A named store entry, lock or marker that is missing must still fail closed as S5 requires. Only transient children of a walked tree
   may vanish. Add nothing else.
3. VERIFY, and paste the output into the results:
   - `git diff --name-only origin/main -- . ':!.task-board'` = the 31 rev6 paths;
   - for every path except pathboundary.go, the +/- lines are identical to rev6;
   - for pathboundary.go, show the resolution hunk.
4. Focused runs with real exit codes:
   - `go test ./internal/pathboundary`
   - `go test ./internal/envprofile -run 'Store|Boundary|Pin|Resolve|Status|Legacy|PathInstall'`
   - `go test ./cmd/curator -run Env`
   - `GOOS=windows go vet ./internal/pathboundary ./internal/envprofile`
5. Append "Revision 7 — re-apply on <trunk sha> (pathboundary.go merged with uyak0e)" to the results, run `resource update`, then
   `task-board handoff TASK-260910-32gki6 --role developer`, then END YOUR TURN. No CHANGELOG/LOGBOOK edit. Do not spell any employer
   name anywhere.
