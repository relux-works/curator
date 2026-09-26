# Review note — TASK-260925-h4syhu (split leaf; orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

h4syhu owns ONE surface split from TASK-260907-2as5sx by the rev9 loop-detector response: lockedNetworkRepository read-failure rows
(internal/install/draftsources_test.go, and a ledger row if needed); it also publishes the Story candidate (STORY-260906-1a2i5a).
Candidate: 73 paths on base ab34556e (2as5sx rev9 had 74). Verify:
1. F1/M1 from TASK-260907-2as5sx_review-verdict-rev9.md: rows absent / Lstat failure (typed manager_state_unreadable, no fallback) /
   present-but-unusable as separate subtests; M1 (`continue` first in the Lstat error branch, draftsources.go ~228) survives before and is
   KILLED now — reproduce it yourself on darwin. Windows: a real Lstat-failure input or a declared ledger row with the exact reason; the
   ERROR_PATH_NOT_FOUND-under-a-file statement is present and correct.
2. Every other path equals 2as5sx revision 9 (per-file patch-id) EXCEPT (a) trunk content combined by the refresh onto ab34556e (both
   sides present) and (b) migrations of trunk-added collapse sites the deny-by-default guard caught — list each and judge it; the guard's
   counts/ratio updated honestly. Name the one path that dropped vs rev9 and why.
3. No revert of trunk, no CHANGELOG/LOGBOOK edit, no stray files; hosted gate green on all lanes (the successor's Windows fix — judge it).
Focused tests only (host memory). accept_cr or changes requested with file:line. No LOGBOOK.md.
