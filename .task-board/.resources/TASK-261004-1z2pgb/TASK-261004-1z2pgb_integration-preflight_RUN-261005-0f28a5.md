# Integration preflight — shell-hook-no-source-and-path-append-design

TASK-261004-1z2pgb — shell-hook-no-source-and-path-append-design
Date: 2026-10-05
Bound run: RUN-261005-0f28a5; role researcher, archetype analyst.

The latest Integration Assignment supersedes the earlier direct integrate and generic lifecycle instructions. No status mutation, handoff, checkpoint or integration command was invoked.

## Observed preconditions
- Board query: integrating (exit 0).
- Workspace status: accepted revision 1, kind story_final, repository delta present; producer binding researcher/analyst (exit 0).
- Accepted candidate tree: 685f82e26cc7dc73e34027422e0582c93bc41c4a.
- HEAD equals recorded checkpoint/base: 54bed271b7609bf206a04369202473c430d0d96a.
- Current run holds the workspace lease. Workspace registered, branch present, checkpoint reachable.
- Git status shows only the two expected untracked .research documents. No tracked changes, product code changes or LOGBOOK.md changes.
- Standalone Python byte comparison against git show of the accepted candidate: 2/2 documents identical, exit 0.
  - CIP draft SHA-256: d7ac7832726d05a4b747ff38f70eafbecd1c9d79f3fcc8dce8169b4f37d74bc3.
  - Evidence SHA-256: 30bcdfd9d4a5a3b021d1ab7406f87a6f07304945901c8590c232be24087731b3.
- git diff --check: exit 0; this checks tracked diffs only, not the untracked documents. Document identity is covered by the byte comparison.
- An attempted query using unsupported field change_request returned exit 1; recovered using worktree status. This was a query syntax failure, not a landing refusal.

## Bounds
No build or tests rerun: hosted-evidence instruction prohibits local Go gates; this run verifies unchanged accepted document bytes and relies on recorded acceptance, without claiming fresh test results. Workspace status includes unrelated uncommitted board activity; runner must apply the authoritative landing gates. Protected authority freshness and transaction success are not established by these local checks.

No repository files changed by this run. No credentials read or printed. Fresh task-scoped evidence attached before exit. Landing remains the synchronous runner's responsibility; status remains integrating.
