# TASK-260917-2tx81l — rework 2 (THE ONLY CURRENT INSTRUCTION, with the earlier 2tx81l briefs)

The rev2 review (`TASK-260917-2tx81l_review-verdict-rev2.md`) verified the framing, the switch, all three mutants and the ledger. It
requested these changes. Read the verdict first.

F1 (blocking): an rc.13 writer changed, and a test was rewritten to hide it.
- internal/marker/marker.go:1231-1235 `Write` now always recomputes ContentSHA256 from dir, even with the switch OFF. Restore the base
  (30b3d678) Write semantics while hashing.EnableV2Writers is off, and recompute only under v2. Document the choice in a code comment.
- Restore TestAuthoritativeCompiledMarkerRoundTripsThroughWriter, and the other rc.13-mode writer assertions in marker_v2_test.go
  (TestWriteAlwaysProducesCanonicalMarkerV2, TestReadLegacyV1AndRewrite*), BYTE-IDENTICAL to base, running with the switch OFF.
- Put every v2-mode expectation in NEW tests with new names. Never mutate an rc.13 test into v2 mode.

F2: explain edited assertions per test in the results, and keep rc.13 coverage.
- config_test.go:203, environments_test.go:152 and envmarker_test.go:58/:72 (3→4) are legitimate reader changes (manager-config v3 and
  agent-environment-marker v3 are now known versions). Keep them, and write a one-line explanation per test in the results.
- For the envprofile tests that now opt into v2 writers (credential_link_test.go:411/448, credential_record_test.go:21/69/83/122,
  migrate_test.go:295/344): restore the original rc.13-mode tests unchanged. Add the v2-mode checks as NEW tests.

Verify it yourself before handoff:
- check out the base versions of every existing test file you modified over your candidate code, and run them with
  CURATOR_CONFORMANCE_ROOT=rc.13. They must pass, except the four explained reader-version edits;
- show that list in the results;
- keep the three mutants killed.

Set status development, update the results, run `task-board handoff TASK-260917-2tx81l --role developer`, then END YOUR TURN. No
CHANGELOG/LOGBOOK edit. Never spell any employer name.

## Continue (binding)
A previous run already started this rework in the worktree and ended without handing off. Inspect `git status` and `git diff` first and continue from there; do not start over. Finish F1 and F2, run the base-test verification, then set status development if needed, update the results, run `task-board handoff TASK-260917-2tx81l --role developer`, and END YOUR TURN.
