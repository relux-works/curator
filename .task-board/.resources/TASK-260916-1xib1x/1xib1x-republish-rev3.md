# Republish unchanged — TASK-260916-1xib1x revision 3 → 4 (bound developer run)

The rev3 gate (run 35709048693) failed only on windows-latest in a package this leaf does not touch:
`internal/managerlock :: TestSubprocessExpectedAcquiredWithTinyDeadlineReportsBlocked`
("uncontended helper with tiny deadline = acquired, want blocked") — a timing-sensitive row on the
hosted Windows runner, unrelated to this candidate (patch = audit labels: internal/scriptpolicy, internal/audit, internal/install, cmd/curator,
internal/envregistry, docs, CHANGELOG). Do NOT change any file. Verify `git status --short` in the
Story worktree (.temp/STORY-260822-2h0v9j/worktree) lists the rev3 paths only, add one line to
results.md ("revision 4 = revision 3 unchanged; gate rerun after a Windows managerlock timing
flake"), then `task-board handoff TASK-260916-1xib1x --role developer`. If the rerun fails on the
SAME test again, attach the failure and stop (the orchestrator files the flake).
