# TASK-260917-2tx81l results

## Implementation and write policy

Implemented explicit hash versions in `internal/hashing`. V1 hashing remains byte-identical for v1 readers. V2 uses the `curator-content-v2\0` domain prefix, `F` entry records, uint64 big-endian path and content lengths, canonical ordering, and the specified empty-tree hash. Version participates in identity comparison, so v1 and v2 identities cannot match.

Carried and validated `hash_version` on install marker v5, context lock v2, agent environment marker v3, audit record v2, manager-config v3 waiver state, registry log entry v2, bundle v2, and log-response v3. Registry matching requires equal versions; context state pins and environment currentness checks also reject version mismatches. V1 reader shapes remain supported.

The single internal v2 writer switch defaults off and carries a TODO for the future rc.14 SPEC_PIN bump. With SPEC_PIN still at rc.13, production writers retain the rc.13 v1 shapes and existing on-disk classifications. V2 conformance-vector drivers opt in only when the vector declares v2. New manager-state reads use `internal/stateread`; managed writes use the existing nofollow helpers. No CHANGELOG or LOGBOOK file was edited.

## Candidate conformance gap ledger

Candidate spec commit: `b1a2efb6fa28d014968a2a8fd7641823b5f3cf28`; manifest SHA-256: `950ee74ad148615c273fe95bbb93f1bc0f9bdf2ea2bd9395f1dc8e3601419e60`.

| Family | Before | After | Result |
|---|---:|---:|---|
| content-hashes-v2 / vectors | 5 | 0 | 5 driven |
| registry-bundle-v2 / schema-cases | 2 | 0 | 2 driven |
| context-lock-v2 / schema-cases | 4 | 0 | 4 driven |
| marker/install-marker-v5 / schema-cases | 29 | 0 | 29 driven |
| manager-config-v3 / schema-cases | 5 | 0 | 5 driven |
| audit-record-v2 / schema-cases | 3 | 0 | 3 driven |
| registry-log-entry-v2 / schema-cases | 2 | 0 | 2 driven |
| log-response-v3 / schema-cases | 2 | 0 | 2 driven |
| agent-environment-marker-v3 / schema-cases | 28 | 0 | 28 driven |
| deferred skillfile-sources-v1 paths | 7 | 0 | Removed; absent from selected b1a2efb index |
| **Total task-owned rows** | **87** | **0** | **80 driven; 7 removed** |

The seven deferred paths were the install-marker-v6 (2), skillfile-lock-v2 (3), and source-audit-v2 (2) cases. They do not exist in the b1a2efb suite index, so they were removed rather than re-owned. The candidate skillfile-sources-v1 schema-case count pin was corrected from 131 to the actual 121.

## Suite and count-pin evidence

Disposable spec checkouts were tested at b1a2efb and the rc.13 SPEC_PIN. Both have 121 skillfile-sources-v1 schema cases. The ledger ratchet and `conformance-case-counts.tsv` exactness tests passed against both roots.

| Suite | Manifest SHA-256 | Count pins | Sum of expected-case pins |
|---|---|---:|---:|
| rc.13 (`23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`) | `be11bb1e4c46f21fb5684d586f9c2a8b0d59f3b437bc7ea7aa5aa530fe4d47ca` | 88 | 1,636 |
| b1a2efb (`b1a2efb6fa28d014968a2a8fd7641823b5f3cf28`) | `950ee74ad148615c273fe95bbb93f1bc0f9bdf2ea2bd9395f1dc8e3601419e60` | 97 | 1,722 |

## Verification run and exit codes

All commands below were direct standalone processes. Candidate commands used `CURATOR_CONFORMANCE_ROOT=/tmp/curator-spec-v2-task-260917/conformance/v1`; rc.13 commands used `CURATOR_CONFORMANCE_ROOT=/tmp/curator-spec-rc13-task-260917/conformance/v1`.

