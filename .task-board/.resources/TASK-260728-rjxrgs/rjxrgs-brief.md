# TASK-260728-rjxrgs — drive install-marker-v3 schema cases and the lifecycle mixed/shim/signing/transaction cases (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md`, this task's README, and the audit `.research/260930_compiled-build-leaves-reconciliation.md` (row
rjxrgs). The production code is on main:
- internal/install/install.go:704,749 and global.go:299,339 (planExternalBuilds / stageExternalBuilds);
- external.go:517 externalMarkerBuild;
- marker.go ExternalSchemaVersion=3.

Gaps against curator-spec rc.13 (SPEC_PIN):
- the 27 conformance/v1/schema-cases/install-marker-v3 cases are not consumed, and there is no row in
  .github/ci/conformance-case-counts.tsv;
- lifecycle vector tables mixed_build_cases (6), path_shim_cases (3), signing_cases (4) and transaction_cases (4) are not consumed;
- status_repair_gc_cases is 1/5 driven, with 4 bound;
- the expected/external-repository/{build-receipt-v2, install-marker-v3-mixed, mixed-build-plan}.json bytes are not read.
  TestExternalReceiptV2CacheKeyVector pins a hard-coded key.

Steps:
1. Drive all 27 install-marker-v3 schema cases through the production marker reader, and add the exact count row.
2. Drive mixed/path_shim/signing/transaction through the production install entry. Lift the 4 status-repair-gc bounds if the manager
   surface now allows it; otherwise restate each reason precisely.
3. Replace the hard-coded receipt-v2 key with the published expected bytes, and compare the mixed plan and marker against the
   expected files.
4. Mismatches: fix curator bugs in production code; if the vector is wrong, report it. Anything left is an owned known gap with a
   reason; nothing skips silently. Coordinate the marker-v4 known gaps with BUG-260923-2afgyq only by reference: do not fix them here.
5. Mutants, with real exit codes, each killed:
   - receipt interpretation aliasing allowed;
   - a mixed-plan order change;
   - a transaction step skipped.
6. No Windows-reserved names. No CHANGELOG/LOGBOOK: put the entry text in the results. Never spell any employer name.
Update the results with the per-table coverage, then run `task-board handoff TASK-260728-rjxrgs --role developer`, then END YOUR TURN.
