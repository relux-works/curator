# TASK-260927-31gaka — re-apply your implementation on trunk, then hand off (THE ONLY CURRENT INSTRUCTION)

Your revision-A implementation was never published (the orchestrator's wrong "wait for the gate" instruction; see the corrected rule in
campaign-producer-rules.md). Trunk then moved to 86552087 (ryh3kw §8.4.1 read-failure discipline) touching managed.go, status.go,
envregistry.go, so the old workspace could not fast-forward. The orchestrator saved your exact work as `refs/campaign/3qf8er-impl-20260928`
(parent 6bd98d49) and discarded the workspace; your Story worktree is fresh on trunk.
1. `task-board m 'set_status(TASK-260927-31gaka, status=development)'`.
2. `git diff 6bd98d49 refs/campaign/3qf8er-impl-20260928 -- . ':!.task-board' > $TMPDIR/31gaka.patch; git apply --3way $TMPDIR/31gaka.patch`;
   resolve KEEPING BOTH SIDES; any manager-state absence read in your code goes through internal/stateread (ryh3kw's guard —
   TestManagerOwnedAbsenceReadsAreGuarded must pass). Add nothing else.
3. VERIFY `git diff --name-only origin/main -- . ':!.task-board'` lists only your 7 paths. Focused runs with real exit codes:
   `go test ./internal/envprofile -run 'Seed|Codex|Mcp|Status|Guarded'`, `go test ./cmd/curator -run 'EnvResolve|Marker|EnvStatus|Seed|Mcp'`.
4. Append "Revision 1 — re-applied on 86552087" to the results, `resource update` it, `task-board handoff TASK-260927-31gaka --role
   developer`, and END YOUR TURN (the runner publishes the CR and runs the gate after you exit). No CHANGELOG/LOGBOOK edit.
