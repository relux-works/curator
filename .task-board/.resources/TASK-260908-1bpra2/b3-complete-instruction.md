# Integration instruction — TASK-260908-1bpra2 (bound developer run)

Accepted revision 1 (candidate tree f4129d308ffb264fdf98ba7dcb5631199fd8d1b8) is ALREADY LANDED on relux-mcp main as commit 027f55b7582591604c8dc963684879671c95dbcd (PR #1, fast-forward) and tagged figma/v1.0.0 and safari/v1.0.0. Board owner is separate: do NOT run integrate, checkpoint, handoff or set_status. Run exactly, from your control root, and attach the full output as `TASK-260908-1bpra2_integration-results.md`:

    task-board worktree complete STORY-260908-2a4936 --cr TASK-260908-1bpra2 --revision 1 --landed-commit 027f55b7582591604c8dc963684879671c95dbcd

Note: this leaf is NOT the Story's final leaf (TASK-260916-bn5kvb remains), so if `complete` refuses because the revision is a task_delta, run `task-board worktree checkpoint TASK-260908-1bpra2` instead and attach that output. If both refuse, attach the exact refusals and stop. No code changes.
