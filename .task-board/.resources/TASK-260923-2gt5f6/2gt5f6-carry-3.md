# TASK-260923-2gt5f6 — republish after carry-forward (THE ONLY CURRENT INSTRUCTION; bound developer run)

Revision 2 was ACCEPTED on content. Trunk moved (now `e8620502`), so the orchestrator ran `worktree converge STORY-260923-1v3no2`: your
accepted delta is carried uncommitted in the Story worktree, combined by three-way merge with trunk's changes on the
intersecting paths listed by converge (typically CHANGELOG.md, sometimes a CI ledger or script). Content is not in question.
1. `task-board m 'set_status(TASK-260923-2gt5f6, status=development)'`.
2. Verify: every path of `TASK-260923-2gt5f6_change-request_rev2.patch` that trunk did NOT touch is byte-identical to revision 2; on
   the intersecting paths both sides are present (nothing dropped/duplicated; resolve a conflict marker keeping both).
   Report per path. Change nothing else.
3. CHANGELOG POLICY (orchestrator, 2026-09-24): revert this task's CHANGELOG.md hunk entirely (the file must equal trunk's); copy
   the entry text verbatim into a "## CHANGELOG entry (for release prep)" section of the results resource — the release-prep leaf
   writes all entries. Remove any stray root TASK-*/BUG-*.md, test/ or ledger/ path too.
4. VERIFY the worktree is not a stale snapshot: `git diff --name-only HEAD -- . ':!.task-board'` must list ONLY the paths of
   `TASK-260923-2gt5f6_change-request_rev2.patch` (minus CHANGELOG.md). If it lists any other file (a revert of trunk), STOP and report — do not publish.
5. Focused bounded run (go test ./internal/envprofile -run 'Credential|Record|Marker' -count=1 and ./cmd/curator -run 'Env|Credential' -count=1) with real exit codes.
6. Append "Revision 3 (carry-forward republish)" to `TASK-260923-2gt5f6_results.md`, check any unchecked DoD item citing it, then
   `task-board handoff TASK-260923-2gt5f6 --role developer`; stay in the turn while the gate runs. A `run_wrote_outside_worktree …
   policy warn` block is a warning — verify status `to-review` and that revision 3 was published.
