# THE ONLY CURRENT INSTRUCTION — TASK-261013-g96cen: complete the handoff of the reviewed tree (W-CS7r3, developer)

**Scope decision (orchestrator, 2026-10-13).** The board's configured handoff gate (`spec-gate.sh`) is allowed here.
- Its Go step is compile-only (R223); the hosted CI runs the Go tests.
- Its Python step is the specification's own source checks.
- Wave 5 passed the same gate.

The earlier "no tests or builds" rule meant no other local test suites; it does not forbid this gate.

**Do this:**
1. Do not change any content. The candidate tree `0ac03793` in the Story worktree is the one reviewer B reviewed and would accept (`TASK-261013-g96cen_review-verdict-rev2.md`).
2. Run `task-board handoff TASK-261013-g96cen --role developer` normally, letting the configured gate run, so that a `ready` revision exists.
3. If the gate fails, do not work around it. Attach its log and stop.
4. END YOUR TURN.
