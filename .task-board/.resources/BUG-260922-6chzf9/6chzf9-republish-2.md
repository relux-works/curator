# BUG-260922-6chzf9 — republish as the Story's final revision (THE ONLY CURRENT INSTRUCTION; bound developer run)

Revision 2 (story_final) was ACCEPTED, but trunk moved again (2kqa77 landed, CHANGELOG only), so the orchestrator carried it forward; a
task_delta last leaf neither by checkpoint (final-leaf refusal) nor by integrate. The orchestrator ran
`worktree converge` onto trunk `1511b345`: your accepted delta is carried uncommitted in the Story worktree
(CHANGELOG.md combined with trunk's). Content is NOT in question.

1. `task-board m 'set_status(BUG-260922-6chzf9, status=development)'`.
2. Verify the carried delta equals revision 2 except the CHANGELOG combination: `git status --short`, and
   per-file `git diff` of the non-CHANGELOG paths vs `BUG-260922-6chzf9_change-request_rev2.patch` (report identity).
   Do NOT change any file (a CHANGELOG conflict marker, if any, is the only thing you may resolve, keeping both sides).
3. Focused bounded run of the touched package/scripts (internal/managerlock) with real exit codes.
4. Append "Revision 3 (refresh republish)" to `BUG-260922-6chzf9_results.md`, check any unchecked DoD items citing it, then
   `task-board handoff BUG-260922-6chzf9 --role developer`. The runtime publishes revision 3 as the Story's final revision and
   runs the gate. A `run_wrote_outside_worktree … policy warn` block is a warning — verify status `to-review`.