- The full affected package group (`internal/hashing`, `internal/conformancecoverage`, `internal/contextlock`, `internal/contextstore`, `internal/contextaudit`, `internal/contextresolve`, `internal/registry`, `internal/marker`, `internal/envmarker`, `internal/config`) exited **0** against each suite root with `go test -count=1`.
- The gatefix regression selectors covering the 35 pre-fix failures were split by test name and rerun against both suite roots; every command exited **0**, with the four reader-version assertion edits explained in the rev3 F2 section below. Coverage was cmd/curator (15), internal/scopes (9), internal/envprofile (3), internal/config (2), internal/environments (2), internal/install (2), internal/contextmaterialize (1), and internal/crossconformance (1).
- Five bounded envprofile selector groups required by the gatefix each exited **0** against both roots: waiver hash-version propagation; Pi agent-root targeting; credential-marker journal publish/recovery; schema-1 unlink migration; and manager-owned absence-read guards/alias scan.
- Candidate and rc.13 schema corpus checks (`TestDraftSourcesSchemaCases`, `TestSkillfileSourcesCorpusCounts`) exited **0**. Candidate deferred-case absence/unowned coverage exited **0**.
- Targeted install and registry E2E cases, including v1 evidence rejection for a v2 artifact and the explicitly enabled v2 attestation/install path, exited **0** against both roots.
- `go build ./...`, `go vet ./...`, `golangci-lint run` (0 issues), `gofmt -l cmd internal` (no files listed), and `git diff --check` each exited **0** after the changes.

### Mutant evidence

For each mutant, the legacy-only selector survived the mutant with exit **0**, while its v2 negative regression selector failed with exit **1**. The original files were restored and the package tests passed afterward.

- Removing v2 length framing: legacy hashing selectors exited 0; `TestContentSHA256V2LengthFramingSeparatesAdjacentRecordBoundary` failed with the expected collision, exit 1.
- Removing registry equal-version matching: legacy match selectors exited 0; `TestMatchesRequiresEqualHashVersion` and `TestResolveVersionedDoesNotAdmitV1RecordForV2Artifact` failed, exit 1.
- Removing marker currentness version comparison: legacy marker selectors exited 0; `TestCurrentRejectsV1MarkerForV2Expectation` failed, exit 1.

## Nonzero attempts and disposition

- Hosted run `36735846431` exited **1** before the gatefix corrections, with the reported 35 existing tests failing across platform lanes. After the changes, the affected selector groups listed above exited 0 locally against both rc.13 and b1a2efb. The handoff runner will publish the new Change Request and run the configured hosted gate after this turn.
- The broad `go test ./internal/install -count=1` process reached its 10-minute time bound and exited **1** while running `TestDraftRetargetFailsAtWriteTimeAndRollsBack`. The bounded install and registry E2E selectors listed above exited 0 against both roots.
- An earlier whole-package `go test ./internal/envprofile -count=1` was interrupted after about 390 seconds and exited **1**. All five required bounded envprofile groups then exited 0 against both roots.
- Initial candidate runs with the wrong conformance-root directory exited **1** because vectors/sibling corpus were not found; runs with the correct `/conformance/v1` roots passed. During development, candidate vector tests first exited 1 because their driver had not yet opted in by vector version; the driver was corrected and the reruns passed. The candidate corpus count check first exited 1 at 121 actual versus 131 pinned; the pin was corrected to 121 and both suite count checks passed. An initial lint run exited 1 on an unchecked G115 integer conversion; explicit version mapping was added and the final lint run exited 0.

## CHANGELOG entry (for release prep)

Content hashes now support domain-separated, length-framed v2 identities while preserving v1 compatibility. Manager-owned carriers validate and compare the framing version; v2 writes remain opt-in until the future rc.14 SPEC_PIN bump.


## Rev3 review rework: F1 and F2

### F1 — rc.13 marker writes preserve the supplied v1 identity

`internal/marker/marker.go` now recomputes `content_sha256` only when `markerHashVersion(m)` is v2. The code comment records that rc.13 writes preserve the caller-provided v1 identity; the explicit v2 writer switch controls the v2 recomputation. The previous unconditional v1 tree walk is gone.

The rc.13 test file `internal/marker/marker_v2_test.go` is byte-identical to base (`HEAD`): this includes `TestAuthoritativeCompiledMarkerRoundTripsThroughWriter`, `TestWriteAlwaysProducesCanonicalMarkerV2`, and both legacy rewrite tests. `marker_v3_test.go` and `marker_v4_test.go` are also byte-identical to base. The full marker package passed with these base tests under rc.13.

New v2-mode cases live in `internal/marker/marker_hash_v2_test.go` under new names:

