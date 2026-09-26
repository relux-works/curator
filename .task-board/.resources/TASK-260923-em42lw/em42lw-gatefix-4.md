# TASK-260923-em42lw — gate fix (THE ONLY CURRENT INSTRUCTION)

Revisions 4 AND 5 (identical tree 038e53d3) fail the hosted gate on every lane with the SAME two real failures — the successor republished
without fixing them:
1. internal/config TestOverlayGapOwnersMatchFirstProductionBlocker (environments_conformance_test.go:380):
   "schema2-overlay-git-https-uppercase differences = [environments.require_source_signers environments.source_signers], want both permissions
   and source_signers fields". Your change implements decision-0018 `permissions`, so the overlay cases' remaining production blocker is
   ONLY source_signers (owner STORY-260916-ioemse). Re-derive every overlay gap row in .github/ci/conformance-gaps.tsv and the test's expected
   blocker sets from the actual config.Load error at the pinned root (dcc7f015): drop `permissions` as a blocker where it no longer blocks,
   fix owner + reason, and report the owner histogram before/after (the ledger ratchet rejects passing gaps — any case that now passes
   fully must leave the ledger).
2. internal/envfragment TestFragmentAuthoritativeSchemaCases: "published cases launch-env-fragment-v1/schema-cases: 0 driven … 0 total" and
   `invalid-fragment-identity.json` ACCEPTED (fragment_schema_test.go:319). v2 emission must not break v1: the v1 schema cases must be found
   (check the case path/root resolution you changed) and every v1 invalid case must still be rejected by the v1 reader. Fix the regression
   in production/test code; do not ledger it.
3. Run locally, bounded and split: `go test ./internal/config -run 'TestOverlayGapOwnersMatchFirstProductionBlocker|TestManagerConfigV2Vectors' -count=1`,
   `go test ./internal/envfragment -count=1` (+ -race on both), ledger scripts. Real exit codes.
4. `task-board m 'set_status(TASK-260923-em42lw, status=development)'` first; append "Revision 6 — gate fix", `resource update`,
   `task-board handoff TASK-260923-em42lw --role developer`; stay in the turn while the gate runs. If the loop detector refuses, stop and report.
No CHANGELOG/LOGBOOK edit. A write-boundary `policy warn` block is a warning.
