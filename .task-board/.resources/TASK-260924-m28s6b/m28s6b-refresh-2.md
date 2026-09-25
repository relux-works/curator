# TASK-260924-m28s6b — refresh onto trunk a48f584c and republish (THE ONLY CURRENT INSTRUCTION; bound developer run)

Revision 3 content (Windows file-URL fixture fix, results "Revision 3") is complete and uncommitted in the Story worktree. Publication
refused `change_request_base_authority_mismatch`: the Story checkpoint c4ced3a9 (1aa9wb) sits on 5b326aa3, trunk is now `a48f584c`
(1kpw4w manifest dependency directory landed). Content is not in question.
1. `task-board m 'set_status(TASK-260924-m28s6b, status=development)'` if needed.
2. Combine trunk: `git diff 5b326aa3 a48f584c -- . ':!.task-board' ':!CHANGELOG.md' | git apply --3way` in the Story worktree (keep both
   sides on any conflict; report each conflicted path and how it was resolved), leave nothing staged, then
   `task-board worktree refresh-candidate TASK-260924-m28s6b` (checkpoint replay conflicts only via its `--replay-resolutions` template;
   CHANGELOG.md → trunk bytes). CHANGELOG.md must equal trunk's.
3. Bounded re-runs: your replay rows + `TestDraftReplayFileURLUsesPortableGitConfigSyntax` + `go test ./internal/manifest ./internal/closure
   -count=1` (1kpw4w areas) + `GOOS=windows go vet ./internal/install`. Real exit codes.
4. Append "Revision 4 — refresh onto a48f584c" to results, `resource update`, `task-board handoff TASK-260924-m28s6b --role developer`;
   stay in the turn while the gate runs. A `run_wrote_outside_worktree … policy warn` block is a warning.
