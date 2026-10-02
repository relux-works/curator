# TASK-261001-3bsyvh — carry-rjxrgs-marker-v3-lifecycle

Ready for review. Recovery run RUN-261001-668160 preserves the unchanged delta already staged by RUN-261001-803009; no repository files were edited in this recovery run. The prior run applied the campaign patch without conflicts, but ended without a confirmed results attachment or role handoff.

## Identity

Accepted source: `git diff 5432c85f refs/campaign/rjxrgs-rev1-20261001 -- . ':!.task-board'`.
Current delta against origin/main has exactly 17 paths, 978 additions and 43 deletions. Every per-path +/- count equals the accepted delta, every affected file equals its accepted blob, and the entire binary-capable patch is byte-identical. Both patch SHA-256 values: `2b9868771156089b7a4ac4200dec967a71c71f9b59b998fd37c79e9540cde45c`.
The required path-count command returned 17 with exit 0; identity assertions and reverse-apply check exited 0. The prior results filename is absent from tracked paths and the worktree filesystem. No conflicts, extra paths, changelog/logbook changes, commits or unstaged edits. Per-path counts and OIDs are recorded in the recovery identity JSON inside the attached evidence archive.

## Commands rerun directly in this recovery run

All commands below exited 0 unless explicitly marked failing. Tests were standalone processes, without pipelines. Root-enabled tests used `CURATOR_CONFORMANCE_ROOT=$TMPDIR/TASK-261001-3bsyvh_spec-pin/conformance/v1`, the prior run's archive of CI revision `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`.

- `go build -p 1 ./...`: exit 0.
- `go vet -p 1 ./...`: exit 0.
- `golangci-lint run --concurrency 1 --timeout 5m`: exit 0, zero issues.
- `gofmt -l` over all 14 changed Go files: exit 0, no unformatted files. Formatting assertion exited 0.
- `git diff --check origin/main -- . ':!.task-board'`: exit 0.
- Root-enabled `go test -json -count=1 -p 1 -timeout 3m ./internal/buildrepo -run '^(TestExternalProtectedCache.*|TestExternalReceipt3.*|TestExternalReceiptV2CacheKeyVector|TestSkillBuildDescriptorRejectsPackageOutputAndPathDestinations|TestCollectSweepsTheReceipt3Namespace|TestPrepareNamespaces.*)$'`: exit 0; 56 passing test events, zero failures/skips. Exercises receipt-evidence rejection, namespace separation, wrapper binding, cache/refusal behavior and descriptor rejection.
- Root-unset `go test -json -count=1 -p 1 -timeout 5m ./internal/install -run '^(TestLegacyMixedBuildProjectInstallProducesMarkerV3|TestExternalCommandNameCollisionFailsBeforeMutation|TestGlobalMixedBuildStagesExternalBeforeLocal|TestSecondBuildFailurePreservesPriorInstallationAndLiveCache|TestSourceMutationDuringStagingBlocksHandoffOfACacheHit)$'`: exit 0; 5 passing events, zero failures/skips. Drives install.Project and global/staging paths, including collision rejection and rollback/source-mutation checks.
- Root-enabled `go test -json -count=1 -p 1 -timeout 3m ./internal/marker ./internal/scopes -run '^Test(ReadAuthoritativeMarkerV[234]SchemaCases|AuthoritativeGarbageCollectionRootsAreRetained)$'`: exit 0; 82 passing events, zero failures/skips.
- Root-enabled `go test -json -count=1 -p 1 -timeout 3m ./internal/skillspec`: exit 0; 425 passing events, zero failures/skips.
- **Failing:** `go test -count=1 -p 1 -timeout 3m ./internal/buildrepo`: exit 1 after 182.184 seconds. Package timeout while running TestResolvedTransportClosedTableFramingStillFallsBack. This was a broader package run, not a successful suite. The focused affected checks were rerun afterward and passed.

## Prior-run evidence accepted and retained

The prior run's already-attached buildrepo/install-focused streams pass. Additional raw split conformance streams survived in TMPDIR and are now attached in the evidence archive. Their package terminal events are pass; the prior transcript reports each command exiting 0. They were not rerun in this recovery session: four standalone root-enabled commands `go test -json -count=1 -timeout 8m ./internal/install -run '^TestAuthoritativeMixedBuildCasesUseProjectInstallEntry$'`, repeated for exact names TestAuthoritativePathShimCasesUseProjectInstallEntry, TestAuthoritativeSigningCasesUseProjectInstallEntry, and TestAuthoritativeTransactionCasesUseProjectInstallEntry. The production entry point is install.Project.

Published-case coverage (not raw test-event counts):

| Family | Driven / total | Known gaps | Bounds | Skipped |
| --- | --- | --- | --- | --- |
| Mixed build, prior run | 6/6 | 0 | 0 | 0 |
| Path/shim, prior run | 3/3 | 0 | 0 | 0 |
| Signing, prior run | 3/4 | 0 | 1 | 0 |
| Transaction, prior run | 4/4 | 0 | 0 | 0 |
| Marker v2, fresh | 14/14 | 0 | 0 | 0 |
| Marker v3, fresh | 22/27 | 5 | 0 | 0 |
| Marker v4, fresh | 22/27 | 5 | 0 | 0 |
| Status/repair/GC, fresh | 1/5 | 0 | 4 | 0 |

Signing bound: external release-pipeline signing has no install.Project operation. Status/repair bounds: the manager does not expose the required external status/protected-state facts or repair entry point. Marker gaps: declared/effective revision mismatch, local/network identity-kind mismatches, and SHA-1/SHA-256 effective revision width validation; ledger rows preserve ownership. These limits are part of the accepted delta and have not been altered.

Prior failing commands remain failures (reported process exits 1): the broad five-package five-minute test run timed out in buildrepo/install; an initial root-enabled conformance run started before extraction returned and failed reading missing fixtures; a combined conformance JSON run timed out in install (retained in the archive), then its groups were split and rerun successfully. Prior pin-verifier command `go run ./cmd/curator-spec-pin --root "$TMPDIR/TASK-261001-3bsyvh_spec-pin" --revision 23435129ebc4c29e5b7f75ec72a0aa0cd3f16065` exited 1: CI uses rc.13 while the compiled verifier expects rc.8 `f8c405aa3ad0a39d260c2ed93684e55c5a346359`. Neither pin-defining file belongs to the accepted delta; no change was made. The pin verifier was not rerun here.

