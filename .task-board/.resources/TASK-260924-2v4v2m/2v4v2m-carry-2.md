# TASK-260924-2v4v2m — republish after carry-forward (THE ONLY CURRENT INSTRUCTION; bound developer run)

Revision 1 was ACCEPTED on content. Trunk moved (now `a48f584c`), so the orchestrator ran `worktree converge STORY-260924-txgta4`: your
accepted delta is carried uncommitted in the Story worktree, combined by three-way merge with trunk's changes on the
intersecting paths listed by converge (typically CHANGELOG.md, sometimes a CI ledger or script). Content is not in question.
1. `task-board m 'set_status(TASK-260924-2v4v2m, status=development)'`.
2. Verify: every path of `TASK-260924-2v4v2m_change-request_rev1.patch` that trunk did NOT touch is byte-identical to revision 1; on
   the intersecting paths both sides are present (nothing dropped/duplicated; resolve a conflict marker keeping both).
   Report per path. Change nothing else.
3. CHANGELOG POLICY (orchestrator, 2026-09-24): revert this task's CHANGELOG.md hunk entirely (the file must equal trunk's); copy
   the entry text verbatim into a "## CHANGELOG entry (for release prep)" section of the results resource — the release-prep leaf
   writes all entries. Remove any stray root TASK-*/BUG-*.md, test/ or ledger/ path too.
4. Focused bounded run (go test ./internal/marker/... -count=1) with real exit codes.
5. Append "Revision 2 (carry-forward republish)" to `TASK-260924-2v4v2m_results.md`, check any unchecked DoD item citing it, then
   `task-board handoff TASK-260924-2v4v2m --role developer`; stay in the turn while the gate runs. A `run_wrote_outside_worktree …
   policy warn` block is a warning — verify status `to-review` and that revision 2 was published.
