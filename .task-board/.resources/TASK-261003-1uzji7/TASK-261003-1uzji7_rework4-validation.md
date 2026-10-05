# TASK-261003-1uzji7 rework 4 — validation: v2 writer flip green (developer)

Base: 77fabd45 (converged). Rev3 hosted gate run 37240734108 was RED (46 ubuntu + 59 macOS failures).
This run fixes families A–D. Spec: curator-spec rc.14 (`daf15ec`) unless noted.

## Spec decisions (binding for the pins)

- **Family B verdict: the v2 flip DOES change the marker schema for every lane.**
  Protocol core §10 (rc.14): "Current writers MUST use marker schema 5 with
  framing version 2 for every installation mutation that writes a core marker."
  Markers 2/3/4 "retain their historical writer bands", "carry framing version 1,
  and their frozen schemas do not acquire a `hash_version` member." So per-lane
  schemas are frozen READS; every core write is now schema 5. Test pins expecting
  2/3/4 from the writer were updated to 5 with §10 citations. The v5 build-record
  rules equal v4/v3 (local go-v1 receipt 1 + execution_policy; external
  go-repository-v1 receipt 2) — "Marker v5 ... retains marker-v4 fields and
  build-record rules."
- **Family C verdict: env writers MUST use schema 3.** Environments §8.2 (rc.14):
  "Current writers MUST use schema 3 and include `hash_version: 2`." Schemas 1/2
  "have no `hash_version` and retain framing-version-1 meaning."
- Rehash-never-relabel holds: every legacy-shaped fixture below recomputes a v1
  digest; no v2 digest is ever presented under a v1 schema/version.

## Fix table

