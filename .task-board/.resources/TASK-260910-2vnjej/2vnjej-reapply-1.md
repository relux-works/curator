# TASK-260910-2vnjej — re-apply accepted rev4 on trunk (THE ONLY CURRENT INSTRUCTION)

Revision 4 was ACCEPTED. Trunk moved and three-way merge now conflicts in `.github/ci/conformance-gaps.tsv`, `internal/install/install.go`
and `internal/registry/snapshot.go`. The accepted content is `refs/campaign/6bo7ej-rev4-20260929`: parent 213a53e5, tree 2742e187, 26 paths.
Your Story worktree is fresh on current trunk.
1. `task-board m 'set_status(TASK-260910-2vnjej, status=development)'`.
2. Apply the accepted delta:
   `git diff 213a53e5 refs/campaign/6bo7ej-rev4-20260929 -- . ':!.task-board' > $TMPDIR/2vnjej.patch; git apply --3way $TMPDIR/2vnjej.patch`.
   Resolve every conflict KEEPING BOTH SIDES: trunk's changes (security landings since 213a53e5: E5 nofollow, E6 path boundary, posture,
   Codex seed, lock ordering, and so on) plus your cross-registry root check.
   - conformance-gaps.tsv: keep trunk's rows as they are, except the rows your rev4 removed, which stay removed. Do NOT re-add rows that
     trunk removed.
   Add nothing else.
3. VERIFY, and paste the output into the results:
   - `git diff --name-only origin/main -- . ':!.task-board'` = the 26 rev4 paths.
   - For each path: `git diff 213a53e5 refs/campaign/6bo7ej-rev4-20260929 -- P` versus `git diff origin/main -- P`. The added and removed
     lines must be the same, apart from the conflict resolutions. List each resolution.
   - `diff <(git show origin/main:.github/ci/conformance-gaps.tsv) .github/ci/conformance-gaps.tsv` shows only rev4's removals.
4. Focused runs, split per package, with real exit codes:
   - `go test ./internal/registry`
   - `go test ./internal/install -run 'Registry|Snapshot|Root|Mirror|Checkpoint|Bootstrap'`
   - `go test ./cmd/curator -run 'Registry|Env'`
   - `GOOS=windows go vet ./internal/registry ./internal/install`
5. Append "Revision 5 — re-apply on <trunk sha>" to the results, run `resource update`, then
   `task-board handoff TASK-260910-2vnjej --role developer`, then END YOUR TURN. The runner publishes the CR and runs the gate; do not wait
   for it. No CHANGELOG/LOGBOOK edit.
