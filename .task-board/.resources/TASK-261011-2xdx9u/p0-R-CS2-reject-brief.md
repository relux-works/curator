# THE ONLY CURRENT INSTRUCTION — route TASK-261011-2xdx9u rev2 back to rework (reviewer, read-only)

Rev2 edits `decisions/0021-sessions-enter-through-curator-run.md`. That file belongs to TASK-261011-kojgyg, which has landed on curator-spec main (ddafa1ff). Trunk has therefore advanced on a path this candidate changes, and the revision cannot be reviewed on its current base.

Confirm that with `git diff --name-only <rev2 base> <rev2 candidate tree>`. Then call `reject_cr` (changes requested) with one finding:
- severity `robustness`;
- severity_reason "P1: candidate edits decisions/0021, which another task owns and has landed";
- required change: "drop the decisions/0021 edits and refresh onto current trunk with `worktree refresh-candidate`, keeping all CIP changes".

Do not review the content: the full R138 pair reviews rev3. END YOUR TURN.
