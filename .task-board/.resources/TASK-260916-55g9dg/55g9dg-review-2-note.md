# Review note — TASK-260916-55g9dg E2 direct-only system modules (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review the newest revision (rev5, re-applied onto trunk 97ca3370, gate green, 8 files) against `55g9dg-sec-brief.md` and curator-spec
v1.0.0-rc.13 (cite clauses: class: system modules are direct-only; a transitive dependency resolving to a system module is refused).
Verify through the production entry (CLI install/resolve):
1. A transitive system module is refused fail-closed with the spec error/exit; a direct system module still installs; no partial writes.
2. The rc.13 vectors for this surface are driven; E2-owned conformance-gaps.tsv rows removed (before/after counts).
3. Kill at least one mutant (check removed / refusal softened to warning) with real exit codes.
4. The candidate diff touches ONLY the E2 paths (no revert of trunk: `git diff --name-only <base> <candidate>` vs the patch), no
   CHANGELOG/LOGBOOK, no stray files. New manager-state reads go through internal/stateread (guard test).
Focused bounded runs. accept_cr or changes requested with file:line. No LOGBOOK.md.
