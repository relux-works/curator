# TASK-260910-2n0233 — rework 1 (THE ONLY CURRENT INSTRUCTION, with 2n0233-sec-brief.md)

Review rev3 = CHANGES REQUESTED (`TASK-260910-2n0233_review-verdict-rev3.md`, finding F1). Mutant M2 (boundary.go:66-68 — unreadable
high-water `snapshot-<digest>.json` treated as absent in openPageChain) SURVIVES: no row drives an unreadable/corrupt state file through the
FetchFn. Do exactly the verdict's "Required rework":
1. internal/registry row(s) that build the FetchFn (NewHTTPFetch / NewHTTPFetchWithPolicy persisting AND the read-only variant) against an
   httptest registry, with the state file unreadable (chmod 000 / directory in place of file) and with corrupt JSON: assert the pageStateError
   (fail closed), no records served, state file bytes unchanged, record cache not written. Show M2 killed (real exit codes).
2. If cheap, the same through a production caller (install resolve or `status --attest`).
3. A fetch-path row for the catalog-lists-it-but-missing branch (boundary.go:69), or state the bound.
Also: trunk moved to 38c68570 — `git fetch origin main`, combine keeping both sides; VERIFY `git diff --name-only origin/main -- . ':!.task-board'`
shows only this leaf's paths (no trunk revert). `task-board m 'set_status(TASK-260910-2n0233, status=development)'` first; bounded runs;
append "Revision 4 — unreadable high-water through the fetch path", `resource update`, handoff; stay in the turn. Unix-only chmod rows must
skip on Windows with a reason (and the directory-in-place variant should run everywhere). No CHANGELOG/LOGBOOK edit.
