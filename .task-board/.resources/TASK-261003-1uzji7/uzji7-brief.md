# THE ONLY CURRENT INSTRUCTION — TASK-261003-1uzji7: atomic v1→v2 identity migration, then the v2 writer flip (developer, code; rc.4)

Operator decision 2026-10-03: rc.3 ships without the v2 writer; rc.4 carries this. Read the attached 1foyf3-blocker-evidence.md: flipping EnableV2Writers alone made ApplyMigration relabel v1 identities as v2 (aliasing old and new identities).
1. Design first, in a short .research note in your CR: every place WriteVersion() feeds (install targets/publication, markers, context resolution/materialization/store/locks, environment marker/migration publication); the atomic migration unit; rollback; how readers accept both versions during migration; how an old identity can never be silently relabelled.
2. Implement: the atomic v1→v2 profile/identity migration (rehash, never relabel), then flip internal/hashing EnableV2Writers to true. Bring back the TestRC14MigrationRehashesLegacyIdentities regression from the evidence. Exercise real entry points plus legacy-reader, NUL and mismatched-version negatives.
3. Conformance: remove the rc.14 known-gap row snapshot-acquisition/cases/byte-exact-snapshot (owned by this task) only if that exact production case now passes; keep exact counts.
4. This lands AFTER the curator v0.15.0-rc.3 tag.

## Produce mode (operator 2026-10-04, binding)
Maximise useful output now. Write the code and targeted tests for the packages you touch, and run ONLY those targeted tests (GOFLAGS=-work). Do NOT run the full local suite: the host has exec stalls, and the hosted gate on the Change Request is the arbiter. State plainly which claims are verified locally and which are left to the hosted gate. Never edit LOGBOOK.md or CHANGELOG.md. If a command hangs for more than 5 minutes, wait instead of retrying in a loop.
Then `task-board handoff TASK-261003-1uzji7 --role developer` and END YOUR TURN.
