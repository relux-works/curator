# TASK-260910-2vnjej — re-apply accepted rev6 on trunk (THE ONLY CURRENT INSTRUCTION)

Revision 6 was ACCEPTED: a fidelity-reviewed re-apply of the accepted rev4. Since then trunk landed BUG-260928-uyak0e, the pathwalk
vanishing-entry fix, among others. It conflicts with you in `internal/pathboundary/pathboundary_test.go`, and pathboundary.go auto-merges.
Your accepted content is saved as `refs/campaign/6bo7ej-rev6-20260929`: parent cea992e2, tree 12b6c386, 26 paths. Your Story worktree is
fresh on trunk.
1. `task-board m 'set_status(TASK-260910-2vnjej, status=development)'`.
2. Apply the accepted delta:
   `git diff cea992e2 refs/campaign/6bo7ej-rev6-20260929 -- . ':!.task-board' > $TMPDIR/2vnjej.patch; git apply --3way $TMPDIR/2vnjej.patch`.
   Resolve pathboundary_test.go KEEPING BOTH SIDES: every uyak0e test plus every one of your tests. Read `git log -1 -p origin/main --
   internal/pathboundary` if you need to. Check the auto-merged pathboundary.go: uyak0e's vanishing-entry skip and your nofollow-read
   additions must both be present. Add nothing else.
3. VERIFY, and paste the output into the results:
   - `git diff --name-only origin/main -- . ':!.task-board'` = the 26 rev6 paths;
   - for every path except the two pathboundary files, the +/- lines are identical to rev6;
   - `go test ./internal/pathboundary` runs every uyak0e test name and every rev6 test name. List them.
4. Focused runs with real exit codes:
   - `go test ./internal/pathboundary ./internal/registry`
   - `go test ./internal/install -run 'Registry|Snapshot|Root|Mirror|Checkpoint|Bootstrap'`
   - `GOOS=windows go vet ./internal/pathboundary ./internal/registry`
5. Append "Revision 7 — re-apply on <trunk sha> (pathboundary_test.go merged with uyak0e)" to the results, run `resource update`, then
   `task-board handoff TASK-260910-2vnjej --role developer`, then END YOUR TURN. The runner publishes the CR and runs the gate; do not wait
   for it. No CHANGELOG/LOGBOOK edit. Never spell any employer name.
