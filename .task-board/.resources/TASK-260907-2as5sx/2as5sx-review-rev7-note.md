# Review note — TASK-260907-2as5sx revision 7 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Revision 5 was CHANGES_REQUESTED only for F1 (named-list guard). Revision 6 made the guard deny-by-default; revision 7 = rework 5
(macOS `TestAnInFlightTransactionKeepsThePublishedCacheEntry` investigation, CHANGELOG hunk reverted per policy, refresh onto trunk 948ae7c9).
Verify in a disposable clone:
1. F1 fixed: the guard scans all non-test .go under internal/ and cmd/ (outside internal/stateread) and fails unless routed via the seam or
   on the reviewed allowlist; ratio line present; re-apply the rev5 reviewer's mutant (`readNewManagerState` with IsNotExist→default) → the
   named test fails.
2. The macOS test: read the producer's base-vs-candidate `-count=30` evidence; a cause in the seam must be fixed in production, a pre-existing
   flake must be shown on the base; the test itself untouched.
3. Refresh: readers R5 added are migrated or allowlisted with reasons; no CHANGELOG.md in the patch; entry text in results.
4. Items 3-4 of the rev5 note (regression fixes, Windows ledger rows) still hold; no stray files; validation log green.
accept_cr on revision 7, or changes requested with file:line. No LOGBOOK.md.
