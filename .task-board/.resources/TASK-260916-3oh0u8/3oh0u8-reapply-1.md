# TASK-260916-3oh0u8 — re-apply accepted rev2 on current trunk (THE ONLY CURRENT INSTRUCTION)

Revision 2 was ACCEPTED on content. Trunk moved to eca2bf27 (TASK-260910-2n0233 records boundary) and `worktree converge` conflicted on
cmd/curator/envstatus.go and cmd/curator/main.go. The exact accepted content is saved as `refs/campaign/2otjbn-rev2-20260927` (commit
36075014, parent d41da0fb, tree d13d9519 == CR rev2). Your Story worktree is fresh on trunk.
1. `task-board m 'set_status(TASK-260916-3oh0u8, status=development)'`.
2. `git diff d41da0fb refs/campaign/2otjbn-rev2-20260927 -- . ':!.task-board' > $TMPDIR/3oh0u8.patch; git apply --3way $TMPDIR/3oh0u8.patch`;
   resolve the 2 conflicts KEEPING BOTH SIDES (trunk's 2n0233 boundary-posture code + your trust-root code). Add NOTHING that is in neither
   side (a delta reviewer checks this with merge-tree).
3. VERIFY `git diff --name-only HEAD -- . ':!.task-board'` lists exactly the 6 rev2 paths; the 4 non-conflicting paths are byte-identical to rev2.
4. Focused runs with real exit codes: `go build ./...`; `go test ./cmd/curator -run 'Umbrella|Provider|EnvStatus|Status|Boundary|Attest'`.
5. Append "Revision 3 — re-apply on eca2bf27"; handoff and WAIT for the gate (never interrupt); hand off only green. No CHANGELOG/LOGBOOK.