| Family | Root cause | Fix (file:line) | Failing tests it should turn green | Verified |
|---|---|---|---|---|
| A1 | Test helpers build v2-style build records (no execution_policy, receipt 0); core-v5 validation retains v4 record rules → `install marker is invalid for schema 5` at Write | `internal/scopes/gc_test.go:92` `buildRecordForBand` always emits v3-style records (receipt 1 + policy), matching install.go's writer upgrade | scopes: TestCollectMarksBuildKeysFromEveryBuildBearingMarkerSchema, TestCollectMarksBuildKeysFromEveryLiveScope, TestCollectSkipsTheBuildSweepOnUnprovableReferences/*, TestCollectStaysFailSafeOnRedirectedUnixScopes/*, TestCollectSweepsOnlyUnreferencedProtectedEntries, TestNonDirectoryScopeMembersDoNotBlockMaintenance, TestARepeatedRegistryMemberNeverEmptiesTheRegistry | **Local exit 0** (8 passed, 1 skipped — AuthoritativeGC needs conformance root, see below) |
| A1 | Same v2-style-record cause in CLI status fixtures | `cmd/curator/builds_test.go:966` `recordedBuildForBand` always emits v3-style records | TestStatusReportFindsASchema8InstallationCurrent/skill_schema_6, TestStatusReportMarksCompiledStateThatMovedDuringTheCheck | Hosted-pending (cmd/curator forbidden locally) |
| A1 | Band coverage: status/GC schema rows never admitted v5 | `internal/scopes/gc_test.go:192` +v5 absorb row; `cmd/curator/builds_test.go:1002` +v5 classify row (+`markerAtSchema` HashVersion); production already accepts v5 via BuildBearingSchema | (strengthening, no red test; guards the written schema) | Scopes row local exit 0; builds row hosted-pending |
| A2 | Substituted installs resolve via HEAD → marker (revision/HEAD/commit); core-v5 validation requires revision ref == commit (immutable binding) → Write fails | `internal/install/install.go:1220` `buildMarker`: under v2 writer, symbolic revision ref records the resolved commit; provenance stays in `Substituted` | TestDevSubstitutionAppearingBeforeHomeLockRestartsClosureResolution, TestProjectNonGitWithDevSubstitution/substituted, crossconformance TestDraftSourcesSemanticCasesBatch4/legacy-substitution-current | Hosted-pending (internal/install forbidden locally). No validation weakened; no test asserts Ref==HEAD |
| A3 | Registry fixtures serve rc.13 v1 records keyed by v1 hashes; `MatchesVersioned` refuses cross-version records → strict-policy unknown | `internal/install/registry_e2e_test.go:846` pin v1 lane (test's stated "frozen v1 lane" intent; serial, restores switch) | TestRegistrySnapshotSlowSiblingDoesNotFlipInstantRegistry | Hosted-pending |
| B | Writer now emits v5 for every lane per core §10; pins expect 2/3/4 | `cmd/curator/builds_test.go:1053` band pins → v5 (all three) | TestStatusReportFindsASchema8InstallationCurrent/skill_schema_7, /skill_schema_8 | Hosted-pending |
| B | Same §10 cause, install entry | `internal/install/external_lifecycle_conformance_test.go:142,205` mixed-build pins → v5; receipt assertions (local 1 / external 2) unchanged | TestLegacyMixedBuildProjectInstallProducesMarkerV3, TestGlobalMixedBuildStagesExternalBeforeLocal | Hosted-pending |
| B | Same §10 cause; legacy install must still never be the DRAFT shape | `internal/install/draftbuild_test.go:439` accept core v5 (Package==nil, hash 2, receipt 1); keep rejecting Package/receipt-3 | TestLegacyBuildsKeepReceipt1 | Hosted-pending |
| B | Published vectors at CI pin 43bf0a2 still carry marker_version 2/3/4 (spec-vector lag; normative §10 says v5) | `internal/install/external_lifecycle_conformance_test.go:407` pin v1 lane so the vector binding stays exact; no ledger change, counts stay 6 driven | TestAuthoritativeMixedBuildCasesUseProjectInstallEntry | Hosted-pending |
| C1 | Downgrade fixtures set version=1 but leave hash_version=2 from the v3 publish → Parse correctly rejects (§8.2: schemas 1/2 carry no hash_version) | `cmd/curator/env_credential_marker_test.go:155,212` delete hash_version in both downgrades | TestEnvResolveKeepsSchema1BytesForMetadataOnly, TestEnvResolvePreservesPreRuleCodexSeedAndReportsUnstrippedHome (also need P4) | Hosted-pending |
| C2 | `legacyMarkerProjectionMatches` projects candidate to Version=1 but keeps HashVersion=2 → projected.Marshal() always fails → repair always rewrites v1 markers | `internal/envprofile/managed.go:1470` reset HashVersion=0 in the V1 projection (+comment) | Same two schema-1-bytes tests (bytes-kept half) | Hosted-pending. Scoped: runs only for prior-V1 + lock-v2 (synthetic or post-migration states); genuine v1 pins never match v2 candidates, so real upgrades still publish |
| C3 | Writer publishes schema 3 per §8.2; tests pin schema 2 | `cmd/curator/env_credential_marker_test.go:44,91,124` expect VersionV3 + HashVersion 2; record-field assertions unchanged | TestEnvResolveCredentialRecordProvisionAndRepair, TestEnvResolveCredentialRecordPathlessKeyring, TestEnvResolveCredentialRecordIsolatedKeychain | Hosted-pending |
| D | Status readers recompute with frozen v1 (`hashing.ContentSHA256`) and compare against v2 digests → false content-drift, masking the true verdict (current/needs-install/unresolvable) | `internal/marker/marker.go:1383` new `(*Marker).ContentHashVersion()` (marker's own framing: v2 iff core v5, else v1); `cmd/curator/draft_status.go:114`, `cmd/curator/main.go:1320` (project+global drift), `cmd/curator/main.go:2094` (hybrid) recompute with it | TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands, TestCLIEndToEndInstallStatusAndTamperCheck, TestCuratorStatusProviderPostureAndCheck, TestGlobalStatusFailsCheckWhenTheClosureCannotBeProven, TestGlobalStatusKeepsTheDeclaredSkillSurfaceWithoutCompiledCommands, TestGlobalStatusReportsATransitivelyResolvedCompiledCommand, TestGlobalStatusReportsCompiledCurrentnessAndFailsCheck, TestCompiledProjectRepairsCorruptCompiledState/*, TestCompiledProjectRestoresCacheWhenCommitFails/*, TestCompiledProjectStatusAndUntrustedRecovery, TestStatusReportsATransitivelyResolvedCompiledCommand, TestStatusReportsAnUnusableToolchainPerCompiledCommand, TestGCRetainsAndReportsReferencedCompiledState, TestDraftStatusLegacyMarkerNeedsInstall, TestClassifyDraftMemberPresenceRows/legacy-marker | Hosted-pending. New unit test `TestContentHashVersionFollowsTheMarkerNotTheWriter` (`internal/marker/marker_hash_v2_test.go:35`) **local exit 0** proves the mapping under both switch positions |
| D | Legacy-marker rewrite left hash_version=2 + v2 digest under schema 1 → invalid, and would miscompare | `cmd/curator/status_test.go:1977` rewrite to genuine v1 (delete hash_version + v1 rehash) | TestStatusAcceptsAnUnchangedLegacyMarkerSchema | Hosted-pending (parallel test; cannot pin switch, hence genuine-bytes rewrite) |
| D | "Legacy marker" draft tests wrote v5 markers via the ambient writer; they no longer proved the legacy path | `cmd/curator/draft_status_test.go:30` `pinV1MarkerWriters` + use in both legacy tests (serial file, restores switch) | TestDraftStatusLegacyMarkerNeedsInstall, TestClassifyDraftMemberPresenceRows/legacy-marker (now genuine v1 proofs) | Hosted-pending |
| D | Same genuine-v1 cause in scopes downgrade | `internal/scopes/gc_test.go:507` downgrade deletes hash_version + v1 rehash | TestCollectMarksRuntimeFromEverySupportedMarkerSchema | **Local exit 0** |

## Commands actually run (real exit codes, GOFLAGS=-work, via mini-build-lock)

1. `~/.local/bin/mini-build-lock run uzji7 -- env GOFLAGS=-work go test ./internal/marker -run 'TestContentHashVersionFollowsTheMarkerNotTheWriter|TestWriteEmitsCoreMarkerV5WithHashVersion2|TestWritePreservesContentHashInRC13Mode|TestReadLegacyV1AndRewriteAsCoreMarkerV5WithHashVersion2' -count=1 -timeout=6m` → **exit 0**, `ok internal/marker 0.395s`.
2. `... go test ./internal/scopes -run 'TestCollectMarksBuildKeysFromEveryBuildBearingMarkerSchema|TestAbsorbKeysBuildLivenessOnTheSchemaBandNotOnAPopulatedBuildsMap|TestCollectMarksBuildKeysFromEveryLiveScope|TestCollectMarksRuntimeFromEverySupportedMarkerSchema|TestCollectSkipsTheBuildSweepOnUnprovableReferences|TestCollectSweepsOnlyUnreferencedProtectedEntries|TestNonDirectoryScopeMembersDoNotBlockMaintenance|TestARepeatedRegistryMemberNeverEmptiesTheRegistry|TestAuthoritativeGarbageCollectionRootsAreRetained' -count=1 -timeout=6m -v` → **exit 0**, 8 passed + 1 skipped (AuthoritativeGC skips without CURATOR_CONFORMANCE_ROOT; fails on hosted for the fixed helper reason, so hosted-pending, not locally proven).
3. `... go test ./internal/scopes -run 'TestCollectStaysFailSafeOnRedirectedUnixScopes' -count=1 -timeout=6m -v` → **exit 0**, 4/4 subtests pass.
4. `gofmt -l internal/ cmd/` → **exit 0**, no output (clean).
5. `git diff --check` → **exit 0**, no output.

## NOT run (binding rework4/host rules)

- `cmd/curator`, `internal/install`, `internal/envprofile`, `internal/crossconformance` tests: forbidden locally; hosted CR gate is the arbiter.
- `go vet`, lint, full suite, race/platform matrix: not run; hosted gate owns them.
- No `go build` was run (only the compilations implied by the allowed `go test` runs above).
- LOGBOOK.md, CHANGELOG.md, scripts/remote-gate.sh: untouched.
- Conformance counts/ledger: untouched (mixed_build_cases stays 6 driven via the v1 lane; no new gaps; the owned byte-exact-snapshot row removal from the carried delta is unchanged).

## Design note for review

- Every `WriteVersion()` call site was re-audited: all remaining production uses are writer decisions (marker shape, install build upgrade + attestation hash, context lock/store/materialize writes, env migration triggers, plan-hash binding). The three status readers that consumed the frozen v1 entry point now consume the marker's own version. `internal/registry/attest.go` still uses frozen v1 `Resolve` (unchanged, out of scope: no failure implicates it; source+commit matching still admits v1 records).
- macOS-only compiled-status failures are the same family-D cause: those tests skip on Linux via `requireNativeControlInventoryPlatform` (portable policy covers macOS/Windows only), which is why ubuntu stayed green on them.
- The `validCoreV5Identity` revision rule (ref==commit) is intentionally NOT weakened: substituted installs now record the resolved commit (provenance stays in `Substituted`). The rc.14 v5 JSON schema itself only requires a non-empty ref, so normalized markers remain schema-valid.
