# Review note — TASK-260925-h4syhu revision 3 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Revision 2 was ACCEPTED. Revision 3 = refresh onto 9f0da708 (`h4syhu-refresh-2.md`): trunk added 18ex37 (ledger/pins), 11burj
(internal/install/draftsources.go declaredDependencyReplaySources), 10d3l1, 11jgkt. Orchestrator pre-check: same 73 paths as rev2, no
stray files, gate green. Verify: (1) every path equals rev2's content except where trunk changed the same file — there both sides present;
(2) the deny-by-default guard now also covers 11burj's new code: any new collapse site migrated onto the stateread seam or allow-listed
with a reason, counts/ratio honest; (3) M1 still killed; (4) validation log green on all lanes. Focused checks only (host memory).
accept_cr or changes requested with file:line. No LOGBOOK.md.
