# TASK-260907-2as5sx — rework 3 (orchestrator, binding). THIS IS THE ONLY CURRENT INSTRUCTION.

Revision 3 (gate run 35947681073) is green everywhere except Windows:
    cmd/curator TestContextExposureDistinguishesAbsentAndUnreadableBuildRoots: builds_test.go:533:
      unreadable build-root path = ("", <nil>), want typed unreadable error
chmod-based unreadability does not exist on Windows. Make this the LAST round of this kind: enumerate EVERY test (or subtest) this
Story added or changed that makes a path unreadable via POSIX mode bits (grep your delta for Chmod/0o000/0000/unreadable), and for
each one that is not already declared: skip on Windows with the registered reason and give it its own ledger row
`linux,darwin  windows  platform-control <POSIX mode-bit unreadability>` exactly like the three rows added in rework 1. Do NOT touch
shared DACL helpers or internal/godriver. List every such test in results with its ledger line.
Bounded runs; `sh .github/ci/ledger-consistency.sh`, `sh .github/ci/gate-selftest.sh`, cross-compile the Windows test binaries
(`GOOS=windows go test -c` for each touched package). Append "Revision 4" to results, `task-board resource update` it, then
`task-board handoff TASK-260907-2as5sx --role developer`. A `run_wrote_outside_worktree … policy warn` block is a warning — verify
status `to-review`.
