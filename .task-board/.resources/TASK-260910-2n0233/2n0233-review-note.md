# Review note — TASK-260910-2n0233 records boundary high-water check (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review the newest revision (rev3, gate green, base 0ffe2e1d, 20 files) against `2n0233-sec-brief.md`, `2n0233-gatefix-1.md` and
curator-spec v1.0.0-rc.13 (records boundary binding / client boundary high-water check; cite clauses). Verify through the production entry:
1. The client refuses a records boundary below its recorded high-water mark (rollback), fail-closed with the spec error/exit.
2. The high-water state read goes through internal/stateread: absent → first-use path; UNREADABLE → fail closed (not treated as absent);
   a row covers the unreadable case; TestManagerOwnedAbsenceReadsAreGuarded passes.
3. rc.13 vectors driven; Story-owned conformance-gaps.tsv rows removed or re-attributed with owners (before/after counts).
4. Kill at least two mutants (check removed; unreadable treated as absent) with real exit codes.
5. Candidate touches only this leaf's paths vs its base (no trunk revert), no CHANGELOG/LOGBOOK, no stray files.
Focused bounded runs. accept_cr or changes requested with file:line. No LOGBOOK.md.