Full repository tests, race tests and hosted platform execution were not rerun. Prior static ledger checks are not proof of cross-platform runtime tests. Existing accepted tests were carried unchanged; adding tests or fixing unrelated validation issues would violate the carrier brief. The explicit prohibition on LOGBOOK/CHANGELOG edits makes the conditional logbook item N/A; findings are recorded here and on the board.

All deliverables are outside the repository and attached through task-board resource CRUD. Code remains uncommitted for the developer handoff snapshot.


## Revision 2 (refresh)
Run RUN-261001-b43aa3. Ready for review. This is a base refresh of the accepted carrier, with no new product behavior or test-code changes. Revision 1 verdict TASK-261001-3bsyvh_review-verdict-rev1.md explicitly says ACCEPTED, rather than rejected; there is no rejection finding to answer. The existing tests are carried unchanged.
### Refresh and identity
Initial status and worktree-status commands exited 0. The first refresh-candidate exited 1: staged state outside checkpoint baseline. All 17 current file blobs first matched revision 1, were preserved externally, and were unstaged. Two overlapping code files merged cleanly with git merge-file (exit 0 each); the counts-file merge returned exit 1 with an adjacency conflict. Resolution retains both sets of rows. The second refresh-candidate exited 0 with refresh_advanced onto bd126a9acdc51b6061917ba8c4d7d26a7abafd41. No replay-checkpoint conflict appeared, so --replay-resolutions was unnecessary; no replay worktree was committed.
The refresh preserves the entire old candidate: 22 trunk-only paths initially reverted incoming content. Those paths were restored exactly from bd126a9a and all 22 blobs independently verified. Infrastructure-owned .task-board projection differences were excluded from every source proof; no board file was edited directly.
Final non-board candidate tree: `d87c0e707db9a1ccca5a51b820ef109d5618501d`. `git diff --name-only origin/main -- . ':!.task-board' | wc -l` returned 17, exit 0. Every path has exactly the accepted +/- line multiset in both directions and the accepted per-file numstat. Total remains 978 additions / 43 deletions. Fourteen candidate blobs equal revision 1; protected.go and external.go differ solely by trunk Windows artifact-path changes; the counts file differs solely by four trunk acquisition rows. These shared-path differences are required base movement, not product edits.
Full-tree proof: git merge-tree --write-tree --merge-base=bab2433b bd126a9a 1e1c5d83 returned exit 1 (expected counts-file conflict; not a passing merge). Replacing only that conflicted blob with the independently verified union produced tree `d87c0e707db9a1ccca5a51b820ef109d5618501d`, exactly the candidate tree. This proves the two clean shared-code merges and every trunk-only path. The identity verification script exited 0 before and after the temporary mutant. No non-board stray paths, old results file, CHANGELOG or LOGBOOK changes. All carrier additions/removals equal the accepted delta, so no new name-bearing content was introduced.
Carrier stats against new trunk (also attached in carrier-per-file-numstat.tsv):
| Path | + | - |
| --- | ---: | ---: |
| `.github/ci/conformance-case-counts.tsv` | 4 | 0 |
| `.github/ci/conformance-gaps.tsv` | 5 | 0 |
| `.github/ci/root-artifacts.tsv` | 2 | 1 |
| `internal/buildrepo/buildrepo.go` | 5 | 0 |
| `internal/buildrepo/buildrepo_test.go` | 12 | 0 |
| `internal/buildrepo/pipeline.go` | 14 | 4 |
| `internal/buildrepo/pipeline_test.go` | 44 | 9 |
| `internal/buildrepo/protected.go` | 1 | 1 |
| `internal/buildrepo/receipt_v3_test.go` | 17 | 10 |
| `internal/install/external.go` | 9 | 0 |
| `internal/install/external_lifecycle_conformance_test.go` | 813 | 0 |
| `internal/install/global.go` | 9 | 9 |
| `internal/install/install.go` | 12 | 7 |
| `internal/install/stage_test.go` | 8 | 1 |
| `internal/marker/schema_coverage_test.go` | 4 | 0 |
| `internal/scopes/gc_conformance_test.go` | 16 | 1 |
| `internal/skillspec/parse.go` | 3 | 0 |

Stats versus revision 1 (all differences are trunk movement; the remaining 14 carrier files have no blob difference):
| Path | + | - |
| --- | ---: | ---: |
| `.github/ci/conformance-case-counts.tsv` | 4 | 0 |
| `.github/ci/platform-cases.tsv` | 1 | 0 |
| `README.md` | 1 | 1 |
| `cmd/curator/env.go` | 4 | 0 |
| `cmd/curator/firstrun_ux_test.go` | 285 | 0 |
| `cmd/curator/main.go` | 68 | 1 |
| `cmd/curator/native_blackbox_test.go` | 254 | 0 |
| `cmd/curator/profile.go` | 4 | 0 |
| `cmd/curator/umbrella.go` | 1 | 0 |
| `docs/cli.md` | 13 | 4 |
| `docs/external-build-repositories.md` | 75 | 0 |
| `internal/buildrepo/acquisition_conformance_test.go` | 775 | 0 |
| `internal/buildrepo/admission.go` | 6 | 3 |
| `internal/buildrepo/admission_test.go` | 68 | 9 |
| `internal/buildrepo/adopt_test.go` | 2 | 2 |
| `internal/buildrepo/protected.go` | 29 | 3 |
| `internal/buildrepo/protected_artifact_test.go` | 67 | 0 |
| `internal/buildrepo/protection_windows_test.go` | 2 | 2 |
| `internal/buildrepo/testdata/acquisitiongitshim/main.go` | 77 | 0 |
| `internal/config/config.go` | 5 | 1 |
| `internal/crossconformance/draftsources_semantic_external_test.go` | 3 | 1 |
| `internal/envprofile/managed.go` | 4 | 1 |
| `internal/envprofile/status.go` | 9 | 2 |
| `internal/install/external.go` | 2 | 2 |
| `internal/install/targets.go` | 1 | 1 |

