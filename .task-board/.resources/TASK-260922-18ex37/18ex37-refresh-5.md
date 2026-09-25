# TASK-260922-18ex37 — refresh onto trunk ab34556e and republish (THE ONLY CURRENT INSTRUCTION; bound developer run)

Revision 4 was ACCEPTED on content (pin dcc7f015 kept, 5p8b0z gap row dropped, CHANGELOG = trunk). Integration refused
`integration_base_moved`: trunk advanced on .github/ci/gate-selftest.sh (BUG-3v8k23 pinned every lane's timeout expression;
BUG-306v4m rustup probe) — both also touched by this Story. Content is not in question; combine and republish.
1. `task-board m 'set_status(TASK-260922-18ex37, status=development)'` if needed.
2. Combine trunk: `git diff c278af4f ab34556e -- . ':!.task-board' ':!CHANGELOG.md' | git apply --3way` in the Story worktree. On
   .github/ci/gate-selftest.sh (and any other overlap) keep BOTH sides: 3v8k23's per-lane timeout pins and 306v4m's rustup rows plus this
   Story's gap-ledger / pin rows. Report each conflicted path and its resolution. Leave nothing staged.
3. `task-board worktree refresh-candidate TASK-260922-18ex37` (the 3bbvrs checkpoint replay conflicts only via its
   `--replay-resolutions` template; CHANGELOG.md → trunk bytes). CHANGELOG.md must equal trunk's.
4. Bounded re-runs: `bash .github/ci/gate-selftest.sh`, ledger scripts, `go test ./internal/scriptpolicy ./internal/crossconformance
   -count=1`. Real exit codes.
5. Append "Revision 5 — refresh onto ab34556e" to results, `resource update`, `task-board handoff TASK-260922-18ex37 --role developer`; stay
   in the turn while the gate runs. A `run_wrote_outside_worktree … policy warn` block is a warning.
