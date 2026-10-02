# Review note — TASK-260728-rjxrgs marker-v3 + lifecycle conformance (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev1 (base 5432c85f, tree ac008990, 17 paths, gate green) against `rjxrgs-brief.md`. The producer reported:
- marker v3: 22/27 driven;
- mixed 6/6, shim/PATH 3/3, transactions 4/4;
- signing 3/4, with release signing bound;
- status/repair/GC 1/5, with 4 bounds;
- three mutants killed;
- a production change in internal/skillspec/parse.go: package data cannot request artifact signing.

Verify:
1. Every install-marker-v3 case is driven or an OWNED known gap with a precise reason. For the 5 non-driven, say whether they are real
   curator bugs that must be fixed here. The exact count row is in conformance-case-counts.tsv.
2. The parse.go change matches rc.13 (cite the clause), and nothing that rc.13 permits is refused.
3. The receipt-v2 key uses the published expected bytes, and the mixed plan/marker are compared against the expected files.
4. The bounds are justified precisely: signing 1, status/repair/GC 4.
5. Re-run the three mutants yourself with real exit codes: aliasing allowed; mixed-plan order change; transaction step skipped.
6. TASK-260728-rjxrgs_results.md must NOT be committed at the worktree root; flag it if it is in the tree. No CHANGELOG/LOGBOOK; no
   Windows-reserved names.
accept_cr, or changes requested with file:line. Never spell any employer name.
