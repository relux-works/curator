# Review note — TASK-260917-2tx81l rev3 re-review (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Re-review rev3 (base 30b3d678, tree 92a072c4, 45 paths, gate green) against your rev2 verdict (F1, F2) and `2tx81l-rework-2.md`.
This is a security path, so treat it as an important review. Verify:
1. F1:
   - marker.Write keeps the base (30b3d678) semantics when hashing.EnableV2Writers is off;
   - recomputation happens only under v2, documented in a comment;
   - TestAuthoritativeCompiledMarkerRoundTripsThroughWriter and the other rc.13-mode writer tests in marker_v2_test.go are
     BYTE-IDENTICAL to base and run with the switch OFF;
   - v2-mode expectations live in NEW tests.

   Prove it: check out the base versions of every existing test file modified over the candidate code, and run them with
   CURATOR_CONFORMANCE_ROOT=rc.13. They must pass, except the four explained reader-version edits.
2. F2:
   - the results explain the four reader-version edits per test (config 203, environments 152, envmarker 58/72);
   - the envprofile credential/migrate tests that opt into v2 have unchanged rc.13-mode originals beside the new v2 tests.
3. The three mutants are still killed: re-run them with real exit codes. Framing and the ledger are unchanged from rev2: spot-check.
accept_cr, or changes requested with file:line. Never spell any employer name.
