# TASK-260907-2as5sx — rework 5 + refresh onto trunk (THE ONLY CURRENT INSTRUCTION)

Revision 6 (deny-by-default guard) failed its gate (run 35969169433) ONLY on macOS:
    internal/install TestAnInFlightTransactionKeepsThePublishedCacheEntry: commit_test.go:534: a run that kept what it rebuilt did not say so in its result
1. Decide with evidence whether your seam changed this path (a read in install/buildcache now reporting absent/unreadable differently so
   the "kept rebuilt entry" outcome is no longer recorded) or it is timing: run it `-count=30` on the base and on your candidate; read the
   code path from the test to the result field. If yours, fix the cause in production (keep the §8.4 distinction); if a pre-existing flake,
   say so with the base-run evidence and leave the test untouched.
2. CHANGELOG POLICY (2026-09-24): revert this Story's CHANGELOG.md hunk to trunk bytes; put the entry text in the results resource under
   "## CHANGELOG entry (for release prep)".
3. Refresh onto trunk `948ae7c9` (R5 landed): combine `git diff 1511b345 948ae7c9 -- . ':!.task-board' ':!CHANGELOG.md'` into the worktree (3-way;
   keep both sides — R5 touched internal/install and godriver: migrate any new manager-state reader it added onto your seam or allowlist it
   with a reason so the deny-by-default guard stays green), leave nothing staged, then `task-board worktree refresh-candidate
   TASK-260907-2as5sx` (replay conflicts only via its template).
4. Bounded runs: the failing test (-count=30), the guard test, `go test ./internal/install/... ./internal/envprofile/...` per package,
   gofmt. Results stay a board resource. Append "Revision 7", `resource update`, `task-board handoff TASK-260907-2as5sx --role developer`.
A `run_wrote_outside_worktree … policy warn` block is a warning — verify status `to-review`.
