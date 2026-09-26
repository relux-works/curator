# TASK-260922-cww1ov — republish after carry-forward (THE ONLY CURRENT INSTRUCTION; bound developer run)

Revision 5 was ACCEPTED on content. Trunk moved (now `1511b345`), so the orchestrator ran `worktree converge STORY-260922-1cenbr`: your
accepted delta is carried uncommitted in the Story worktree, combined by three-way merge with trunk's changes on the
intersecting paths listed by converge (typically CHANGELOG.md, sometimes a CI ledger or script). Content is not in question.
1. `task-board m 'set_status(TASK-260922-cww1ov, status=development)'`.
2. Verify: every path of `TASK-260922-cww1ov_change-request_rev5.patch` that trunk did NOT touch is byte-identical to revision 5; on
   the intersecting paths both sides are present (nothing dropped/duplicated; resolve a conflict marker keeping both).
   Report per path. Change nothing else.
3. Focused bounded run (go test ./internal/envprofile -run 'Credential|Isolat|Passthrough|Migrat') with real exit codes.
4. Append "Revision 8 (carry-forward republish)" to `TASK-260922-cww1ov_results.md`, check any unchecked DoD item citing it, then
   `task-board handoff TASK-260922-cww1ov --role developer`; stay in the turn while the gate runs. A `run_wrote_outside_worktree …
   policy warn` block is a warning — verify status `to-review` and that revision 8 was published.
