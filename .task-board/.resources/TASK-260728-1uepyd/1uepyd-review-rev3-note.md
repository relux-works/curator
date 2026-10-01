# Review note — TASK-260728-1uepyd rev3 re-review (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Re-review rev3 (base 5432c85f, tree 5a1f7895, 5 paths, gate green) against your rev2 verdict (F1) and `1uepyd-rework-1.md`. Verify:
1. The cases test compares the FULL logged fetch argv (with operation-private paths resolved under one private root) and the FULL
   environment Git received, against the vector plus the transport additions and the platform allowlist.
2. Re-run M4 (extra `-c fetch.prune=true` at the call site) and M5 (`GIT_SSL_NO_VERIFY=1` leak at the call site) yourself: both must be
   killed, with real exit codes. M1–M3 must still be killed.
3. The residuals are handled: R-a fixed or removed; R-b driven through RunPipeline or stated as a bound.
4. No production change beyond rev2's reorder; no CHANGELOG/LOGBOOK; no stray files.
accept_cr, or changes requested with file:line. Never spell any employer name.
