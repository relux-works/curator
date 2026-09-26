# Review note — TASK-260923-em42lw launch-env-fragment-v2 permissions (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review the newest revision (rev9, base 316438cc, gate green) against `em42lw-brief.md` (scope) and the reworks `em42lw-rework-3.md`,
`em42lw-gatefix-4/5/6.md`. Verify:
1. `curator env resolve --format json` emits launch-env-fragment-v2 with the REQUIRED closed `permissions` {mode, locked, source} per the
   lattice (locked iff source=global; mode=native whenever source is global or default; profile source from permissions.<profile>; §12.2
   system-file force-native lock) — rows through the CLI; kill one lattice mutant yourself.
2. The pre-0018 vector accommodation is removed now that SPEC_PIN carries 0018 (or its retention is justified).
3. Gap ledger owners re-derived after `permissions` stopped blocking (overlay rows now owned by source_signers/ioemse only); ratchet honest.
4. v1 fragments still read/validated (the v1 invalid cases rejected; coverage counts real).
5. Windows: the v2 schema's POSIX-only env path pattern — check the chosen precedent/skip and its declared reason; the permissions lattice
   is still asserted on Windows.
6. Merge with 2elcdc correct (isolation admits isolated; no stale refusal rows); no revert of trunk; no CHANGELOG/LOGBOOK; no stray files.
Focused bounded runs (host memory). accept_cr or changes requested with file:line. No LOGBOOK.md.