- `TestWritePreservesContentHashInRC13Mode` checks marker schemas 2/3/4, the in-memory hash, and the serialized wire hash when the v2 switch is off.
- `TestWriteEmitsCoreMarkerV5WithHashVersion2` checks recomputation and the v5/hash-v2 wire identity for skill schema 6, 7, and 8.
- `TestAuthoritativeCompiledMarkerRoundTripsThroughV2Writer` exercises the compiled-marker fixture with v2 enabled.
- `TestReadLegacyV1AndRewriteAsCoreMarkerV5WithHashVersion2` and `TestReadLegacyV1AndRewriteAsCoreMarkerV5PreservesEmptyRequirer` cover v2-mode migration as separate tests.

The envprofile rc.13 tests were restored from base without changing their assertions. New v2 counterparts have separate names: `TestPiProvisionTargetsAgentRootV2WriterMode`, `TestResolvePublishesCredentialMarkerThroughLockedJournalV2WriterMode`, `TestResolveRecoversCredentialMarkerAfterRenameCrashV2WriterMode`, and `TestMigrateSchema1UnlinkPublishesCompleteSchema3MarkerV2WriterMode`.

### F2 — reader-version assertion edits

These four existing assertions now probe the next unsupported version because the prior value is a supported schema:

- `internal/config/config_test.go`, `TestParseRejections/schema`: manager-config v3 is now a known version; the rejection probe uses v4.
- `internal/config/environments_test.go`, `TestUnknownSchemaVersionsRejected`: manager-config v3 is now known; the unknown-version set uses v4.
- `internal/envmarker/envmarker_test.go`, `TestUnsupportedVersionIsRejected`: agent-environment marker v3 is now known; the unsupported-version probe uses v4.
- `internal/envmarker/envmarker_test.go`, `TestReadKeepsUnsupportedVersionDiagnostic`: v3 without its required hash version now reports a v3 validation error; v4 preserves the unsupported-version diagnostic case.

The two config tests, when run from their base versions against the current reader, exited 1 because v3 is accepted. The base envmarker run exited 1 because `TestReadKeepsUnsupportedVersionDiagnostic` received `version 3 requires hash_version 2`; `TestUnsupportedVersionIsRejected` passed. The current config and envmarker packages both passed against rc.13 and b1a2efb.

### Base-test replay (previous run)

I created a Go test overlay from git show HEAD:<path> at base 30b3d678 and replayed the base tests over the candidate. The overlay covered every currently modified tracked test file; internal/envprofile/state_read_guard_test.go was already byte-identical to base and was included in the envprofile replay.

Base test files covered:

- internal/config/config_test.go
- internal/config/environments_conformance_test.go
- internal/config/environments_test.go
- internal/conformancecoverage/content_hash_v2_gaps_test.go
- internal/contextaudit/contextaudit_test.go
- internal/contextlock/schema_conformance_test.go
- internal/contextmaterialize/system_module_admission_test.go
- internal/contextstore/contextstore_test.go
- internal/envmarker/envmarker_test.go
- internal/envmarker/marker_env_schema_test.go
- internal/envprofile/credential_link_test.go
- internal/envprofile/credential_record_test.go
- internal/envprofile/envprofile_policy_test.go
- internal/envprofile/managed_test.go
- internal/envprofile/migrate_test.go
- internal/install/registry_e2e_test.go
- internal/interop/environments/context_materialization_test.go
- internal/interop/environments/context_resolution_test.go
- internal/marker/schema_coverage_test.go
- internal/registry/registry_test.go
- internal/registry/schema_conformance_test.go

Results against rc.13:

- go test -count=1 -overlay=/tmp/curator-v2-base-tests.M4Mo1B/overlay.json -run '^Test' for internal/conformancecoverage, internal/contextaudit, internal/contextlock, internal/contextmaterialize, internal/contextstore, internal/marker, internal/registry, and internal/interop/environments: exit **0**.
- Base internal/config package replay: exit **1**, only TestParseRejections/schema and TestUnknownSchemaVersionsRejected fail because config v3 is now recognized. Their edited candidate probes use v4.
- Base internal/envmarker package replay: exit **1**, only TestReadKeepsUnsupportedVersionDiagnostic fails because v3 without hash_version: 2 now returns a known-schema validation error. The base TestUnsupportedVersionIsRejected selector passes with exit **0**; its candidate probe uses v4.
- The 15 base credential-link tests, 6 base credential-record/policy/read-scan tests, 18 base managed tests, and 22 base migration tests in internal/envprofile each pass in separate -run groups: exit **0** each.
- The 15 base registry E2E tests in internal/install/registry_e2e_test.go pass with exit **0**.