### Counts resolution
Manifest SHA-256: `be11bb1e4c46f21fb5684d586f9c2a8b0d59f3b437bc7ea7aa5aa530fe4d47ca`, from the existing archive of workflow revision `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`. Independently loaded vector arrays measured lifecycle mixed/shim/signing/transaction counts 6/3/4/4 and incoming acquisition cases/environment/argv/forbidden-feature counts 12/17/55/11. Assertions matched every measured count to the merged TSV, exit 0. The union retains every existing manifest+family key, including unchanged content-hash-v2 pins; it adds four acquisition rows and contains 193 keys without duplicate keys. No counts guessed. Existing coverage parsing and production-entry consumers verify the merged policy.
### Validation rerun in this run
Each command ran as a standalone process with direct output redirection, no tee/pipeline. Root-enabled commands below use the pinned root. Exit statuses are from the actual processes. JSON streams and command records are attached in the revision 2 archive.
- `env CURATOR_CONFORMANCE_ROOT="$TMPDIR/TASK-261001-3bsyvh_spec-pin/conformance/v1" go test -json -count=1 -p 1 -timeout 3m ./internal/marker -run '^TestReadAuthoritativeMarkerV[234]SchemaCases$'`: exit 0 (stream `marker-final.log`).
- `go test ./internal/conformancecoverage -count=1`: exit 0 (stream `coverage-final.log`).
- `env CURATOR_CONFORMANCE_ROOT="$TMPDIR/TASK-261001-3bsyvh_spec-pin/conformance/v1" bash .github/ci/ledger-consistency.sh /tmp/TASK-261001-3bsyvh-refresh/ledger`: exit 0 (stream `ledger.log`).
- `go build -p 1 ./...`: exit 0 (stream `build.log`).
- `env CURATOR_CONFORMANCE_ROOT="$TMPDIR/TASK-261001-3bsyvh_spec-pin/conformance/v1" go test -json -count=1 -p 1 -timeout 5m ./internal/install -run '^(TestLegacyMixedBuildProjectInstallProducesMarkerV3|TestExternalCommandNameCollisionFailsBeforeMutation|TestGlobalMixedBuildStagesExternalBeforeLocal|TestAuthoritativeMixedBuildCasesUseProjectInstallEntry|TestLegacyBuildsKeepReceipt1|TestDraftBuildsPublishReceipt3OnBothArms|TestSecondBuildFailurePreservesPriorInstallationAndLiveCache|TestSourceMutationDuringStagingBlocksHandoffOfACacheHit)$'`: exit 0 (stream `mixed-final.log`).
- `env CURATOR_CONFORMANCE_ROOT="$TMPDIR/TASK-261001-3bsyvh_spec-pin/conformance/v1" go test -json -count=1 -p 1 -timeout 3m ./internal/buildrepo -run '^(TestExternalReceiptV2CacheKeyVector|TestExternalReceipt3.*|TestExternalProtectedCache.*|TestSubstitutionCannotAliasDeclaredCacheKey|TestProtectedArtifact.*|TestCacheArtifact.*|TestCollectSweepsTheReceipt3Namespace|TestPrepareNamespaces.*|TestExternalRepositoryAcquisitionConformance/(common-fetch-argv|clean-environment|forbidden-fetch-features))$'`: exit 0 (stream `receipts.log`).
- `go vet -p 1 ./...`: exit 0 (stream `vet.log`).
- `golangci-lint run --concurrency 1 --timeout 5m`: exit 0 (stream `lint.log`).
- `env CURATOR_CONFORMANCE_ROOT="$TMPDIR/TASK-261001-3bsyvh_spec-pin/conformance/v1" go test -json -count=1 -p 1 -timeout 3m ./internal/buildrepo -run '^TestExternalRepositoryAcquisitionConformance$/^(common-fetch-argv|clean-environment|forbidden-fetch-features)$'`: exit 0 (stream `acquisition-count-tables.log`).
- `env CURATOR_CONFORMANCE_ROOT="$TMPDIR/TASK-261001-3bsyvh_spec-pin/conformance/v1" go test -json -count=1 -p 1 -timeout 3m ./internal/skillspec`: exit 0 (stream `skillspec.log`).
- `env CURATOR_CONFORMANCE_ROOT="$TMPDIR/TASK-261001-3bsyvh_spec-pin/conformance/v1" go test -json -count=1 -p 1 -timeout 3m ./internal/scopes ./internal/skillspec -run '^(TestAuthoritativeGarbageCollectionRootsAreRetained|TestSkillBuildDescriptorRejectsPackageOutputAndPathDestinations)$'`: exit 0 (stream `scopes-skillspec.log`).
- `env CURATOR_CONFORMANCE_ROOT="$TMPDIR/TASK-261001-3bsyvh_spec-pin/conformance/v1" go test -json -count=1 -p 1 -timeout 4m ./internal/install -run '^TestAuthoritativePathShimCasesUseProjectInstallEntry$'`: exit 0 (stream `path-shim.log`).
- `env CURATOR_CONFORMANCE_ROOT="$TMPDIR/TASK-261001-3bsyvh_spec-pin/conformance/v1" go test -json -count=1 -p 1 -timeout 4m ./internal/install -run '^TestAuthoritativeSigningCasesUseProjectInstallEntry$'`: exit 0 (stream `signing.log`).
- `env CURATOR_CONFORMANCE_ROOT="$TMPDIR/TASK-261001-3bsyvh_spec-pin/conformance/v1" go test -json -count=1 -p 1 -timeout 5m ./internal/install -run '^TestAuthoritativeTransactionCasesUseProjectInstallEntry$'`: exit 0 (stream `transactions.log`).
- `env CURATOR_CONFORMANCE_ROOT="$TMPDIR/TASK-261001-3bsyvh_spec-pin/conformance/v1" go test -json -count=1 -p 1 -timeout 3m ./internal/buildrepo -run '^TestExternalRepositoryAcquisitionConformance$/^common-fetch-argv$'`: exit 0 (stream `counts-restored.log`).
- Formatting: gofmt -l over all 14 carrier Go files exited 0, no output; assertion exited 0.
- git diff --check origin/main -- . ':!.task-board': exit 0 after restoration.
- Ledger consistency: 482 rows across linux/darwin/windows, exit 0. This checks build inclusion, not platform runtime execution.
- Exploratory marker/coverage/mixed commands before the 22 trunk-only restorations also exited 0. They are retained as preliminary evidence; the named final-candidate commands above were rerun after restoration. The first receipts regex did not select acquisition table subtests; these were run explicitly afterward in acquisition-count-tables.log. The scopes/skillspec mask selected the GC consumer but no skillspec cases; the entire skillspec package was run explicitly afterward.

