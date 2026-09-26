# TASK-260922-cww1ov — refresh 12 onto trunk 60498052 (THE ONLY CURRENT INSTRUCTION)

Revision 11 was ACCEPTED (verdict rev11). Trunk moved to `60498052`: 11burj (internal/install/draftsources.go —
declaredDependencyReplaySources), 1f2ng0 (.github/ci/platform-cases.tsv, cmd/curator/profile_test.go, envprofile), 18ex37, 10d3l1, 11jgkt.
Safety ref: refs/campaign/cww1ov-rev11-delta-20260926.
1. `task-board m 'set_status(TASK-260922-cww1ov, status=development)'` if needed.
2. Combine trunk: `git diff 0a628621 60498052 -- . ':!.task-board' ':!CHANGELOG.md' ':!LOGBOOK.md' | git apply --3way`; keep BOTH sides on
   draftsources.go, platform-cases.tsv, profile_test.go and any overlap; migrate any trunk-added state reader (11burj's replay recovery,
   1f2ng0 envprofile changes) onto the stateread seam and list it. Leave nothing staged.
3. VERIFY `git diff --name-only 60498052 -- . ':!.task-board'` over the working tree == this Story's paths only (no trunk revert, no
   LOGBOOK.md/CHANGELOG.md, no stray files).
4. `task-board worktree refresh-candidate TASK-260922-cww1ov` (F-C1/F-C2 replay via `--replay-resolutions` only).
5. Bounded runs: 0017 rows, `go test ./internal/envprofile ./internal/stateread ./cmd/curator -count=1` in parts, the playbook acceptance
   row, go vet, `GOOS=windows go vet ./internal/envprofile ./internal/install`. Real exit codes.
6. Append "Revision 12 — refresh onto 60498052", `resource update`, `task-board handoff TASK-260922-cww1ov --role developer`; stay in the turn
   while the gate runs. A write-boundary `policy warn` block is a warning.
