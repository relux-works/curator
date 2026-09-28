# TASK-260916-yvxbs1 — re-apply rev4 on trunk (THE ONLY CURRENT INSTRUCTION)

Rev4 (tree d23daad5) is gate-GREEN, but it sits on eca2bf27 and mixes hand-copied trunk content; trunk is now 86552087 (E4 provider roots,
E3 Codex seed MCP, ryh3kw §8.4.1 read-failure discipline). The orchestrator saved rev4 as `refs/campaign/wgt8vz-rev4-20260928` (parent
eca2bf27) and discarded the workspace; your Story worktree is fresh on trunk.
1. `task-board m 'set_status(TASK-260916-yvxbs1, status=development)'`.
2. `git diff eca2bf27 refs/campaign/wgt8vz-rev4-20260928 -- . ':!.task-board' > $TMPDIR/yvxbs1.patch; git apply --3way $TMPDIR/yvxbs1.patch`.
   Expected conflicts: root-artifacts.tsv, internal/envprofile/{envprofile.go, managed.go, status.go}, internal/envregistry/envregistry.go.
   Resolve KEEPING BOTH SIDES: trunk's current code is authoritative for everything E4/E3/ryh3kw changed; your E6 rules (path-kind MCP
   refusal, trusted direct path-overlay system modules admitted per rc.13 §3, path-source boundary verified at every env resolve and
   mutating op per environments §4, Windows fixtures protected) layered on top. Any manager-state absence read goes through
   internal/stateread (TestManagerOwnedAbsenceReadsAreGuarded). Where rev4 carried a hand-copied older version of a trunk file, take trunk's.
3. VERIFY `git diff --name-only origin/main -- . ':!.task-board'` lists ONLY E6's own paths (no revert of trunk: for each path trunk changed
   since eca2bf27, every trunk-added line must still be present unless E6 deliberately changes it — say which).
4. Focused runs with real exit codes: `go build ./...`; `go test ./internal/pathboundary`; `go test ./internal/envprofile -run
   'Path|Boundary|Overlay|Mcp|System|Update|Resolve|Status|Guarded'` (split); `go test ./cmd/curator -run 'Profile|Path|Status'`.
5. Append "Revision 5 — re-applied on 86552087" to the results, `resource update` it, `task-board handoff TASK-260916-yvxbs1 --role
   developer`, and END YOUR TURN (the runner publishes the CR and runs the gate). No CHANGELOG/LOGBOOK edit.
