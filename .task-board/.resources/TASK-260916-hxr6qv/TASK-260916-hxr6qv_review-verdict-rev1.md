# TASK-260916-hxr6qv review verdict revision 1

CHANGES_REQUESTED. Route to to-dev; no acceptance or commit acknowledgement.

Candidate tree 932592c577e4cd5779e3d951c221956d90181867, base 62ea2d2ced3fac7c3c5f7a2ff4e887e0f92f4540. Tracked workspace bytes match the candidate; both untracked v2 test files hash to the candidate blobs. Predecessor results and accepted verdict read; its loader decisions are not reopened. No product code edits.

## Blocking finding

P1 — Required provenance is discarded in real CLI operations. internal/install/external.go:60-66 introduces a nil-default DraftTransportTrace and internal/install/drafttransport.go:101 forwards it. cmd/curator/main.go:1483-1497 productionExternalDeps never assigns a sink; repository-wide search finds the only non-nil assignment in internal/install/drafttransport_v2_test.go:40. internal/buildrepo/transport.go:898-903 emits only if trace != nil, otherwise records die with the call. Therefore a real opted-in mirror acquisition records no machine-private operation diagnostics, contrary to repository-transport section 7. The test manually supplies the missing production dependency and cannot detect this gap.

Required rework: connect sanitized records to a real machine-private operation diagnostics sink at the production composition boundary. Add an end-to-end test using production dependency construction, asserting canonical identity and listed/resolved mirror provenance are available there, no secrets appear, and portable artifacts omit those properties. A nil/remove-sink mutation must fail this test. Keep the strict port/alias refusals and frozen grammar unchanged.

## Independent evidence

Commands ran in zsh, direct exit codes (no pipelines):
- go test -p 1 ./internal/install -run 'Test(AcquireDraftNetworkRevision2|UserConfigIgnoredByResolvedLane|DraftTransportPlanV2|DraftPlanNeedsSSHRevision2)' -count=1 -timeout=90s: exit 0, 10.880s.
- go test -p 1 ./internal/buildrepo -run 'Test(ValidateTransportPlan|PortAlias|ParseTransportEndpointURL|ExhaustionV2)' -count=1 -timeout=90s: exit 0, 0.452s.
- go test -p 1 ./internal/config -run 'Test.*(V2|Revision2|SourcePolicy|ResolveRepository)' -count=1 -timeout=90s: exit 0, 1.331s.
- go vet ./internal/buildrepo ./internal/install: exit 0. git diff --check: exit 0.
- Combined cmd/curator TestDraftTransport(LegacyGolden|ResolvedMatrix|ProviderAdmission) run: exit 1, 90-second timeout during ProviderAdmission/distinct_providers_keep_their_own_material. Not claimed green.
- Separate TestDraftTransportLegacyGolden rerun terminated after about 98 seconds without results; not claimed green. Concurrent host execution was observed; cause not established.

Mutation attacks used temporary Go overlays, leaving candidate code unchanged. Narrowing executor cap from 2 to 1 was killed by the mirror fallback positive test (exit 1, explicit repository_policy_invalid assertion failure). Collapsing mirror failure class timed out (105s), not counted as killed. Collapsing alias failure class was terminated after 37s without assertions, not counted as killed. Measured: 1/3 selected mutants assertion-killed, 2/3 inconclusive, no claim of exhaustive mutation coverage. Overlay directory removed.

Production refusal table exercises 15/15 authored cases at acquireDraftNetwork, but this is below CLI dependency composition. Three-endpoint rejection remains helper-level in the new tests. Producer results contain no leaf-specific executed mutation evidence for attempt bounds and both new classes; complete and attach this evidence during rework.

Hosted validation reused, not replayed: TASK-260916-hxr6qv_change-request_rev1-validation.log reports exit 0, run https://github.com/relux-works/curator/actions/runs/35181995612. Independent gh query confirms success, head bf43ce900c88df636e08eb8f64ca6e49deb9d369; git rev-parse confirms its tree equals the exact candidate. 11 jobs passed per attached log; rose-air and Candidate suite skipped. ARM behavior unverified. Earlier producer run 35178895680 is not the handoff evidence used here.

Initial spawn goal query: not goal-bound. No directives at checked checkpoint. Findings recorded in this task-scoped outcome and board notes; LOGBOOK edits prohibited by campaign rules, no logbook CLI found. Reviewer does not close done or supply commit_ack.
