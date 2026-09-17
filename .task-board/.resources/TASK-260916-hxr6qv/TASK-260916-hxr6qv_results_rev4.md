# TASK-260916-hxr6qv rev4 results — rework-3 gate fix (test only, story_final)

## Scope
Rework-3: ubuntu gate failure (run 35197231529) in
TestProductionExternalDepsFalseDrivesMirrorFetchToSink. The real install
reaches the external build step and refuses with
go-v1 build_execution_control_unavailable because
rc5-native-control-inventory-v1 covers macOS and Windows only.
Fix is TEST ONLY in cmd/curator/draft_transport_provenance_test.go.
Product code is byte-identical to rev3 (verified by sha256 before/after
the mutant overlays; no MUTANT markers remain; git diff stat unchanged).

## Change
Keep productionExternalDeps(cfg, false) composition. Branch only the
build-completion expectation on godriver.InventoryPlatform(runtime.GOOS):
- Covered host (darwin/windows): require exitOK, 6 fetches
  (plan primary/mirror/ssh + stage primary/mirror/ssh), 6 sink lines,
  2 mirror records. Assertions identical to rev3.
- Uncovered host (linux): require the inventory refusal on stderr
  (build_execution_control_unavailable, rc5-native-control-inventory-v1,
  no record for host linux), then run the SAME sink assertions with
  acquisition-complete counts: 5 fetches (plan 3 + staging partial
  primary/mirror; the second staging fetch never runs because the first
  staging build refuses after its acquisition), 5 sink lines, 2 mirror
  records (plan + staging acquisitions both record before the refusal).
  Rationale: probeNativeControls fires in godriver Build, after
  planExternalBuilds Probe and stageExternalBuilds Establish plus the
  staging acquisition, so provenance lands independent of build outcome.
Sink provenance content checks (canonical identity, listed/resolved
mirror, mirror_of, allowlisted shape, no broker secret) and portable
artifact checks (stdout/stderr, cache receipts, markers, Skillfile.json
carry no endpoint provenance and no secret) run unconditionally on every
host. Both sink mutants fail at the pre-install depsCheck on every lane.

## Evidence (darwin arm64, shell sh, set -o pipefail style EXIT via PIPESTATUS)
- gofmt -l cmd internal: clean, exit 0
- go vet ./cmd/curator/: clean, exit 0
- golangci-lint run cmd/curator/... internal/install/... internal/buildrepo/...: 0 issues, exit 0
- go test -p 1 -count=1 -run TestProductionExternalDeps ./cmd/curator/: ok 76.075s, exit 0
- go test -p 1 -count=1 -run TestDraftTransport ./cmd/curator/: ok 64.488s, exit 0
- go test -p 1 -count=1 -run <draft|transport|resolution|attempt|provenance mask> ./internal/install/ ./internal/buildrepo/: ok both, exit 0
- revision-2 unit mask (TestValidateTransportPlan*, TestPortAliasRefusalNarrows, TestAcquireDraftNetworkRevision2*, etc.): ok both packages, exit 0
- Mutant A overlay (removed sink, trace=nil): target test FAIL in 1.5s, exit 1, at the productionExternalDeps(cfg,false) nil-trace check
- Mutant B overlay (dry-run-only sink): target test FAIL in 1.7s, exit 1, same check
- Post-restore target test re-run: ok 64.480s, exit 0; main.go sha256 matches pre-mutant value
- Linux/Windows lanes: not runnable from this host; reported as unverified here, covered by the hosted gate on handoff. The uncovered branch reuses the exact refusal vocabulary asserted by TestCompiledInstallFollowsTheNativeControlInventoryExactly.

## Checklist
- [x] Rework-3 implemented test-only; product unchanged
- [x] Narrow tests green on this host; remote gate left to handoff landing suite
- [x] Both sink mutants killed with recorded exit codes (fail on every lane by construction: pre-install check)
- [x] Lint clean (gofmt, vet, golangci-lint on touched packages)
- [x] Legacy/rev1 goldens untouched; closed grammar untouched
- [x] Workspace left UNCOMMITTED: 27cv45 checkpoint plus this leaf delta (6 modified product files from revs 1-3, 4 test files incl. this rev)
- [x] story_final: a base_authority_mismatch refusal at handoff is orchestrator-handled; do not retry
