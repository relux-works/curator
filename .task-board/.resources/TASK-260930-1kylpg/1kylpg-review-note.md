# Review note — TASK-260930-1kylpg named-path absence rows (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev2 (base 30b3d678, tree 2554a2ec, 2 test paths, gate green on every lane incl. Windows) against `r1rows-brief.md` and
`1kylpg-gatefix-1.md`. Verify:
1. The pathboundary row asserts that a missing NAMED target on the production named route yields *Failure wrapping ErrNotExist (or the
   documented class). The Windows fixture root is created through the product's private-directory helper; there is no skip and no
   relaxed assertion.
2. The envprofile row removes a lock-named store entry and asserts environment_store_untrusted through the production resolve entry. The
   results must state what is observable at the boundary step versus at pin recomputation.
3. Re-run the mutant yourself (`return nil` on IsAbsent in the named-route Lstat loops, plus dropping `path == target ||` from vanished)
   and confirm it is killed. Give the real exit code.
4. Tests only: no production change, no CHANGELOG/LOGBOOK.
accept_cr, or changes requested with file:line. Never spell any employer name.
