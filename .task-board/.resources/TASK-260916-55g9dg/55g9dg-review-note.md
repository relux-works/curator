# Review note — TASK-260916-55g9dg E2 direct-only system modules (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review the newest revision (on current trunk, 8 paths, gate green) against `55g9dg-sec-brief.md` and `55g9dg-rework-fresh.md`. Normative
text: curator-spec v1.0.0-rc.13 — class: system modules only from packages named directly by the root or by the machine-config allowlist;
context_system_module_transitive + its scoped waiver; context-system-module-present stays always-warn. Verify through the production entry:
direct admitted, transitive refused with context_system_module_transitive (package + module named), waived transitive admitted with the
waiver scope honoured, warning kept. All rc.13 E2 vectors driven; gap-ledger honest (69 rows, histogram; any E2∩E4 case attributed to
STORY-260916-2otjbn with the exact E4 blocker). Kill at least two mutants (transitive admitted / waiver ignores scope). No revert of trunk,
no CHANGELOG/LOGBOOK. accept_cr or changes requested with file:line. No LOGBOOK.md.
