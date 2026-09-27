# TASK-260916-55g9dg — re-apply the accepted rev4 delta on a fresh workspace (THE ONLY CURRENT INSTRUCTION; bound developer run)

Revision 4 was ACCEPTED. The carry converge left the old files in place (a trunk revert) — you caught it, thank you. The orchestrator discarded
the workspace; your new one starts on trunk `97ca3370`. The accepted delta is `git diff bd3c0f43 refs/campaign/55g9dg-rev4-full-20260927 -- . ':!.task-board'` (8 paths).
1. `task-board m 'set_status(TASK-260916-55g9dg, status=development)'`.
2. Apply it with `git apply --3way`; on the intersecting paths keep BOTH sides (trunk's em42lw permissions/fragment-v2 + your E2 change);
   list every conflict and its resolution. Leave nothing staged.
3. VERIFY `git diff --name-only HEAD -- . ':!.task-board'` == the 8 rev4 paths and that no trunk content is reverted (diff each against HEAD).
4. Bounded runs: `go test ./internal/contextmaterialize ./internal/contextresolve -count=1`, the E2 vectors. Real exit codes.
5. Append "Revision 6 — re-apply on 97ca3370", `resource update`, `task-board handoff TASK-260916-55g9dg --role developer`; stay in the turn.
No CHANGELOG/LOGBOOK edit. A write-boundary `policy warn` block is a warning.
