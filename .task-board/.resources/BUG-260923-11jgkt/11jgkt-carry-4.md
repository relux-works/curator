# BUG-260923-11jgkt — republish after carry-forward (THE ONLY CURRENT INSTRUCTION; bound developer run)

Revision 1 was ACCEPTED on content. Trunk moved (now `1511b345`), so the orchestrator ran `worktree converge STORY-260923-laeycm`: your
accepted delta is carried uncommitted in the Story worktree, combined by three-way merge with trunk's changes on the
intersecting paths listed by converge (typically CHANGELOG.md, sometimes a CI ledger or script). Content is not in question.
1. `task-board m 'set_status(BUG-260923-11jgkt, status=development)'`.
2. Verify: every path of `BUG-260923-11jgkt_change-request_rev1.patch` that trunk did NOT touch is byte-identical to revision 1; on
   the intersecting paths both sides are present (nothing dropped/duplicated; resolve a conflict marker keeping both).
   Report per path. Change nothing else.
3. Focused bounded run (go test ./internal/snapshot/...) with real exit codes.
4. Append "Revision 4 (carry-forward republish)" to `BUG-260923-11jgkt_results.md`, check any unchecked DoD item citing it, then
   `task-board handoff BUG-260923-11jgkt --role developer`; stay in the turn while the gate runs. A `run_wrote_outside_worktree …
   policy warn` block is a warning — verify status `to-review` and that revision 4 was published.
