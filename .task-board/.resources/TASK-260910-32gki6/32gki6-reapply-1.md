# TASK-260910-32gki6 — re-apply accepted rev5 on trunk (THE ONLY CURRENT INSTRUCTION)

Revision 5 was ACCEPTED. Trunk moved and three-way merge now conflicts in `.github/ci/conformance-case-counts.tsv` and `internal/envprofile/status.go`. The accepted content is `refs/campaign/148pj1-rev5-20260929`: parent 213a53e5, tree 0fd98b72, 31 paths.
Your Story worktree is fresh on current trunk.
1. `task-board m 'set_status(TASK-260910-32gki6, status=development)'`.
2. Apply the accepted delta:
   `git diff 213a53e5 refs/campaign/148pj1-rev5-20260929 -- . ':!.task-board' > $TMPDIR/32gki6.patch; git apply --3way $TMPDIR/32gki6.patch`.
   Resolve every conflict KEEPING BOTH SIDES: trunk's changes (security landings since 213a53e5: E5 nofollow, E6 path boundary, posture,
   Codex seed, lock ordering, and so on) plus your profile-store protected boundary (S5).
   - conformance-case-counts.tsv / conformance-gaps.tsv: keep trunk's rows; apply only rev5's own changes; recompute counts if the
     ledger-consistency script requires it (run `bash .github/ci/ledger-consistency.sh` if it exists). Do NOT re-add rows trunk removed.
   Add nothing else.
3. VERIFY, and paste the output into the results:
   - `git diff --name-only origin/main -- . ':!.task-board'` = the 31 rev5 paths.
   - For each path: `git diff 213a53e5 refs/campaign/148pj1-rev5-20260929 -- P` versus `git diff origin/main -- P`. The added and removed
     lines must be the same, apart from the conflict resolutions. List each resolution.
   - `diff <(git show origin/main:.github/ci/conformance-gaps.tsv) .github/ci/conformance-gaps.tsv` shows only rev5's own ledger changes.
4. Focused runs, split per package, with real exit codes:
   - `go test ./internal/envprofile -run 'Store|Boundary|Pin|Resolve|Status|Legacy'`
   - `go test ./internal/pathboundary ./internal/stateread`
   - `go test ./cmd/curator -run 'Env'`
   - `GOOS=windows go vet ./internal/envprofile ./internal/pathboundary`
5. Append "Revision 6 — re-apply on <trunk sha>" to the results, run `resource update`, then
   `task-board handoff TASK-260910-32gki6 --role developer`, then END YOUR TURN. The runner publishes the CR and runs the gate; do not wait
   for it. No CHANGELOG/LOGBOOK edit.