The four edited reader-version assertions are explained individually in the F2 section above. internal/marker/marker_v2_test.go, marker_v3_test.go, and marker_v4_test.go compare byte-identically to base; the full marker package and the authoritative compiled-marker round trip passed with the writer switch off.

### Review regression and mutant evidence

All mutant test commands were direct standalone runs. Mutant exit **1** is expected and means the named negative test killed that mutation. Each source file was restored from a byte-for-byte backup; the post-restore selectors passed.

| Mutant or regression | Legacy / narrowed survivor | Negative selector under mutation | Result after restore |
|---|---|---|---|
| Remove v2 length-frame writes from both hash paths | go test -count=1 -run '^(TestByteLayout|TestMarkerExcluded|TestEmptyTree|TestOrderIndependence|TestNormalize)$' ./internal/hashing: exit **0** | CURATOR_CONFORMANCE_ROOT=/tmp/curator-spec-v2-task-260917/conformance/v1 go test -count=1 -run '^(TestContentSHA256V2ConformanceVectors|TestContentSHA256V2SeparatesTheV1Collision|TestContentSHA256V2LengthFramingSeparatesAdjacentRecordBoundary)$' ./internal/hashing: exit **1**; the adjacent-record pair collided and vector digests differed | Same v2 selector with CURATOR_CONFORMANCE_ROOT=/tmp/curator-spec-v2-task-260917/conformance/v1: exit **0** |
| Remove equal-version checks from MatchesVersioned and MatchesExactVersioned | go test -count=1 -run '^(TestMatches|TestMatchesExact)$' ./internal/registry: exit **0** | go test -count=1 -run '^(TestMatchesRequiresEqualHashVersion|TestResolveVersionedDoesNotAdmitV1RecordForV2Artifact)$' ./internal/registry: exit **1**; both v1-for-v2 admissions were detected | Combined legacy and negative selectors: exit **0** |
| Remove the marker currentness version comparison | go test -count=1 -run '^(TestAuthoritativeCompiledMarkerRoundTripsThroughWriter|TestWriteAlwaysProducesCanonicalMarkerV2|TestReadLegacyV1AndRewriteAsV2WithoutChangingContentHash|TestReadLegacyV1AndRewriteV2PreservesEmptyRequirer)$' ./internal/marker: exit **0** | go test -count=1 -run '^TestCurrentRejectsV1MarkerForV2Expectation$' ./internal/marker: exit **1**; the v1 marker became current for a v2 expectation | Combined base and negative selectors: exit **0** |
| Reintroduce the reviewed F1 defect by recomputing a digest with the v2 writer switch off | The test runs in rc.13 mode | CURATOR_CONFORMANCE_ROOT=/tmp/curator-spec-rc13-task-260917/conformance/v1 go test -count=1 -run '^TestWritePreservesContentHashInRC13Mode$' ./internal/marker: exit **1** for schemas 6, 7, and 8 | Same selector: exit **0** |
| Narrow the v2 marker writer branch to skill schema 6 | TestWritePreservesContentHashInRC13Mode under the narrowed mutant: exit **0** | go test -count=1 -run '^TestWriteEmitsCoreMarkerV5WithHashVersion2$' ./internal/marker: exit **1** for schemas 7 and 8 | Combined rc.13 and v2 marker selectors: exit **0** |

TestWritePreservesContentHashInRC13Mode verifies both the in-memory marker and serialized wire digest stay caller-supplied, with no hash_version field, when the switch is off. The new TestWriteEmitsCoreMarkerV5WithHashVersion2 covers skill schemas 6, 7, and 8, so the narrowing mutant cannot pass by handling only one schema.

### Previous-run verification and evidence provenance

Reran directly in the previous run:

