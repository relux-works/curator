# Review note — TASK-260910-2n0233 rev4 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Re-review after rework-1 (`2n0233-rework-1.md`). Rev4: base d41da0fb (= trunk), tree c05b7e41, 20 paths (same path set as rev3), gate green.
Your rev3 verdict F1: mutant M2 (boundary.go openPageChain treats an unreadable high-water state as absent) survived. Verify:
1. New FetchFn rows (persisting AND read-only variants) with an unreadable state file (chmod 000 and directory-in-place) and corrupt JSON:
   fail closed with pageStateError, no records, state bytes unchanged, record cache not written. Re-run M2 → must be KILLED now; re-check M1/M3/M4.
2. Catalog-lists-it-but-missing branch row or a stated bound.
3. Windows: chmod rows skip with a reason; the directory-in-place row runs everywhere (check the ledger/platform-cases).
4. Rebase fidelity: on paths not touched by trunk between 0ffe2e1d and d41da0fb, rev4 == rev3 plus only the rework; no trunk revert.
Bounded runs. accept_cr or changes requested with file:line. No LOGBOOK.md.
