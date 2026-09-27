# TASK-260918-ryh3kw — re-apply rev4 on trunk + drive the restore vectors (THE ONLY CURRENT INSTRUCTION)

Review rev4 = CHANGES REQUESTED (`TASK-260918-ryh3kw_review-verdict-rev4.md`): content fine and gate green, but it was never combined with
trunk and the ledger names a blocker that no longer exists. The orchestrator saved rev4 as `refs/campaign/1ll22r-rev4-20260927` (parent
0ffe2e1d, tree 93f1102d == CR rev4) and discarded the old workspace; your Story worktree is fresh on trunk (≥ eca2bf27).
1. `task-board m 'set_status(TASK-260918-ryh3kw, status=development)'`.
2. `git diff 0ffe2e1d refs/campaign/1ll22r-rev4-20260927 -- . ':!.task-board' > $TMPDIR/ryh3kw.patch; git apply --3way $TMPDIR/ryh3kw.patch`;
   resolve conflicts KEEPING BOTH SIDES (trunk: 1wc76r restore-backups, E2 55g9dg, 2n0233, gocke2 …); add nothing in neither side.
3. The 2 environments-read-failure restore vectors (backup-record-absent-restore-nothing, backup-record-unreadable-restore-stops) are
   ledgered against TASK-260927-1wc76r "no env unmanage --restore-backups production entry" — that entry now exists (cmd/curator/env.go
   `--restore-backups`). Drive both through it and remove the rows; kill one mutant each (real exit codes). If one genuinely cannot be
   driven, re-attribute it to the concrete blocker with an owner.
4. VERIFY `git diff --name-only origin/main -- . ':!.task-board'` lists only your paths (the rev4 18 + at most the new rows' test file).
5. Focused runs (real exit codes): `go test ./internal/stateread`, `go test ./internal/envprofile -run 'ReadFailure|Guarded|Restore|Unmanage'`,
   `go test ./cmd/curator -run 'Unmanage|Profile.*Operand|Status'`.
6. Append "Revision 5 — combined with trunk; restore vectors driven"; handoff and WAIT for the gate; hand off only green. No CHANGELOG/LOGBOOK.