- Full test selectors for internal/hashing, internal/conformancecoverage, internal/contextlock, internal/contextstore, internal/contextaudit, and internal/contextresolve under both rc.13 and b1a2efb: exit **0** for each root.
- Full test selectors for internal/registry, internal/marker, internal/envmarker, and internal/config under both roots: exit **0** for each root.
- Context audit and context resolve packages after the final rc.13 reader adjustment under both roots: exit **0**.
- Six v2 envprofile selectors covering waiver propagation, environment-marker mismatch, Pi provisioning, journal publication/recovery, and migration under both roots: exit **0** for each root.
- go build ./...: exit **0**; go vet ./...: exit **0**; golangci-lint run: exit **0**, 0 issues; gofmt -l cmd internal: exit **0**, no files; git diff --check: exit **0**.
- Downloaded and inspected run 36735846431's test-evidence-* artifacts with the instructed gh run download command: exit **0**. The temporary downloaded directories were removed after review.

Accepted from the already-attached earlier results resource, not rerun in this rework:

- The 35 gatefix regression selectors across CLI, scopes, environment profile, config, environment interop, install, context materialization, and cross-conformance were reported green under both suite roots. The current rework reran the core packages, envprofile, marker, and the base-file replay listed above.
- Exact conformance pin calculations for both suite digests remain as previously verified: rc.13 has 88 pins / 1,636 expected cases; b1a2efb has 97 pins / 1,722 expected cases. The candidate ledger still has 80 driven rows removed and 7 deferred skillfile-sources rows removed because those case IDs are absent from the b1a2efb index.

The final v1 writer behavior preserves the caller-provided digest; the v2 write cutover remains off by default with the rc.14 TODO. No CHANGELOG or LOGBOOK file was edited.
### Final revision checks

- `go build ./...`: exit 0.
- `go vet ./...`: exit 0.
- `golangci-lint run`: exit 0, 0 issues.
- `gofmt -l cmd internal`: exit 0, no files listed.
- `git diff --check`: exit 0.

No CHANGELOG or LOGBOOK file was edited. The release-prep migration entry remains in the section above. The continuation read the task README, the original framing Story README, the review verdict, and the campaign producer rules from the authoritative board.


## Post-counterpart rerun — 2026-10-01

After adding the separate compiled-marker and legacy-rewrite v2 cases, `go test -count=1 ./internal/marker` exited 0 under both rc.13 and b1a2efb, including the unchanged authoritative rc.13 writer round-trip. Post-addition validation also exited 0: `go build ./...`, `go vet ./...`, `golangci-lint run` (0 issues), `gofmt -l cmd internal` (no files), and `git diff --check`.


## Rework continuation verification — 2026-10-01

Resumed the existing F1/F2 changes without replacing them. The only additional test edit improves `TestWritePreservesContentHashInRC13Mode`'s failure diagnostic to show both the supplied and recomputed digest. All three mutant source files were restored byte-for-byte from backups. No CHANGELOG or LOGBOOK edit, commit, branch switch, or stray repository artifact was introduced.

### Base assertions replayed over the candidate

Built and byte-verified a fresh Go overlay from base `30b3d678` for all 21 modified tracked test files listed above. The three existing marker writer test files are byte-identical to base. The original credential-link, credential-record and migration files retain their entire base contents, with v2 tests appended under new names.

Every command below was run directly, with real process exits. `CURATOR_CONFORMANCE_ROOT` selected rc.13; each replay used `-count=1` and the base overlay.

| Base replay scope | Successful process exits |
|---|---|
| conformancecoverage, contextaudit, contextlock, contextmaterialize, contextstore, marker, registry, interop/environments | 0 |
| Credential links: 15 original tests | 0 |
| Credential records, policy and state-read scan: 6 original tests | 0 |
| Managed environment: 18 original tests, split 5 / 4 / 1 / 5 / 3 | 0 for every successful group |
| Migration: 22 original tests, split 3 / 3 / 3 / 2 / 11 | 0 for every successful group |
| Install registry E2E: 15 original tests, split 4 / bootstrap / mirror / 9 | 0 for every successful group |
| Config base tests | 1, expected: only TestParseRejections/schema and TestUnknownSchemaVersionsRejected fail for newly recognized v3 |
| Environment-marker base tests | 1, expected: only TestReadKeepsUnsupportedVersionDiagnostic fails for the newly recognized v3 diagnostic |
| Base TestUnsupportedVersionIsRejected in isolation | 0 |

The four reader-version assertion edits are explained per test in F2 above. No rc.13 writer assertion was weakened. The companion JSON outcome includes the exact base-file list, selectors, observed test outcomes and candidate file hashes.

