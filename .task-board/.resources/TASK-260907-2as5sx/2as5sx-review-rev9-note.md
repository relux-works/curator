# Review note — TASK-260907-2as5sx revision 9 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Revision 7 was ACCEPTED (deny-by-default AST guard). Revision 8 = refresh onto a48f584c (stale on internal/marker/marker.go vs 1kpw4w);
revision 9 fixes the Windows-only failure of the new row `TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts/
unreadable_checkout_stops_fallback` (POSIX mode bits not honoured on Windows). Verify: (1) the refresh merged marker.go correctly (both
1kpw4w's directory changes and this task's changes present; guard counts still honest); (2) the Windows fix makes "unreadable" real on
Windows (no skip, no ledger row) or fixes a real production defect as claimed; (3) no CHANGELOG edit, no stray files; (4) hosted gate green
on all lanes. Focused tests only (host memory is tight). accept_cr or changes requested with file:line. No LOGBOOK.md.
