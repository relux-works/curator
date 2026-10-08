# THE ONLY CURRENT INSTRUCTION — TASK-260927-1e5qqm: revision-B flip (developer, code; LANDING HELD)
Revision A shipped in curator v0.15.0-rc.3, which satisfies the warning-release prerequisite. Implement: flip the registry CodexSeedRevision to B and drive the B vectors; remove the flip-leaf gap rows. Keep it production-entry tested with exact conformance counts.
LANDING IS HELD: the operator decides when the B release ships. Produce and review now; the orchestrator does not land it until scheduled.
## Mode (tb-R181, after a healthy mini restart)
Full codex pace. Run targeted tests with GOFLAGS=-work and keep the build lock. The hosted CR gate is the arbiter. Never edit LOGBOOK.md or CHANGELOG.md. Before the handoff, append a section to your task results resource and `resource update` it, because a handoff without a NEW or UPDATED outcome builds no Change Request.
Then `task-board handoff TASK-260927-1e5qqm --role developer` and END YOUR TURN.
