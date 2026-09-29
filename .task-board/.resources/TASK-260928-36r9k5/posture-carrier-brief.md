# TASK-260928-36r9k5 — carrier re-apply: security_posture revision A + unreachable-registry cases (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md`. The accepted TASK-260927-4pv4au (security_posture revision A; review verdicts rev2/rev3) and the finished
TASK-260910-1sapuy (unreachable-registry-permissive-warns / -hardened-refuses on top of it; its results resource describes it) sit on a stale
Story branch the board cannot land. The orchestrator saved the combined work as `refs/campaign/2qmrb8-combined-20260928` (parent 6bd98d49).
Your Story worktree is fresh on trunk (E1, E3, E4, E5, E6, 2n0233, ryh3kw, global adopt, 3ed9m3 landed since 6bd98d49).
1. `task-board m 'set_status(TASK-260928-36r9k5, status=development)'`.
2. `git diff 6bd98d49 refs/campaign/2qmrb8-combined-20260928 -- . ':!.task-board' > $TMPDIR/posture.patch; git apply --3way $TMPDIR/posture.patch`;
   resolve KEEPING BOTH SIDES; trunk is authoritative for everything the landed leaves changed (install/registry/status/main.go, managed
   writes via E5 nofollow helpers, manager-state reads via internal/stateread). Ledgers (.github/ci/*.tsv): start from TRUNK's file and apply
   ONLY this content's rows — never re-add a row trunk removed (the gap ratchet fails on it). Add nothing else.
3. VERIFY `git diff --name-only origin/main -- . ':!.task-board'` = the combined paths only; the permissive warning is stderr-only and never
   on the enforced launch path (4pv4au review F1).
4. Focused runs with real exit codes: `go test ./internal/config -run 'Posture|Security|Hardened|Permissive'`, `go test ./cmd/curator -run
   'SecurityPosture|StatusJSON|Enforced'`, `go test ./internal/install -run 'Posture|Unreachable|Registry'` (split), `go test
   ./internal/scriptworker -run Enforced`, `go test ./internal/envprofile -run Guarded`.
5. Results: say which parts are 4pv4au (accepted) vs 1sapuy (new); copy 1sapuy's CHANGELOG entry text. `resource update`, handoff, END YOUR
   TURN. No CHANGELOG/LOGBOOK edit.
