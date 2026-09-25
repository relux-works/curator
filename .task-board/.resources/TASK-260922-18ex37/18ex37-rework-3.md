# TASK-260922-18ex37 — rework 3: keep pin dcc7f015 + drop the 5p8b0z gap row + refresh (THE ONLY CURRENT INSTRUCTION)

Unblocks curator-spec PR #88 (its "Implementations" check fails only `internal/scriptpolicy
TestScriptHostExecutionPolicySectionsAreAllClassified` until curator main classifies #88's sections) and so rc.13 → curator rc.2.
OPERATOR DECISION 2026-09-25: manifest v9 (#89, TASK-260924-2am4qa) stays OUT of the accepted skillfile-sources revision — do NOT pin
5470c04b.
BUG-260924-5p8b0z LANDED on trunk `c278af4f` (Windows System32 hard-link trust now requires a manager-captured SystemRoot).
1. `task-board m 'set_status(TASK-260922-18ex37, status=development)'` if needed.
2. Pin: KEEP every spec pin at curator-spec `dcc7f015` (= #88 erratum, already classified in revision 3). Do not vendor #89 (v9) or
   #90 content. Confirm the bidirectional classification tests pass at dcc7f015.
3. Gap ledger: REMOVE the `windows-exec-uncaptured-systemroot-hardlinks` row owned by BUG-260924-5p8b0z (the case now passes; the
   ratchet rejects passing gaps). Re-derive any other row whose owner is now done.
4. CHANGELOG POLICY: CHANGELOG.md must equal trunk's bytes (this also drops the checkpointed 3bbvrs entry); both entries (3bbvrs, 18ex37)
   verbatim in results under "## CHANGELOG entry (for release prep)". No root TASK-*/BUG-* files (the working tree already deletes
   TASK-260922-3bbvrs_results.md — keep that deletion), no .temp/ or test/ or ledger/ paths in the candidate.
5. Refresh onto trunk `c278af4f`: `git diff 948ae7c9 c278af4f -- . ':!.task-board' ':!CHANGELOG.md' | git apply --3way` (keep both sides), leave
   nothing staged, `task-board worktree refresh-candidate TASK-260922-18ex37` (checkpoint replay conflicts only via its
   `--replay-resolutions` template; CHANGELOG → trunk bytes).
6. Bounded runs: the conformance packages you touch, `internal/scriptpolicy`, `internal/scriptworker` executable-identity rows,
   gate-selftest, ledger scripts. Append "Revision 4 — drop 5p8b0z gap row + refresh", `resource update`, `task-board handoff TASK-260922-18ex37
   --role developer`. A `run_wrote_outside_worktree … policy warn` block is a warning.
