# TASK-260922-cww1ov — rework 11: revert LOGBOOK.md + refresh onto trunk 0a628621 (THE ONLY CURRENT INSTRUCTION)

Revision 10 CHANGES_REQUESTED, F1 (verdict rev10): producers never edit LOGBOOK.md — this Story added two hunks (LOGBOOK.md:6-11 and
4626-4636). Points 1–3 of the rev10 note were not re-verified and will be on the next revision.
1. `task-board m 'set_status(TASK-260922-cww1ov, status=development)'`. Safety: record `git stash create` oid in results.
2. Make LOGBOOK.md equal trunk bytes (`git show 0a628621:LOGBOOK.md > LOGBOOK.md` after step 3's combine, or restore it from HEAD of the
   refreshed base); move that content into your results resource under "## Notes (moved from LOGBOOK)".
3. Combine trunk: `git diff faf509ae 0a628621 -- . ':!.task-board' ':!CHANGELOG.md' | git apply --3way` (keep both sides; migrate any
   trunk-added state reader onto the stateread seam and list it). Leave nothing staged. VERIFY `git diff --name-only 0a628621` over the
   working tree lists only this Story's paths (no trunk revert, no LOGBOOK.md, no CHANGELOG.md, no stray files).
4. `task-board worktree refresh-candidate TASK-260922-cww1ov` (F-C1/F-C2 replay via `--replay-resolutions` only; CHANGELOG → trunk bytes).
5. Bounded runs: 0017 rows, `go test ./internal/envprofile ./internal/stateread ./cmd/curator -count=1` in parts, go vet, `GOOS=windows go
   vet ./internal/envprofile ./internal/install`. Real exit codes.
6. Append "Revision 11 — LOGBOOK reverted, refresh onto 0a628621", `resource update`, `task-board handoff TASK-260922-cww1ov --role developer`;
   stay in the turn while the gate runs. A write-boundary `policy warn` block is a warning.
