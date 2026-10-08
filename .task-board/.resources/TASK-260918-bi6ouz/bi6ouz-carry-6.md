# TASK-260918-bi6ouz — republish after carry-forward (THE ONLY CURRENT INSTRUCTION; bound developer run)

Revision 5 was ACCEPTED on content. Trunk moved (now `0a628621`), so the orchestrator ran `worktree converge STORY-260916-12lbww`: your
accepted delta is carried uncommitted in the Story worktree, combined by three-way merge with trunk's changes on the
intersecting paths listed by converge (typically CHANGELOG.md, sometimes a CI ledger or script). Content is not in question.
1. `task-board m 'set_status(TASK-260918-bi6ouz, status=development)'`.
2. Verify: every path of `TASK-260918-bi6ouz_change-request_rev5.patch` that trunk did NOT touch is byte-identical to revision 5; on
   the intersecting paths both sides are present (nothing dropped/duplicated; resolve a conflict marker keeping both).
   Report per path. Change nothing else.
3. CHANGELOG POLICY (orchestrator, 2026-09-24): revert this task's CHANGELOG.md hunk entirely (the file must equal trunk's); copy
   the entry text verbatim into a "## CHANGELOG entry (for release prep)" section of the results resource — the release-prep leaf
   writes all entries. Remove any stray root TASK-*/BUG-*.md, test/ or ledger/ path too.
4. VERIFY the worktree is not a stale snapshot: `git diff --name-only HEAD -- . ':!.task-board'` must list ONLY the paths of
   `TASK-260918-bi6ouz_change-request_rev5.patch` (minus CHANGELOG.md). If it lists any other file (a revert of trunk), STOP and report — do not publish.
5. Focused bounded run (go test ./internal/envprofile/... -count=1) with real exit codes.
6. Append "Revision 6 (carry-forward republish)" to `TASK-260918-bi6ouz_results.md`, check any unchecked DoD item citing it, then
   `task-board handoff TASK-260918-bi6ouz --role developer`; stay in the turn while the gate runs. A `run_wrote_outside_worktree …
   policy warn` block is a warning — verify status `to-review` and that revision 6 was published.
