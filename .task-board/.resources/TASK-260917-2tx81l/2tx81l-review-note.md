# Review note — TASK-260917-2tx81l curator content-hash framing v2 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev2 (base 30b3d678, tree d9d1b2e8, 49 paths, gate green on every lane; rev1 broke 35 tests because writers switched to v2)
against `2tx81l-brief.md`, `2tx81l-decision-1.md`, `2tx81l-gatefix-1.md` and curator-spec b1a2efb (PR #116). Verify:
1. Framing: internal/hashing v2 equals the spec. Recompute content-hashes-v2.json (colliding pair, empty tree, exact hex) through the
   curator implementation. v1 stays byte-identical, and v1 identities never equal v2 identities.
2. Writers still emit exactly the rc.13 shapes (SPEC_PIN is rc.13). The v2 write cut-over sits behind one explicit switch that is OFF
   by default and has a TODO for rc.14. Readers accept and validate the v2 shapes. Registry matching compares versions
   (internal/registry). The 35 rev1-broken tests pass UNCHANGED: diff the test files rev1-base to rev2 and flag any weakened assertion.
3. Ledger:
   - the 80 core-v2 rows are driven and gone from conformance-gaps.tsv;
   - the 7 skillfile-sources rows are re-owned to TASK-260930-3ny11n or removed, per decision-1 (confirm which, and that it is
     correct for the b1a2efb suite digest);
   - exact counts hold for both the rc.13 and the b1a2efb digests.
4. Re-run the three mutants yourself with real exit codes:
   - length framing removed;
   - version not compared in registry matching;
   - v1 marker accepted for a v2 expectation.

   Each must be killed.
5. Reads go through stateread; managed writes use the nofollow helpers. No Windows-reserved names; no CHANGELOG/LOGBOOK; no stray
   files.
accept_cr, or changes requested with file:line. Never spell any employer name.
