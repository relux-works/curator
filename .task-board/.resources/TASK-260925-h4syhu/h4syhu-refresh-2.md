# TASK-260925-h4syhu — refresh 2 onto trunk 9f0da708 and republish (THE ONLY CURRENT INSTRUCTION)

Revision 2 was ACCEPTED (verdict rev2). Integration refused: stale — trunk moved from ab34556e to `9f0da708` (18ex37 gap ledger/pins;
11burj changed internal/install/draftsources.go: declaredDependencyReplaySources for transitive replay; 10d3l1; 11jgkt). Safety ref:
refs/campaign/h4syhu-rev2-delta-20260926.
1. `task-board m 'set_status(TASK-260925-h4syhu, status=development)'` if needed.
2. Combine trunk: `git diff ab34556e 9f0da708 -- . ':!.task-board' ':!CHANGELOG.md' | git apply --3way`; keep BOTH sides on draftsources.go and
   any other overlap. The deny-by-default guard must also cover 11burj's new code: migrate any new collapse site onto the stateread seam or
   allow-list it with a reason; update counts/ratio honestly; list each. Leave nothing staged.
3. VERIFY `git diff --name-only 9f0da708 -- . ':!.task-board'` over the working tree lists only this Story's paths (no trunk revert, no
   CHANGELOG/LOGBOOK, no stray files).
4. `task-board worktree refresh-candidate TASK-260925-h4syhu` (187z6x replay via `--replay-resolutions` only).
5. Bounded runs: guard test, M1 kill, `go test ./internal/install -run 'LockedNetworkRepository|Draft|Playbook' -count=1` in parts, go vet,
   `GOOS=windows go vet ./internal/install`. Real exit codes.
6. Append "Revision 3 — refresh onto 9f0da708", `resource update`, `task-board handoff TASK-260925-h4syhu --role developer`; stay in the turn
   while the gate runs. A write-boundary `policy warn` block is a warning.
