# Review note — TASK-260910-32gki6 S5 profile-store protected boundary (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev5 (base 213a53e5, tree 0fd98b72, 31 paths, gate green on every lane) against `32gki6-sec-brief.md`, `32gki6-decision-1.md`
(pin tree hash from the LOCAL object DB; missing commit object → environment_store_untrusted; no fetch; lock not versioned) and
gatefix notes 1–4. Verify against curator-spec rc.13 environments.md §4 "protected state" and §9 resolve steps:
1. env resolve checks the environments root, profile store root, lock file, markers and every lock-named store entry with all five
   boundary checks (ownership; private permissions / owner-only DACL on Windows; containment; regular file types; lstat link safety) plus
   pin-hash recomputation, through internal/pathboundary (reuse, not a second walker) and internal/stateread.
2. Failure classes: enclosing boundary → refuse the whole operation; entry → environment_store_untrusted + rebuild from a revalidated
   snapshot; unreadable lock never rebuilt. Missing hash never counts as passed.
3. Rows at the production entry: tampered entry → untrusted; missing commit object → untrusted; match → passes; each boundary check has a
   failing row. Fixtures that previously used fake pins now use real local commits and keep their original assertions (no assertion weakened).
4. Windows fixes (gatefix-4) did not relax the DACL check; any Windows skip is registered in skip-classes.tsv + platform-cases.tsv with an
   exact reason. conformance-gaps.tsv: rows now passing left the ledger; before/after counts reported.
5. Mutants — run them yourself with real exit codes: pin check skipped; missing object treated as pass; one boundary check removed;
   entry failure escalated/demoted between classes. Each must be killed.
6. No CHANGELOG/LOGBOOK; no stray files.
accept_cr or changes requested with file:line. No LOGBOOK.md.
