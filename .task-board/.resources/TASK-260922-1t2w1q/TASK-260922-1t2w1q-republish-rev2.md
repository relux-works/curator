# Republish unchanged — TASK-260922-1t2w1q revision 2 → 3 (bound developer run)

The rev2 gate (run 35702368558) failed only on windows-latest in a package this leaf does not touch:
`internal/managerlock :: TestSubprocessExpectedAcquiredWithTinyDeadlineReportsBlocked`
("uncontended helper with tiny deadline = acquired, want blocked") — a timing-sensitive row on the
hosted Windows runner, unrelated to this candidate (patch = cmd/curator env*, internal/envprofile,
internal/envregistry, docs, CHANGELOG). Do NOT change any file. Verify `git status --short` in the
Story worktree (.temp/STORY-260922-1cenbr/worktree) lists the rev2 paths only, add one line to
results.md ("revision 3 = revision 2 unchanged; gate rerun after a Windows managerlock timing
flake"), then `task-board handoff TASK-260922-1t2w1q --role developer`. If the rerun fails on the
SAME test again, attach the failure and stop (the orchestrator files the flake).
