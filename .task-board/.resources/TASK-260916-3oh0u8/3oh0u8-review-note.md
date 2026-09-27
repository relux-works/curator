# Review note — TASK-260916-3oh0u8 E4 umbrella provider trust roots (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev2 (base d41da0fb = trunk at spawn, tree d13d9519, 6 paths, gate green) against `3oh0u8-sec-brief.md`, `3oh0u8-restart-1.md` and
curator-spec v1.0.0-rc.13 provider-resolution trust roots (landed spec side TASK-260916-1x0ogh; cite clauses). The old rev1 was 15 files on
an ancient base — rev2 is a re-implementation; check that nothing the spec requires from rev1's scope was silently dropped (compare with
`git diff refs/campaign/2otjbn-full-20260927^ refs/campaign/2otjbn-full-20260927` and the rev1 review verdict).
Verify through the production entry: umbrella provider lookup only inside declared trust roots; outside → refused fail-closed with the spec
diagnostic (PATH/ambient lookups, symlink escapes out of a root); env status/launcher prints the resolved provider path; rc.13 vectors driven;
Story-owned gap rows removed; at least one mutant per rule killed (real exit codes); no trunk revert, no CHANGELOG/LOGBOOK, no stray files;
new manager-state reads via internal/stateread. Bounded runs. accept_cr or changes requested with file:line. No LOGBOOK.md.