### Negative regression and narrowing mutant
Named existing regression: `TestExternalRepositoryAcquisitionConformance/common-fetch-argv`, through production acquisition command construction and conformancecoverage.RunOutcomes. Temporarily narrowed the new exact count from 55 to 54. Standalone command `env CURATOR_CONFORMANCE_ROOT="$TMPDIR/TASK-261001-3bsyvh_spec-pin/conformance/v1" go test -json -count=1 -p 1 -timeout 3m ./internal/buildrepo -run '^TestExternalRepositoryAcquisitionConformance$/^common-fetch-argv$'` exited 1: the family publishes 55 cases, want pinned count 54. This expected-red gate FAILED; it is not reported green. Restored the complete counts file byte-for-byte (SHA-256 42ab139a7734f67add17da39e5701c0ba436b27895171e7dd4c5f618674c9b46), reran the same command green (exit 0), and reran the full identity proof unchanged. No mutant is left in the candidate.

### Measured bounds
| Family | Driven / total | Known gaps | Bounds | Skipped |
| --- | ---: | ---: | ---: | ---: |
| Mixed build | 6/6 | 0 | 0 | 0 |
| Path/shim | 3/3 | 0 | 0 | 0 |
| Signing | 3/4 | 0 | 1 | 0 |
| Transaction | 4/4 | 0 | 0 | 0 |
| Marker v2 | 14/14 | 0 | 0 | 0 |
| Marker v3 | 22/27 | 5 | 0 | 0 |
| Marker v4 | 22/27 | 5 | 0 | 0 |
| Status/repair/GC | 1/5 | 0 | 4 | 0 |
| Acquisition argv | 55/55 | 0 | 0 | 0 |
| Acquisition clean environment | 17/17 | 0 | 0 | 0 |
| Acquisition forbidden features | 11/11 | 0 | 0 | 0 |

Production entries remain marker.Read, install.Project and Global. Signing/status/repair bounds and marker known gaps are unchanged, as described in revision 1. The acquisition 12-case table was independently counted but its behavioral cases were not rerun; the three selected acquisition tables were rerun. Full repository tests, race, complete buildrepo/install packages, and hosted platform execution were not rerun locally. The task-board handoff runs its configured remote validation and records that separate result; the older green remote run was not substituted for current local evidence. The previously attached original product mutants are retained historical evidence, not rerun claims.

No manual commit, branch switch, Git rebase/merge or integration was performed. Repository changes remain uncommitted for the developer handoff. The explicit no-LOGBOOK/CHANGELOG instruction makes those repository edits inapplicable; anomalies and decisions are recorded here and in board notes.


## Revision 3 (re-apply)

Ready for review. Applied the revision 2 delta onto fresh trunk `e87d488b8fd892bc86c037ae5e325e596df8e745` with `git diff bd126a9a 4a3d18c7 -- . ':!.task-board'` followed by `git apply --3way`: both processes exited 0, all paths applied cleanly, no conflict resolution or product edits. No new tests were authored: the accepted test delta is carried unchanged, as required by the binding re-apply brief.

### Identity and trunk context

Exactly 17 non-board paths, 978 additions and 43 deletions. Path lists, per-file numstat and per-path +/- line multisets match the original accepted delta (`5432c85f` to `refs/campaign/rjxrgs-rev1-20261001`) and revision 2 (`bd126a9a` to `4a3d18c7`) in both directions. Twelve of fourteen Go files are byte-identical to revision 2. `internal/install/install.go` retains only trunk's versioned marker/write rules and registry hash/fetch changes; `internal/marker/schema_coverage_test.go` retains only trunk's marker-v5 consumer. These context differences are independently proven by `git merge-tree --write-tree --merge-base=bd126a9a e87d488b 4a3d18c7`: exit 0, clean merge. Every non-board blob in that full merge tree equals the candidate index, including all trunk-only paths.

Candidate index tree: `9cf8a21bc3c9cb35ed61ca716d010c92403873de`. Independent merge tree: `9999351b18706f15e7f6f6c5821e9109e0842803`. Their differences are solely .task-board projections, excluded from the source comparison. Non-board blob-map SHA-256: `2b2fb4ee10b0643576dd7a9b2f9979cfa41069c3e3f4d9cd7220e178f1de39a5`. Identity script's conclusive run exited 0. Two preliminary proof-script runs exited 1: the first incorrectly required every Go blob to equal rev2 despite permitted trunk context; the second incorrectly assumed root-artifacts has a unique package key, although it permits multiple artifact declarations per package. Corrected the proof, without editing repository content, to compare full merge blobs and preserve declaration rows.

No stray untracked files or old `TASK-260728-rjxrgs_results.md`, no CHANGELOG/LOGBOOK edits, and no source additions/removals beyond the accepted line multisets. No new name-bearing content introduced. No branch switch, rebase, manual commit, integration or replay-worktree commit. Source remains uncommitted for handoff.

Stats against current trunk:

| Path | + | - |
| --- | ---: | ---: |
| `.github/ci/conformance-case-counts.tsv` | 4 | 0 |
| `.github/ci/conformance-gaps.tsv` | 5 | 0 |
| `.github/ci/root-artifacts.tsv` | 2 | 1 |
| `internal/buildrepo/buildrepo.go` | 5 | 0 |
| `internal/buildrepo/buildrepo_test.go` | 12 | 0 |
| `internal/buildrepo/pipeline.go` | 14 | 4 |
| `internal/buildrepo/pipeline_test.go` | 44 | 9 |
| `internal/buildrepo/protected.go` | 1 | 1 |
| `internal/buildrepo/receipt_v3_test.go` | 17 | 10 |
| `internal/install/external.go` | 9 | 0 |
| `internal/install/external_lifecycle_conformance_test.go` | 813 | 0 |
| `internal/install/global.go` | 9 | 9 |
| `internal/install/install.go` | 12 | 7 |
| `internal/install/stage_test.go` | 8 | 1 |
| `internal/marker/schema_coverage_test.go` | 4 | 0 |
| `internal/scopes/gc_conformance_test.go` | 16 | 1 |
| `internal/skillspec/parse.go` | 3 | 0 |

