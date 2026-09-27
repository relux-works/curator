# TASK-260916-1zgucp — re-apply accepted rev5 on trunk 86552087 (THE ONLY CURRENT INSTRUCTION)

Revision 5 was ACCEPTED (delta review 2). Trunk moved (E4 provider roots, E3 Codex seed MCP, ryh3kw §8.4.1 read-failure discipline, 2n0233)
and converge conflicts (internal/envprofile/status.go and ledgers). The accepted content is `refs/campaign/ioemse-rev5-20260927` (parent
d41da0fb, tree 1b69fd0c). Your Story worktree is fresh on trunk.
1. `task-board m 'set_status(TASK-260916-1zgucp, status=development)'`.
2. `git diff d41da0fb refs/campaign/ioemse-rev5-20260927 -- . ':!.task-board' > $TMPDIR/1zgucp.patch; git apply --3way $TMPDIR/1zgucp.patch`;
   resolve KEEPING BOTH SIDES; add nothing in neither side EXCEPT: any new manager-state absence read your code performs must now go
   through internal/stateread (ryh3kw landed the deny-by-default discipline — TestManagerOwnedAbsenceReadsAreGuarded must pass); name
   such routing changes in the results.
3. VERIFY `git diff --name-only HEAD -- . ':!.task-board'` = the 31 rev5 paths; paths trunk did not touch byte-identical to rev5.
4. Focused runs (real exit codes): `go build ./...`; `go test ./internal/config`; `go test ./internal/contextresolve ./internal/contextlock`;
   `go test ./internal/envprofile -run 'Status|Surfacing|Signer|Delta|Guarded'`; `go test ./cmd/curator -run 'Profile|Signer|Delta|Status'`.
5. Append "Revision 6 — re-apply on 86552087" to the results and `resource update` it (a re-handoff without an updated outcome publishes
   nothing); handoff and WAIT for the gate; hand off only green. No CHANGELOG/LOGBOOK edit.
