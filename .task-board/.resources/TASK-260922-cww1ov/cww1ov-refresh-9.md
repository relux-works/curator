# TASK-260922-cww1ov — refresh onto trunk ab34556e and publish (THE ONLY CURRENT INSTRUCTION)

Revision 8 (0017 credential modes, refresh onto 1511b345) is withdrawn by the orchestrator: trunk has since moved to `ab34556e` (R5
script worker, 6chzf9, 3mazfw, 1kpw4w, 5p8b0z, m28s6b+1aa9wb Skillfile default-on and lock replay, 2v4v2m, 3v8k23, krcm6m). Your
work is complete and uncommitted in the Story worktree; the Story checkpoints (F-C1 1t551d, F-C2 1t2w1q) sit on 1511b345.
1. `task-board m 'set_status(TASK-260922-cww1ov, status=development)'` if needed. Safety first: `git stash create` → record the
   oid in your results (do not pop/drop anything).
2. Combine trunk: `git diff 1511b345 ab34556e -- . ':!.task-board' ':!CHANGELOG.md' | git apply --3way` in the Story worktree. Keep BOTH
   sides everywhere; where trunk added or changed a state reader your stateread migration covers (R5 script-worker readers, lock
   replay in internal/install/draftsources.go), migrate trunk's version onto the seam too and list each one. Leave nothing staged.
3. CHANGELOG POLICY: CHANGELOG.md must equal trunk's bytes (this also drops the F-C1/F-C2 checkpoint entries); put every entry text of
   this Story verbatim in results under "## CHANGELOG entry (for release prep)". Remove any stray root TASK-*/BUG-* file, test/ or
   ledger/ path.
4. `task-board worktree refresh-candidate TASK-260922-cww1ov`; checkpoint replay conflicts only via its `--replay-resolutions`
   template (CHANGELOG → trunk bytes); never hand-commit the replay worktree.
5. Bounded re-runs: your 0017 rows (hazards, migration/recovery, no-copy scan), `go test ./internal/envprofile ./cmd/curator -count=1`
   in parts, `go vet` on touched packages, `GOOS=windows go vet ./internal/envprofile`. Real exit codes.
6. Append "Revision 9 — refresh onto ab34556e" to results, `resource update`, `task-board handoff TASK-260922-cww1ov --role developer`;
   stay in the turn while the gate runs. A `run_wrote_outside_worktree … policy warn` block is a warning.