### Current candidate checks rerun in this continuation

- Hashing, registry, marker, conformancecoverage, config and envmarker package tests: exit 0 against both rc.13 and b1a2efb.
- Contextaudit, contextresolve, contextlock and contextstore package tests: exit 0 against both roots.
- Eight bounded envprofile selectors covering v2 waiver propagation, marker mismatch, Pi provisioning, journal publication/recovery, migration and state-read guards: exit 0 against both roots.
- After restoring all mutations: the hashing/registry/marker negative selectors exit 0. The last marker regression and authoritative rc.13 round-trip selector exits 0.
- After the diagnostic edit and mutant restoration: `go build ./...`, `go vet ./...`, `golangci-lint run` (0 issues), `gofmt -l cmd internal` (no files), and `git diff --check` each exit 0.

The spec checkouts resolve to rc.13 `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065` / manifest `be11bb1e4c46f21fb5684d586f9c2a8b0d59f3b437bc7ea7aa5aa530fe4d47ca`, and b1a2efb `b1a2efb6fa28d014968a2a8fd7641823b5f3cf28` / manifest `950ee74ad148615c273fe95bbb93f1bc0f9bdf2ea2bd9395f1dc8e3601419e60`. Count-table totals were recomputed as 88/1,636 and 97/1,722. Task-owned gap rows remain 0: 80 core-v2 rows driven, seven absent deferred rows removed. Exact all-family count verification and the 35 historical gatefix selectors outside the replay scopes above are accepted from the earlier attached evidence; the hosted gate is the runner's next step after handoff.

### Mutants and the F1 regression

The three required mutations each survive the legacy selector with exit 0 and fail their v2 negative selector with exit 1: removed length framing, removed registry equal-version checks, and a v1 marker admitted for a v2 expectation. Their restored selectors exit 0.

The additional narrowing mutant preserves supplied rc.13 hashes only for skill schema 6, while recomputing them for schemas 7 and 8. `TestAuthoritativeCompiledMarkerRoundTripsThroughWriter` still exits 0 under that mutation; the new production-Write regression `TestWritePreservesContentHashInRC13Mode` exits 1 for subtests 7 and 8. After restoration it exits 0. Coverage is 3/3 schema bands, with the narrowing detected in 2/3. This was rerun after improving the diagnostic.

### Failed and interrupted attempts, retained as failures

- The initial 18-test managed group, 11-test first migration group and 15-test install group each exceeded `-timeout=4m` and exited 1.
- A five-test managed subgroup and four-test install subgroup each exceeded `-timeout=6m` and exited 1. Smaller groups and isolated long tests subsequently exited 0.
- An initial isolated bootstrap replay and a three-test managed replay returned exit 143. The latter log contained a package PASS, but the process exit remains 143 and is not counted as passing evidence. Both were rerun successfully with exit 0; no cancellation directive was present.
- Base reader probes retain their real expected nonzero exits as listed above. Required mutant runs retain their expected exit 1.

All replay processes have returned before handoff. No long command is left running. Local verification is reported here; this continuation makes no new hosted-platform claim.


### Handoff precondition repair

The first `task-board handoff TASK-260917-2tx81l --role developer` returned exit 1 because checklist item 15 was unchecked: “If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches.” The rev2 changes-requested verdict is already attached as `TASK-260917-2tx81l_review-verdict-rev2.md`, the task was explicitly routed to development at the start of this continuation, and the F1/F2 response is attached in these results and the new verification outcome. I check that already-satisfied review-routing item and retry the developer handoff. The command also emitted a board-content comparison warning; I do not count that attempt as a successful handoff.


## Revision 4 (refresh)

This is a base refresh of accepted revision 3, with no new product change. The revision 3 verdict was read and says ACCEPTED, with no blocking findings; the generic review-round rejection wording does not describe that verdict. Its original acceptance was released because trunk moved on the count table.

`task-board worktree refresh-candidate TASK-260917-2tx81l` returned exit **0**, `refresh_advanced`, onto `bd126a9acdc51b6061917ba8c4d7d26a7abafd41`. A follow-up invocation returned exit **0**, `refresh_already_current`. No checkpoint replay conflict occurred, so no `--replay-resolutions` packet was needed. No commit, branch switch, or hand-committed replay was made.

### Identity and resolution

