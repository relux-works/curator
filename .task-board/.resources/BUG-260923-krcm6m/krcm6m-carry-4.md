# BUG-260923-krcm6m — republish after carry-forward (THE ONLY CURRENT INSTRUCTION; bound developer run)

Revision 3 was ACCEPTED on content. Trunk moved (now `948ae7c9`), so the orchestrator ran `worktree converge STORY-260923-3qwrnl`: your
accepted delta is carried uncommitted in the Story worktree, combined by three-way merge with trunk's changes on the
intersecting paths listed by converge (typically CHANGELOG.md, sometimes a CI ledger or script). Content is not in question.
1. `task-board m 'set_status(BUG-260923-krcm6m, status=development)'`.
2. Verify: every path of `BUG-260923-krcm6m_change-request_rev3.patch` that trunk did NOT touch is byte-identical to revision 3; on
   the intersecting paths both sides are present (nothing dropped/duplicated; resolve a conflict marker keeping both).
   Report per path. Change nothing else.
3. CHANGELOG POLICY (orchestrator, 2026-09-24): revert this task's CHANGELOG.md hunk entirely (the file must equal trunk's); copy
   the entry text verbatim into a "## CHANGELOG entry (for release prep)" section of the results resource — the release-prep leaf
   writes all entries. Remove any stray root TASK-*/BUG-*.md, test/ or ledger/ path too.
4. Focused bounded run (go test ./internal/registry -run FutureBound -count=20) with real exit codes.
5. Append "Revision 4 (carry-forward republish)" to `BUG-260923-krcm6m_results.md`, check any unchecked DoD item citing it, then
   `task-board handoff BUG-260923-krcm6m --role developer`; stay in the turn while the gate runs. A `run_wrote_outside_worktree …
   policy warn` block is a warning — verify status `to-review` and that revision 4 was published.
