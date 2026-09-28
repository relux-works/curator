# TASK-260916-19shmj — re-apply accepted rev5 on trunk 97e85642 (THE ONLY CURRENT INSTRUCTION)

Revision 5 was ACCEPTED (nofollow managed writes; Windows entry replacement). Trunk moved (E1, E3, E4, ryh3kw …) and the three-way merge
conflicts on .github/ci/root-artifacts.tsv, internal/envprofile/managed.go, internal/envprofile/state_read_guard_test.go. The accepted content
is `refs/campaign/73a5zg-rev5-20260927` (parent eca2bf27; tree = eca2bf27 + exactly 9 own paths). Your Story worktree is fresh on trunk.
1. `task-board m 'set_status(TASK-260916-19shmj, status=development)'`.
2. `git diff eca2bf27 refs/campaign/73a5zg-rev5-20260927 -- . ':!.task-board' > $TMPDIR/19shmj.patch; git apply --3way $TMPDIR/19shmj.patch`;
   resolve KEEPING BOTH SIDES. Every write trunk added in managed.go since eca2bf27 (E3 seed strip/record, ryh3kw stateread routing, E1
   signer/delta, 31gaka not yet) must ALSO follow your nofollow rule (atomicManagedFile/atomicManagedLink/managedPath) — route them, and name
   each such routing in the results. state_read_guard_test.go: keep both allow-list/route sets. Add nothing else.
3. VERIFY `git diff --name-only origin/main -- . ':!.task-board'` = the 9 own paths.
4. Focused runs (real exit codes): `go test ./internal/envprofile -run 'Nofollow|Credential|Link|Drift|Passthrough|Seed|Repair|Guarded'`
   (split), `GOOS=windows go vet ./internal/envprofile`.
5. Append "Revision 6 — re-apply on 97e85642" to the results, `resource update`, `task-board handoff TASK-260916-19shmj --role developer`,
   END YOUR TURN. No CHANGELOG/LOGBOOK edit.
