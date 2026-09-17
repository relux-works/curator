# TASK-260916-hxr6qv review verdict revision 2

CHANGES_REQUESTED — focused test rework. Candidate 448ac5f85acbcff88119cc9dcba72477e5633af7, base 62ea2d2ced3fac7c3c5f7a2ff4e887e0f92f4540. Independently compared all 26 changed file bytes against the exact candidate: 26/26 equal. No repository source changes or commits.

## P2 — requested production-composition end-to-end test is absent
cmd/curator/draft_transport_provenance_test.go:24 constructs productionExternalDeps(cfg, true), then lines 28–32 fabricate an AttemptRecord and call the callback directly. This proves callback assignment/serialization only. internal/install/drafttransport_provenance_test.go:65–69 manually constructs ExternalDeps and assigns the sink itself; it does not exercise productionExternalDeps. Lines 97–99 inspect a formatted snapshot, not emitted locks/receipts/markers/manifests. Thus neither test supplies the end-to-end proof explicitly required by hxr6qv-rework-1.md.

Independent narrowing attack: temporary Go overlay replaces main.go assignment with:
    if dryRun { deps.DraftTransportTrace = install.DraftTransportProvenanceTrace(cfg.Home()) }
This drops diagnostics for real non-dry-run operations. Both new tests remain green:
- go test -overlay=<temporary overlay> -p 1 ./cmd/curator -run '^TestProductionExternalDepsAssignsTransportProvenanceSink$' -count=1 -timeout=60s: exit 0, 0.630s.
- same overlay, ./internal/install -run '^TestAcquireDraftNetworkRevision2ProvenanceSink$': exit 0, 2.713s.
Measured selected narrowing mutations: 0/1 killed, 1/1 survived. This is a demonstrated coverage hole, not a claim that the unmutated candidate currently drops records.

Required next producer: add a bounded test driving an actual external operation using productionExternalDeps(cfg, false), with temporary machine policy and fake transport; preserve the constructed sink. Assert canonical identity and listed/resolved mirror provenance in the actual machine-private sink, no broker secret leakage, and no endpoint provenance in emitted portable artifacts. Kill both removed-sink and dry-run-only-sink mutants. Keep the strict refusals and closed grammar unchanged.

## Independent unmutated checks
zsh; direct go command exit codes, -p 1 -count=1:
- ./cmd/curator, ^TestProductionExternalDepsAssignsTransportProvenanceSink$: exit 0, 1.433s.
- ./internal/install, ^TestAcquireDraftNetworkRevision2: exit 0, 8.563s.
- ./internal/buildrepo, Test(ValidateTransportPlan|PortAliasRefusalNarrows|ParseTransportEndpointURL|ExhaustionV2): exit 0, 0.430s.
- ./internal/install, Test(DraftPolicySchema1Golden|Revision1ReaderRejectsV2Shapes|UserConfigIgnoredByResolvedLane|DefaultAcquireFetchesTheDeclaredURL): exit 0, 3.399s. Initial config invocation with this mask ran no tests; not counted as coverage.
- ./internal/config, ^TestSchema1GoldenUnchanged$: exit 0, 0.390s.
- git diff --check: exit 0.

Read spec sections 6–7, prerequisite loader results/verdict, revision-2 producer results and validation log. Production now assigns the sink at main.go:1503 and forwards it to AcquireNetworkResolved; no additional runtime defect established in this review.

Hosted evidence accepted from attached TASK-260916-hxr6qv_change-request_rev2-validation.log, not rerun: run 35192522852, exit 0, 11 successful jobs; rose-air and Candidate suite skipped. Independently verified local gate commit d6ad68e694d2336bc9ad1d567da9d9823b145b84 tree equals exact candidate. No live hosted-status query performed. Producer results cite a different earlier run; exact revision validation log is the evidence used here. Producer reports four killed mutants; these were not independently replayed and two failure-class examples are helper/loader-level rather than acquisition-level.

Goal query immediately before verdict: not goal-bound. No directives. Findings persisted here and in board notes; LOGBOOK.md writes prohibited by campaign rules. Route to to-dev, no accept_cr or commit_ack.