Stats versus revision 2 (every other changed path is trunk-only and matches the independent merge):

```text
1	1	.github/ci/conformance-case-counts.tsv
0	87	.github/ci/conformance-gaps.tsv
10	7	internal/config/config.go
1	1	internal/config/config_test.go
41	12	internal/config/environments.go
34	0	internal/config/environments_conformance_test.go
1	1	internal/config/environments_test.go
150	116	internal/conformancecoverage/content_hash_v2_gaps_test.go
94	14	internal/contextaudit/contextaudit.go
61	0	internal/contextaudit/contextaudit_test.go
64	11	internal/contextlock/contextlock.go
38	0	internal/contextlock/hash_version_test.go
47	0	internal/contextlock/schema_conformance_test.go
11	17	internal/contextmaterialize/contextmaterialize.go
10	0	internal/contextmaterialize/system_module_admission_test.go
19	10	internal/contextresolve/contextresolve.go
1	1	internal/contextstore/contextstore.go
25	0	internal/contextstore/contextstore_test.go
14	5	internal/envmarker/envmarker.go
4	4	internal/envmarker/envmarker_test.go
33	0	internal/envmarker/marker_env_schema_test.go
83	0	internal/envprofile/credential_link_test.go
119	0	internal/envprofile/credential_record_test.go
23	3	internal/envprofile/envprofile.go
37	0	internal/envprofile/envprofile_policy_test.go
14	0	internal/envprofile/hash_v2_test_helpers_test.go
20	3	internal/envprofile/managed.go
34	0	internal/envprofile/managed_test.go
8	1	internal/envprofile/migrate.go
60	0	internal/envprofile/migrate_test.go
175	0	internal/hashing/hashing.go
152	0	internal/hashing/hashing_v2_test.go
15	11	internal/install/install.go
48	4	internal/install/registry_e2e_test.go
8	1	internal/install/targets.go
4	0	internal/interop/environments/context_materialization_test.go
5	0	internal/interop/environments/context_resolution_test.go
157	29	internal/marker/marker.go
250	0	internal/marker/marker_hash_v2_test.go
15	0	internal/marker/schema_coverage_test.go
150	0	internal/registry/carriers.go
38	7	internal/registry/http.go
115	7	internal/registry/registry.go
103	0	internal/registry/registry_test.go
148	0	internal/registry/schema_conformance_test.go
```

### Ledger union and exact counts

The three TSV files are verified row-for-row as trunk plus the carrier's changes relative to bd126a9a. Trunk's content-hash-v2 count correction to 121 and gap removals remain intact; removed trunk gaps were not resurrected. Counts ledger has 193 manifest+family keys; gaps ledger has 37 manifest+family+case keys; root-artifacts has 30 declaration rows. Carrier adds four lifecycle count rows and five marker-v3 gap rows, one install artifact declaration and the marker-v3 artifact in the existing marker declaration. All carried rows are unchanged.

Loaded the published arrays and schema directories from the preserved workflow-pinned corpus (`23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`) and checked its manifest SHA-256: `be11bb1e4c46f21fb5684d586f9c2a8b0d59f3b437bc7ea7aa5aa530fe4d47ca`. Counts independently recomputed from this suite and asserted equal to the merged ledger:

- `external-repository-lifecycle/mixed_build_cases`: 6.
- `external-repository-lifecycle/path_shim_cases`: 3.
- `external-repository-lifecycle/signing_cases`: 4.
- `external-repository-lifecycle/transaction_cases`: 4.
- `external-repository/acquisition/cases`: 12.
- `external-repository/acquisition/clean-environment`: 17.
- `external-repository/acquisition/common-fetch-argv`: 55.
- `external-repository/acquisition/forbidden-fetch-features`: 11.
- `marker/install-marker-v2/schema-cases`: 14.
- `marker/install-marker-v3/schema-cases`: 27.
- `marker/install-marker-v4/schema-cases`: 27.

Count assertions and full blob/row identity proof exited 0. No count was guessed or rewritten.

### Validation run in this revision

All commands below were rerun locally against this candidate, each as a standalone subprocess with direct log redirection, no tee or pipelines. JSON streams, exact argv, elapsed time, root binding and real exit codes are in `TASK-261001-3bsyvh_reapply-rev3-evidence.tar.gz`. Root-enabled tests use `CURATOR_CONFORMANCE_ROOT` from `root.txt` in that archive. Historical results and mutants were not substituted for these executions.