The incoming candidate matched accepted tree `92a072c4e1b03678ea64a2801069ea471628e8e4` on **45/45** owned paths before refresh. All four overlapping paths merged without conflict. The refresh command preserves the supplied working tree, so all 23 trunk-only non-board paths were also explicitly carried forward and byte-checked against trunk. No board files were edited manually.

After convergence, **41/45** owned paths remain byte-identical to revision 3, including every test file and all five added files. The other four differ only by incoming trunk changes:

- Counts: keep the candidate's 131→121 correction under the candidate digest and trunk's four acquisition rows under the rc.13 digest.
- Config: retain trunk's typed `ErrConfigNotFound` and wrapped absence error beside the existing v3 reader.
- Managed environment: retain trunk's `--repair` hint for stale Resolve failures beside the existing version checks.
- Install targets: retain trunk's artifact basename selection beside the existing versioned hash staging.

No assertion or product behavior was changed by this refresh. The source identity check was repeated after mutation testing, and the index is empty. The complete non-board diff against accepted revision 3 is **27 paths, 1,880 insertions, 33 deletions**, entirely the incoming trunk delta. Per-file statistics follow; untracked candidate additions were independently blob-checked so Git's tracked-only diff cannot misclassify them as deletions.

| Path | Added | Removed | Source |
|---|---:|---:|---|
| `.github/ci/platform-cases.tsv` | 1 | 0 | trunk-only |
| `README.md` | 1 | 1 | trunk-only |
| `cmd/curator/env.go` | 4 | 0 | trunk-only |
| `cmd/curator/firstrun_ux_test.go` | 285 | 0 | trunk-only |
| `cmd/curator/main.go` | 68 | 1 | trunk-only |
| `cmd/curator/native_blackbox_test.go` | 254 | 0 | trunk-only |
| `cmd/curator/profile.go` | 4 | 0 | trunk-only |
| `cmd/curator/umbrella.go` | 1 | 0 | trunk-only |
| `docs/cli.md` | 13 | 4 | trunk-only |
| `docs/external-build-repositories.md` | 75 | 0 | trunk-only |
| `internal/buildrepo/acquisition_conformance_test.go` | 775 | 0 | trunk-only |
| `internal/buildrepo/admission.go` | 6 | 3 | trunk-only |
| `internal/buildrepo/admission_test.go` | 68 | 9 | trunk-only |
| `internal/buildrepo/adopt_test.go` | 2 | 2 | trunk-only |
| `internal/buildrepo/protected.go` | 29 | 3 | trunk-only |
| `internal/buildrepo/protected_artifact_test.go` | 67 | 0 | trunk-only |
| `internal/buildrepo/protection_windows_test.go` | 2 | 2 | trunk-only |
| `internal/buildrepo/testdata/acquisitiongitshim/main.go` | 77 | 0 | trunk-only |
| `internal/crossconformance/draftsources_semantic_external_test.go` | 3 | 1 | trunk-only |
| `internal/envprofile/named_absence_boundary_test.go` | 56 | 0 | trunk-only |
| `internal/envprofile/status.go` | 9 | 2 | trunk-only |
| `internal/install/external.go` | 2 | 2 | trunk-only |
| `internal/pathboundary/named_absence_test.go` | 64 | 0 | trunk-only |
| `.github/ci/conformance-case-counts.tsv` | 4 | 0 | merged |
| `internal/config/config.go` | 5 | 1 | merged |
| `internal/envprofile/managed.go` | 4 | 1 | merged |
| `internal/install/targets.go` | 1 | 1 | merged |

### Exact count recomputation

Ran `python3 /tmp/TASK-260917-2tx81l-refresh4/counts.py` directly: exit **0**. It reads the two suite manifests, schema indexes, sibling skillfile corpus, vector arrays and external-repository fixture arrays, and applies the production consumers' operation/phase splits and CRC32 semantic batching. It asserts every declared pin is resolved and equal; no counts were guessed.

| Suite | Manifest digest | Exact declared pins | Sum of cases |
|---|---|---:|---:|
| rc.13, commit `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065` | `be11bb1e4c46f21fb5684d586f9c2a8b0d59f3b437bc7ea7aa5aa530fe4d47ca` | 92/92 | 1,731 |
| b1a2efb, commit `b1a2efb6fa28d014968a2a8fd7641823b5f3cf28` | `950ee74ad148615c273fe95bbb93f1bc0f9bdf2ea2bd9395f1dc8e3601419e60` | 97/97 | 1,722 |

