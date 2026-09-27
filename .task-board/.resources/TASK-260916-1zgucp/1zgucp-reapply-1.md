# TASK-260916-1zgucp — re-apply accepted rev3 on current trunk (THE ONLY CURRENT INSTRUCTION)

Revision 3 was ACCEPTED on content (review verdict rev3). Trunk moved to d41da0fb (E2 direct-only system modules, 1wc76r restore backups,
gocke2 MCP surfacing, 3aco9f …) and `worktree converge` conflicted on 6 paths (.github/ci/conformance-gaps.tsv, internal/config/config.go,
internal/config/environments.go, internal/config/environments_conformance_test.go, …). The orchestrator saved the exact accepted content as
`refs/campaign/ioemse-full-20260927` (commit 8f0bf583, parent bd3c0f43, tree 78bcb664 == CR rev3 candidate) and discarded the old workspace.
Your Story worktree is fresh on trunk.
1. `task-board m 'set_status(TASK-260916-1zgucp, status=development)'`.
2. In the worktree: `git diff bd3c0f43 refs/campaign/ioemse-full-20260927 -- . ':!.task-board' > $TMPDIR/1zgucp.patch` then
   `git apply --3way $TMPDIR/1zgucp.patch`; resolve every conflict KEEPING BOTH SIDES (trunk's new code + your accepted delta), no markers left.
   On conformance-gaps.tsv: keep trunk's rows, remove only the rows your rev3 removed (E1-owned), re-add none of trunk's removals.
3. VERIFY: `git diff --name-only HEAD -- . ':!.task-board'` lists exactly the 31 rev3 paths (no other file → no trunk revert); on paths trunk did
   not touch the bytes equal rev3; report per conflicted path what trunk changed and how you combined.
4. Focused bounded runs with real exit codes: `go build ./...`; `go test ./internal/contextresolve ./internal/contextlock` ; `go test
   ./internal/config`; `go test ./internal/envprofile -run 'Signer|Delta|Status|Guarded|StateRead|Surfacing'`; `go test ./cmd/curator -run
   'Profile|Signer|Delta|Status'`.
5. Append "Revision 4 — re-apply on d41da0fb" to the results, handoff. The hosted gate on the published revision must be green before handoff
   (campaign-producer-rules.md "Hosted gate before handoff"); if red, fix and republish. No CHANGELOG/LOGBOOK edit. Stay in the turn.