- `go test ./internal/conformancecoverage -count=1`: exit 0 (2.575s), stream `coverage.log`.
- `go test -json -count=1 -p 1 -timeout 3m ./internal/marker`: exit 0 (4.161s), stream `marker.log`.
- `go test -json -count=1 -p 1 -timeout 3m ./internal/buildrepo -run '^(TestExternalReceiptV2CacheKeyVector|TestExternalReceipt3.*|TestExternalProtectedCache.*|TestSubstitutionCannotAliasDeclaredCacheKey|TestCollectSweepsTheReceipt3Namespace|TestPrepareNamespaces.*)$'`: exit 0 (17.921s), stream `receipts.log`.
- `go test -json -count=1 -p 1 -timeout 5m ./internal/install -run '^(TestLegacyMixedBuildProjectInstallProducesMarkerV3|TestExternalCommandNameCollisionFailsBeforeMutation|TestGlobalMixedBuildStagesExternalBeforeLocal|TestAuthoritativeMixedBuildCasesUseProjectInstallEntry|TestLegacyBuildsKeepReceipt1|TestDraftBuildsPublishReceipt3OnBothArms|TestSecondBuildFailurePreservesPriorInstallationAndLiveCache|TestSourceMutationDuringStagingBlocksHandoffOfACacheHit)$'`: exit 0 (78.617s), stream `mixed.log`.
- `go test -json -count=1 -p 1 -timeout 4m ./internal/install -run '^TestAuthoritativePathShimCasesUseProjectInstallEntry$'`: exit 0 (23.955s), stream `shim.log`.
- `go test -json -count=1 -p 1 -timeout 4m ./internal/install -run '^TestAuthoritativeSigningCasesUseProjectInstallEntry$'`: exit 0 (9.127s), stream `signing.log`.
- `go test -json -count=1 -p 1 -timeout 5m ./internal/install -run '^TestAuthoritativeTransactionCasesUseProjectInstallEntry$'`: exit 0 (23.618s), stream `Transactions.log`.
- `go test -json -count=1 -p 1 -timeout 3m ./internal/scopes -run '^TestAuthoritativeGarbageCollectionRootsAreRetained$'`: exit 0 (3.362s), stream `gc.log`.
- `go test -json -count=1 -p 1 -timeout 3m ./internal/skillspec`: exit 0 (19.331s), stream `skillspec.log`.
- `go test -json -count=1 -p 1 -timeout 3m ./internal/envprofile -run '^TestManagerOwnedAbsenceReadsAreGuarded$'`: exit 0 (17.791s), stream `absence.log`.
- `bash .github/ci/ledger-consistency.sh /tmp/TASK-261001-3bsyvh-rev3/ledger`: exit 0 (38.925s), stream `ledger.log`.
- `go build -p 1 ./...`: exit 0 (16.92s), stream `build.log`.
- `go vet -p 1 ./...`: exit 0 (14.47s), stream `vet.log`.
- `golangci-lint run --concurrency 1 --timeout 5m`: exit 0 (53.319s), stream `lint.log`.
- `gofmt -l internal/buildrepo/buildrepo.go internal/buildrepo/buildrepo_test.go internal/buildrepo/pipeline.go internal/buildrepo/pipeline_test.go internal/buildrepo/protected.go internal/buildrepo/receipt_v3_test.go internal/install/external.go internal/install/external_lifecycle_conformance_test.go internal/install/global.go internal/install/install.go internal/install/stage_test.go internal/marker/schema_coverage_test.go internal/scopes/gc_conformance_test.go internal/skillspec/parse.go`: exit 0 (0.105s), stream `format.log`.
- `git diff --check origin/main -- . ':!.task-board'`: exit 0 (0.059s), stream `diff-check.log`.

Formatting emitted no filenames. Ledger consistency passed for linux/darwin/windows build inclusion; it is not platform runtime execution. Existing conformancecoverage negative tests exercise missing classifications, vanished pinned cases, unlisted failures, stale passing gaps, and unreadable-state rejection. No new behavior or gate was added and no product mutants were rerun in this carrier revision.

### Measured published-case bounds

| Family | Driven / total | Known gaps | Bounds | Skipped |
| --- | ---: | ---: | ---: | ---: |
| Mixed build | 6/6 | 0 | 0 | 0 |
| Path/shim | 3/3 | 0 | 0 | 0 |
| Signing | 3/4 | 0 | 1 | 0 |
| Transaction | 4/4 | 0 | 0 | 0 |
| Marker v2 | 14/14 | 0 | 0 | 0 |
| Marker v3 | 22/27 | 5 | 0 | 0 |
| Marker v4 | 22/27 | 5 | 0 | 0 |
| Status/repair/GC | 1/5 | 0 | 4 | 0 |

Production entries remain `marker.Read`, `install.Project`, `Global`, and scopes collection. Signing's release-pipeline operation, status/repair facts and five shared-validator marker gaps retain the accepted bounds described above. Marker-v5 behavior belongs to trunk and is not claimed exercised under the rc.13 root. Full repository tests, full buildrepo/install packages, race tests, hosted runtime lanes, the content-hash-v2 candidate corpus, and historical product mutants were not rerun locally; this re-apply scope does not claim their success. Hosted validation is separately enforced and recorded by the handoff gate.

The prior revision was accepted; this is a freshness re-apply, with no unaddressed rejection verdict. The no-LOGBOOK/CHANGELOG instruction applies; findings and the preliminary proof assumptions are recorded in this outcome and board notes. Evidence is attached before the developer handoff.


## Revision 4 (re-apply)

Ready for review. Run RUN-261001-b06aa6 re-applies accepted revision 3 from `refs/campaign/3bsyvh-rev3-20261001` (`d540fbb1`, base `e87d488b`) onto fresh trunk `67d83539`. Exporting the non-board delta and `git apply --3way` each exited 0. All 17 paths applied cleanly. No conflict resolution, count edits, or product changes were required. The accepted tests were carried unchanged.

### Identity and base movement

Exactly 17 paths, 978 additions and 43 deletions. Per-path numstat and two-way added/removed line multisets equal the ORIGINAL accepted rjxrgs revision 1 and revision 3. No extra paths or stray `TASK-260728-rjxrgs_results.md`; no CHANGELOG/LOGBOOK edits. Identical added/removed content introduces no new name-bearing content. The required origin/main path-count command, with pipefail enabled, returned 17, exit 0.

Thirteen of fourteen Go blobs are byte-identical to revision 3. Only `internal/install/install.go` retains trunk context: the managed gitignore handling added between e87d488b and 67d83539 permits the existing non-repository notice path. That trunk patch is attached as `trunk-install-context.patch` in the evidence archive. No carrier hunk changed.

Independent `git merge-tree --write-tree --merge-base=e87d488b 67d83539 d540fbb1` exited 0, clean merge. Its entire tree equals the staged candidate, proving the base movement and all trunk-only paths. No branch switch, manual commit, rebase, integration, or replay-worktree commit was performed.

Candidate and independent merge tree: `289196d5819814cd573cd6b6cbe2b0016a8c95b2`. Non-board blob-map SHA-256: `cb5a4aa3f0bcc1b2464dba4d2ff7952b6814a2fb43b789d9c604218b2200b7bd`. Full identity assertions exited 0 before the mutant and after restoration.

Per-file stats against trunk:

| Path | + | - |
| --- | ---: | ---: |
| `.github/ci/conformance-case-counts.tsv` | 4 | 0 |
| `.github/ci/conformance-gaps.tsv` | 5 | 0 |
| `.github/ci/root-artifacts.tsv` | 2 | 1 |
| `internal/buildrepo/buildrepo.go` | 5 | 0 |
| `internal/buildrepo/buildrepo_test.go` | 12 | 0 |
| `internal/buildrepo/pipeline.go` | 14 | 4 |
| `internal/buildrepo/pipeline_test.go` | 44 | 9 |
| `internal/buildrepo/protected.go` | 1 | 1 |
| `internal/buildrepo/receipt_v3_test.go` | 17 | 10 |
| `internal/install/external.go` | 9 | 0 |
| `internal/install/external_lifecycle_conformance_test.go` | 813 | 0 |
| `internal/install/global.go` | 9 | 9 |
| `internal/install/install.go` | 12 | 7 |
| `internal/install/stage_test.go` | 8 | 1 |
| `internal/marker/schema_coverage_test.go` | 4 | 0 |
| `internal/scopes/gc_conformance_test.go` | 16 | 1 |
| `internal/skillspec/parse.go` | 3 | 0 |

Full per-file stats versus revision 3 (trunk movement, independently verified by the full-tree merge proof):

```text
106	0	.github/ci/conformance-case-counts.tsv
98	0	.github/ci/conformance-gaps.tsv
29	0	.github/ci/gate-selftest.sh
14	0	.github/ci/platform-cases.tsv
1	0	.github/ci/skip-classes.tsv
174	0	.research/261001_second-operator-requirements-answers.md
145	0	.research/261001_mandates-launch-context-advice.md
3	0	.research/TASK-261001-3qugz9_capture-provider.py
1058	0	.research/TASK-261001-3qugz9_evidence.json
64	0	.research/TASK-261001-3qugz9_probe.py
4	0	CHANGELOG.md
2	0	README.md
2	0	cmd/curator/env.go
11	13	cmd/curator/env_test.go
2	7	cmd/curator/hook_posture_test.go
2	7	cmd/curator/hook_test.go
205	0	cmd/curator/install_gitignore_test.go
477	0	cmd/curator/muse_test.go
8	0	docs/cli.md
154	0	docs/second-operator.md
5	2	internal/buildrepo/acquisition_conformance_test.go
45	5	internal/buildrepo/httpsbroker_pipe_unix.go
166	0	internal/buildrepo/httpsbroker_pipe_unix_test.go
33	11	internal/buildrepo/httpsbroker_test.go
0	4	internal/buildrepo/httpsbroker_test_pipe_unix_test.go
0	4	internal/buildrepo/httpsbroker_test_pipe_windows_test.go
215	0	internal/conformancecoverage/content_hash_v2_gaps_test.go
8	0	internal/conformancecoverage/coverage.go
1	1	internal/contextaudit/contextaudit.go
1	1	internal/contextlock/schema_conformance_test.go
9	0	internal/contextpkg/contextpkg.go
29	1	internal/envfragment/envfragment.go
47	0	internal/envfragment/muse_test.go
555	0	internal/envfragment/testdata/curator-spec-muse/schemas/v1/launch-env-fragment-v3.schema.json
21	1	internal/envprofile/global.go
67	3	internal/envprofile/managed.go
147	0	internal/envprofile/muse.go
535	0	internal/envprofile/muse_test.go
14	0	internal/envprofile/status.go
8	3	internal/envprofile/surfacing_test.go
15	0	internal/envregistry/envregistry.go
16	4	internal/envregistry/envregistry_test.go
28	11	internal/gitignore/gitignore.go
23	18	internal/gitignore/gitignore_test.go
40	0	internal/install/gitignore_spawn_test.go
8	5	internal/install/install.go
1	1	internal/marker/content_hash_v2_marker_cases_test.go
1	1	internal/registry/schema_conformance_test.go
33	0	internal/stateread/stateread.go
21	0	internal/stateread/stateread_test.go
```

### Ledger union and counts

All three TSV files equal trunk plus exactly the revision 3 delta, verified row-for-row. Counts have 296 unique manifest+family keys, gaps have 134 unique manifest+family+case keys, and root-artifacts has 30 declarations. The four added lifecycle count rows, five marker-v3 gap rows, install artifact declaration, and marker-v3 extension remain unchanged. Trunk's Muse rows and other candidate-suite rows remain intact.

Counts were independently loaded from the workflow-pinned suite at revision `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`. Manifest SHA-256 is `be11bb1e4c46f21fb5684d586f9c2a8b0d59f3b437bc7ea7aa5aa530fe4d47ca`. Lifecycle mixed/shim/signing/transaction arrays recompute as 6/3/4/4; acquisition cases/environment/argv/forbidden-feature arrays as 12/17/55/11; marker v2/v3/v4 schema families as 14/27/27. Every assertion matches the merged ledger.

Additionally, all 103 Muse family counts were recomputed from `/tmp/TASK-261001-2yvag1-spec/conformance/v1`, selected by manifest digest `bd03456b92a7368d90ea74fe6953db10bc188588020683024a6a6b8735a40783`; all matched with exit 0. Counts were measured, never guessed. The preserved content-hash-v2 rows are proven unchanged, not claimed independently recomputed in this run.

### Local validation and real exits

Each command ran directly as a standalone subprocess, without tee or pipeline. Evidence includes exact argv, root binding, elapsed time, process exit, and full streams. All root-enabled commands use `/tmp/curator-spec-rc13/conformance/v1`; its Git revision and manifest digest are recorded in suite-binding.json. Every command below was rerun in this run, not accepted from old evidence.