The rc.13 acquisition arrays recompute to 12 cases, 17 clean-environment entries, 55 fetch arguments and 11 forbidden features: 95 additional declared cases. The historical rc.13 88/1,636 total above is superseded for revision 4 by 92/1,731. Candidate counts remain unchanged. Both sibling skillfile schema indexes still contain exactly 121 cases. This exactness claim covers the declared pins, not every possible unpinned suite family.

### Verification rerun here

All validation commands are standalone processes with output redirected directly to individual log files, preserving their real exit status; no gate was piped through tee. Commands and log SHA-256 digests are in the attached verification JSON.

| Command / scope | rc.13 exit | Candidate exit |
|---|---:|---:|
| `go test ./internal/conformancecoverage -count=1` | 0 | 0 |
| Content-hash v2, carrier schema and mismatch regression selectors | 0 | 0 |
| `go test ./internal/crossconformance -count=1 -run '^(TestDraftSourcesSchemaCases|TestSkillfileSourcesCorpusCounts)$' -v` | 0 | 0 |
| Full hashing, registry, marker, contextlock, envmarker, config and conformancecoverage packages, `-count=1 -v`, after all trunk paths were carried | 0 | 0 |
| `bash .github/ci/ledger-consistency.sh <evidence-dir>` across linux/darwin/windows compile inventories | 0 | 0 |

The current candidate log records **80/80** owned v2 cases as driven across all nine families, with zero known gaps, bounds or skips for these rows. The 80 core rows and seven absent deferred rows remain removed; the gap file is byte-identical to accepted revision 3.

Additional current-tree checks: `go build ./...` exit **0**; `go vet ./...` exit **0**; `golangci-lint run` exit **0**, **0 issues**; `git diff --check -- . ':!.task-board'` exit **0**. The newly carried acquisition production-entry conformance test and `TestResolveStaleUnprovisioned` each ran under rc.13 with `-count=1 -timeout=4m -v` and exit **0**. No process was left running.

### Regression and narrowing evidence

No new regression test was added because the binding refresh brief forbids product changes and all existing tests are byte-identical to revision 3. The existing named regression `TestWritePreservesContentHashInRC13Mode` was rerun, along with its narrowing mutant: preserve caller-supplied hashes only for skill schema 6 while recomputing schemas 7 and 8. The authoritative compiled-marker control still passes; the named regression fails in both schemas 7 and 8. Its unmutated coverage remains **3/3** schema bands, with the narrowed regression detected in **2/3**.

All four mutants were applied through canonical-path Go overlays outside the repository; source files were never changed.

| Mutant | Separate legacy-control exit | Negative exit | Unmodified regression exit |
|---|---:|---:|---:|
| M1: remove v2 length writes; adjacent-record collision test | 0 | 1 | 0 |
| M2: omit both registry version comparisons; mismatch and ResolveVersioned admission tests | 0 | 1 | 0 |
| M3: admit a v1 marker for a v2 expectation; Current regression | 0 | 1 | 0 |
| F1 narrowing: preserve supplied hashes only for schema 6; rc.13 Write regression | 0 | 1 | 0 |

The nonzero mutant commands are expected failures: each admits the invalid behavior and its negative test detects it. The unmodified regressions were rerun after all overlay commands with exit **0**. Coverage is **3/3** required mutants plus the prior F1 narrowing regression. Legacy survivor controls describe current mutated code, not a new claim that these v2-only sites existed in the original base.

### Evidence carried from revision 3 and limits

Accepted from the attached revision 3 verdict and evidence, not rerun here: the independent base-test overlay (21 modified files, 76/76 individually passing environment/install selectors), the four reader-version exceptions, and the revision 3 hosted platform gate. Every existing test file remains identical to that accepted candidate. This refresh makes no claim that the full unreleased b1a2efb suite is green; the previously reproduced snapshot-acquisition vector/driver mismatch in the rev3 verdict remains outside this task's 80 core rows. The hosted gate for revision 4 belongs to the runner after handoff.

No CHANGELOG or LOGBOOK was edited. The release-prep migration text remains under `## CHANGELOG entry (for release prep)` above. New task-scoped verification and log artifacts are attached before the developer handoff.
