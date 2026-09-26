# Review note — TASK-260925-h4syhu revision 7 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Revision 3 was ACCEPTED; since then the Story was refreshed onto trunk 316438cc, which now includes cww1ov (0017; its own stateread seam and
reader migrations), 1f2ng0, 2elcdc, 20o9dk, 2gt5f6. Revisions 4–6 fixed the merge; revision 7 = platform-correct assertion + CHANGELOG
revert. Candidate: 69 paths, gate green on all lanes. Verify:
1. ONE stateread seam: cww1ov's and this Story's APIs reconciled (no duplicate helpers, no divergent semantics); every reader migrated once.
2. lockedNetworkRepository returns the typed manager_state_unreadable for a failed checkout Lstat (environments §8.4.1), never
   source_snapshot_unavailable (operator addendum: that code is only for an unreachable source); cww1ov's other wrapping unchanged.
3. The test assertion is platform-correct but still checks the underlying cause (lstat / GetFileAttributesEx); M1 killed (reproduce).
4. The deny-by-default guard scans the merged code; counts/ratio honest; new trunk collapse sites migrated or allow-listed with reasons.
5. Every other path equals revision 3 content except trunk combination (both sides present); CHANGELOG.md = trunk; no LOGBOOK; no stray files.
Focused bounded runs (host memory). accept_cr or changes requested with file:line. No LOGBOOK.md.
