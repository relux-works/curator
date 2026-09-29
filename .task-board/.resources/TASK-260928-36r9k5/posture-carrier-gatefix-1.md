# TASK-260928-36r9k5 — gate fix (THE ONLY CURRENT INSTRUCTION, with posture-carrier-brief.md)

Rev1 (tree 0bea8b4b, base 213a53e5) fails 6 tests (run 36421064199):
1. internal/config TestManagerConfigV2SchemaCases / TestSystemConfigV2SchemaCases: "known-gap case …/valid-security-posture-hardened.json
   (and system-config-v2 …/valid-security-posture-hardened-locked.json) now passes; remove its ledger row" — remove those rows from
   .github/ci/conformance-gaps.tsv (your content now drives them).
2. cmd/curator TestGlobalAdoptCLIAdoptsAndIsIdempotent, TestGlobalAddPublishesProfileLockBeforeNativeMaterialization,
   TestProfileUpdateSystemDeltaConfirmationGolden, TestProfileUpdateMCPDeltaConfirmationGolden — tests that landed on trunk AFTER 4pv4au
   was written (global adopt 2s0jsc, lock ordering 3ed9m3, E1 delta goldens) assert exact stderr/goldens and now see the spec-mandated
   `warning: security_posture_permissive: …` (profiles/manager.md §7.1: every operation under permissive warns once). Do NOT suppress or
   move the warning. Update those tests to the specified behaviour: either expect the warning line exactly once in stderr (golden files
   regenerated with it), or give the fixture machine config an explicit posture when the test is about something else — choose per test
   and say why; keep at least one assertion per command that the warning appears exactly once under permissive.
Run `go test ./internal/config` and `go test ./cmd/curator -run 'GlobalAdopt|GlobalAdd|GlobalInstall|ProfileUpdate|SecurityPosture|Golden'`
with real exit codes. `task-board m 'set_status(TASK-260928-36r9k5, status=development)'` first; update results; handoff; END YOUR TURN.
No CHANGELOG/LOGBOOK edit.
