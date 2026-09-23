# Integration instruction 4 — TASK-260908-1bfk8y (bound developer run, curator)

Revision 1 (comment-only change to `.github/ci/platform-exclusions.tsv`) is ACCEPTED. Attempt 2 failed only
because `worktree converge` requires `--reason`. The control root trunk is now at `6c19e5ee` (2qvzwk delivered).
No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator, in order:

    task-board worktree converge STORY-260907-2bddfc --reason "re-parent the uncommitted one-file comment change of accepted TASK-260908-1bfk8y rev1 onto trunk 6c19e5ee" 2>&1 | tee .temp/converge-2bddfc-3.log
    task-board worktree integrate STORY-260907-2bddfc --cr TASK-260908-1bfk8y --revision 1 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-1bfk8y-4.log

Attach both logs as outcome resources (`TASK-260908-1bfk8y_converge-3.md`, `TASK-260908-1bfk8y_integration-4.md`)
and stop. If the integrate refuses because the acceptance must be re-established on the new base (stale CR),
attach the refusal and stop — the orchestrator routes invalidate-acceptance + refresh. Change no file.