- `go test ./internal/conformancecoverage -count=1`: exit 0, `coverage.log`.
- `go test -json -count=1 -p 1 -timeout 3m ./internal/marker`: exit 0, `marker.log`.
- `go test -json -count=1 -p 1 -timeout 3m ./internal/buildrepo -run '^(TestExternalReceiptV2CacheKeyVector|TestExternalReceipt3.*|TestExternalProtectedCache.*|TestSubstitutionCannotAliasDeclaredCacheKey|TestCollectSweepsTheReceipt3Namespace|TestPrepareNamespaces.*)$'`: exit 0, `receipts.log`.
- `go test -json -count=1 -p 1 -timeout 5m ./internal/install -run '^(TestLegacyMixedBuildProjectInstallProducesMarkerV3|TestExternalCommandNameCollisionFailsBeforeMutation|TestGlobalMixedBuildStagesExternalBeforeLocal|TestAuthoritativeMixedBuildCasesUseProjectInstallEntry|TestLegacyBuildsKeepReceipt1|TestDraftBuildsPublishReceipt3OnBothArms|TestSecondBuildFailurePreservesPriorInstallationAndLiveCache|TestSourceMutationDuringStagingBlocksHandoffOfACacheHit)$'`: exit 0, `mixed.log`.
- `go test -json -count=1 -p 1 -timeout 4m ./internal/install -run '^TestAuthoritativePathShimCasesUseProjectInstallEntry$'`: exit 0, `shim.log`.
- `go test -json -count=1 -p 1 -timeout 4m ./internal/install -run '^TestAuthoritativeSigningCasesUseProjectInstallEntry$'`: exit 0, `signing.log`.
- `go test -json -count=1 -p 1 -timeout 5m ./internal/install -run '^TestAuthoritativeTransactionCasesUseProjectInstallEntry$'`: exit 0, `Transactions.log`.
- `go test -json -count=1 -p 1 -timeout 3m ./internal/scopes -run '^TestAuthoritativeGarbageCollectionRootsAreRetained$'`: exit 0, `gc.log`.
- `go test -json -count=1 -p 1 -timeout 3m ./internal/skillspec`: exit 0, `skillspec.log`.
- `go test -json -count=1 -p 1 -timeout 3m ./internal/envprofile -run '^TestManagerOwnedAbsenceReadsAreGuarded$'`: exit 0, `absence.log`.
- `bash .github/ci/ledger-consistency.sh /tmp/TASK-261001-3bsyvh-rev4/ledger`: exit 0, `ledger.log`.
- `go build -p 1 ./...`: exit 0, `build.log`.
- `go vet -p 1 ./...`: exit 0, `vet.log`.
- `golangci-lint run --concurrency 1 --timeout 5m`: exit 0, `lint.log`.
- `gofmt -l internal/buildrepo/buildrepo.go internal/buildrepo/buildrepo_test.go internal/buildrepo/pipeline.go internal/buildrepo/pipeline_test.go internal/buildrepo/protected.go internal/buildrepo/receipt_v3_test.go internal/install/external.go internal/install/external_lifecycle_conformance_test.go internal/install/global.go internal/install/install.go internal/install/stage_test.go internal/marker/schema_coverage_test.go internal/scopes/gc_conformance_test.go internal/skillspec/parse.go`: exit 0, `format.log`.
- `git diff --check origin/main -- . ':!.task-board'`: exit 0, `diff-check.log`.
- `go test -json -count=1 -p 1 -timeout 5m ./internal/install -run '^TestAuthoritativeMixedBuildCasesUseProjectInstallEntry$'`: exit 0, `restored.log`.

Formatting emitted no paths. Ledger consistency verifies build inclusion on linux/darwin/windows, not hosted runtime execution.

### Named regression and narrowing mutant

The required previous verdict resource `TASK-261001-3bsyvh_review-verdict-rev3.md` explicitly says ACCEPTED. There is no rejection finding; the current instruction is a freshness re-apply with no content change.

Named carried regression: `TestAuthoritativeMixedBuildCasesUseProjectInstallEntry`. It drives `install.Project` and observes its marker and receipts; published-vector tally checking is called by `conformancecoverage.RunOutcomes` at `internal/install/external_lifecycle_conformance_test.go:397`.

Temporarily narrowed the rc.13 mixed_build_cases exact-count pin from 6 to 5 while retaining the test and every production gate. The same named test ran with -count=1 and exited 1, reporting `want pinned count 5` for the six published cases. This expected-red mutant FAILED; it is not a passing gate. The counts file was restored in a finally block from an external byte copy. SHA-256 and byte identity were checked; the same command then exited 0. Identity and whitespace checks were rerun after restoration. No mutant remains.

This is a count-ratchet proof, not a claim to have rerun the historical product-gate mutants. Existing refusal tests for receipt evidence, command collisions, staging rollback, source mutation, and coverage policy were rerun unchanged.

### Measured bounds and unrun checks

| Family | Driven / total | Known gaps | Bounds | Skipped |
| --- | ---: | ---: | ---: | ---: |
| Mixed build | 6/6 | 0 | 0 | 0 |
| Path/shim | 3/3 | 0 | 0 | 0 |
| Signing | 3/4 | 0 | 1 | 0 |
| Transaction | 4/4 | 0 | 0 | 0 |
| Marker v2 | 14/14 | 0 | 0 | 0 |
| Marker v3 | 22/27 | 5 | 0 | 0 |
| Marker v4 | 22/27 | 5 | 0 | 0 |
| Status/repair/GC | 1/5 | 0 | 4 | 0 |

These are the unchanged accepted bounds: signing release-pipeline operation has no install.Project entry; status/repair lacks the external facts or repair entry; five shared-validator marker gaps remain ledger-owned. Raw published tallies are attached. Full repository tests, full buildrepo/install packages, race tests, hosted runtime lanes, and content-hash-v2/Muse behavior suites were not rerun locally: this carrier runs the required affected rows and static/count checks, and does not claim broader suite success. Handoff validation is separate evidence.

No new tests were authored because the binding brief requires unchanged content; the normative accepted tests are preserved. The explicit prohibition on CHANGELOG/LOGBOOK edits applies. Findings are preserved in this outcome and board notes. Source remains uncommitted for the developer handoff snapshot. Outcomes are attached before role handoff.

Evidence packaging note: the first report generation exited 1 because the adapted identity writer retained its revision 3 filename. The first results-update and identity-add commands also exited 1 with missing-file errors; the initial evidence archive attached successfully but lacked the generated report. Corrected the external artifact filename, reran identity and report generation (both exit 0), and replaced the archive through resource update. No repository content changed and no validation result was affected.
