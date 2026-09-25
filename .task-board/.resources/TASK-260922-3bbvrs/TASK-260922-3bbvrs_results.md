# TASK-260922-3bbvrs results

## Changes

- Added `internal/conformancecoverage`: one runner for driven, known-gap, bound, and skipped outcomes; committed family count pins; one ledger loader; exact tally, duplicate, missing, extra, stale-row, unlisted-failure, and passing-gap checks.
- Migrated complete published-case consumers across manager/system config, draft-sources, marker schemas, skillspec and environment schemas, build drivers, source identities, module roots, manager lifecycle, environment profile, shell trust, closure/GC, and interop/context vectors. The semantic draft-sources matrix uses four parallel workers; every case still runs and the parent checks the complete tally after all case tests finish.
- Added `.github/ci/conformance-case-counts.tsv` for the committed curator-spec rc.12 root (`dced9b8317e0e8af79edf2d0539b32bd22b6c85b`) and the separately pinned draft-sources corpus (`802caee548ddc8b19408746d26c7972d39b39cc2`). `SPEC_PIN` was not changed.
- Added five marker v4 schema gaps to `.github/ci/conformance-gaps.tsv`, each owned by `BUG-260923-2afgyq`, with a one-line reason. This bug tracks the five external identity/kind/revision checks currently admitted by `marker.Read`.
- Added the owed-implementation rule to `docs/ci-gates.md` and `CHANGELOG.md`; updated root-artifact declarations for the newly covered root consumers. No product behavior changed.

Creating `BUG-260923-2afgyq` under `STORY-260822-2lvw0e` automatically reopened that story and `EPIC-260822-18ylpq` from done to backlog. This records the outstanding implementation work for the five ledger rows; the board transition was left intact.

## Validation

All commands below were run as standalone processes. Commands using `CURATOR_CONFORMANCE_ROOT` used:

`/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-260922-2goxjs/worktree/.temp/conformance-spec-pinned/conformance/v1`

The temporary checkout was removed after validation.

| Command | Exit | Result |
| --- | ---: | --- |
| `go test -count=1 ./internal/conformancecoverage` | 0 | Coverage fixtures and committed policy parse pass. |
| `go test -count=1 -parallel=4 -run '^TestDraftSourcesSemanticCases$' ./internal/crossconformance` | 0 | All 94 semantic cases; 347.338s. |
| `go test -count=1 -run '^TestDraftSourcesSchemaCases$' ./internal/crossconformance` | 0 | All 115 schema cases. |
| `go test -count=1 -run '^TestDraftSourcesSnapshotVectors$' ./internal/crossconformance` | 0 | All 3 snapshot vectors. |
| `go test -count=1 -run '^(TestDraftSourcesPin|TestDraftSourcesCorpusCounts|TestDraftSourcesSemanticCoverage)$' ./internal/crossconformance` | 0 | Corpus pin, counts, and driver matrix. |
| `CURATOR_CONFORMANCE_ROOT=<pinned-root> go test -count=1 ./internal/config ./internal/marker ./internal/skillspec ./internal/envmarker ./internal/envfragment ./internal/scriptpolicy ./internal/shell ./internal/moduleroots` | 0 | Config, marker, schema, script, shell, and module-root groups. |
| `CURATOR_CONFORMANCE_ROOT=<pinned-root> go test -count=1 ./internal/interop ./internal/interop/environments ./internal/contextmaterialize ./internal/closure ./internal/runtimestore` | 0 | Interop, environment, materialization, closure, and lifecycle groups. |
| `CURATOR_CONFORMANCE_ROOT=<pinned-root> go test -count=1 -run '^TestEnvironmentsEnvPassthroughVectors$' ./internal/envprofile` | 0 | Full environment pass-through family. |
| `CURATOR_CONFORMANCE_ROOT=<pinned-root> go test -count=1 ./internal/buildsource ./internal/buildcache ./internal/scopes` | 0 | Build-source, cache rejection, and GC families. |
| `CURATOR_CONFORMANCE_ROOT=<pinned-root> go test -count=1 -run '^(TestValidPackageGraphVectors|TestFixedEnvironmentAndFiveDirectArgvFormsVector|TestToolchainIdentityVectors|TestModuleRootVectorsDriveTheWholeBuild|TestCapabilityEvidenceCasesMatchTheAcceptedVector|TestIdentityProtocolAndPackageInfluenceCodesMatchTheAcceptedVector)$' ./internal/godriver` | 0 | Full positive, argv, fixed-environment, toolchain, module-root, and host-policy case groups. |
| `CURATOR_CONFORMANCE_ROOT=<pinned-root> go test -count=1 -run '^TestReleasedSourceIdentityVectors$' ./internal/buildrepo` | 0 | All 10 source-identity vectors. |
| `CURATOR_CONFORMANCE_ROOT=<pinned-root> go test -count=1 -run '^(TestAuthoritativeDryRunCasesMutateNothingPersistent|TestAuthoritativeCacheOutcomesDriveInstallation|TestAuthoritativeCacheRejectionsAreRebuiltNeverAdopted)$' ./internal/install` | 0 | Dry-run, cache-positive, and all rejection cases; 134.133s. |
| `go build ./...` | 0 | Repository builds. |
| `golangci-lint run` | 0 | 0 issues. The linter printed a non-fatal stale-path warning for a removed neighboring worktree. |
| `git diff --check` | 0 | No whitespace errors. |
| `gofmt -l internal` | 0 | No files listed. |

### Narrowing mutants

Each mutant was temporary and reverted immediately after its standalone test run.

| Mutant removed | Named test | Exit | Expected failure |
| --- | --- | ---: | --- |
| `ValidateTally` call | `TestPublishedCaseTallyRejectsMissingClassification` | 1 | A missing result returned no error. |
| Refusal for an unlisted failure | `TestUnlistedFailingCaseIsRejected` | 1 | An unlisted failing case returned no error. |
| Refusal when a listed gap passes | `TestPassingKnownGapIsRejected` | 1 | The stale ledger row returned no error. |

`TestVanishedGapCaseIsRejected` and `TestPinnedCountRejectsVanishedPublishedCase` pass in the helper suite and cover vanished ledger and published cases.

### Corrected intermediate attempts

- The initial config group used a relative `CURATOR_CONFORMANCE_ROOT` and exited 1 because Go package working directories resolved it incorrectly; rerunning with the absolute pinned path exited 0.
- The first interop/materialization group exited 1 on a compile error and a selected-case bookkeeping error; after fixes, the narrowed rerun exited 0.
- The first combined env-profile/build-source/cache/GC run was interrupted with exit 1 after surfacing a build-cache compile error. The cache issue was fixed; the env-profile family and build-source/cache/GC groups then passed in bounded commands.
- The first install vector run exited 1 on a stale import; removing it and rerunning the selected install group exited 0.
- The first serial semantic matrix exited 1 at Go's 10-minute timeout. The parallel case runner was added for this matrix only; rerunning all 94 cases exited 0 in 347.338s.
- The first lint run exited 1 on unchecked file closes, missing trusted-path annotations, and an unused helper. Those were fixed; the final lint run exited 0.

The hosted landing gate was not run locally. The handoff runtime owns its single hosted CI run. Work remains uncommitted in the Story worktree for handoff.

## Revision 2 — platform-case gate rework

- Folded the identity assertions into `TestBuildSourceConformanceVectors`, so every build-source vector now enters the same counted consumer and reaches its production assertion. Accepted record vectors also verify their published digest. The previous “outside the accepted non-empty record identity assertion” skip is gone; no skip class or product behavior changed.
- On this Darwin host, the `invalid-unicode-build-source-path` and `duplicate-build-source-path` assertions remain host-capability skips because the filesystem cannot represent those inputs. The consumer reported `8 driven, 0 known-gap, 0 bound, 2 skipped, 10 total`; the platform gate recorded both skips as `host-capability` and passed.

### Revision 2 validation

Commands were run as standalone processes. The rc.12 root for Go tests was materialized under the worktree from the pinned curator-spec commit `dced9b8317e0e8af79edf2d0539b32bd22b6c85b` and removed after validation.

| Command | Exit | Result |
| --- | ---: | --- |
| `CURATOR_CONFORMANCE_ROOT=<pinned-root> go test ./internal/buildsource/... ./internal/crossconformance/...` before the final buildsource test refactor | 0 | Buildsource 1.132s; crossconformance executed and passed in 416.851s. |
| Same command after the final refactor | 0 | Buildsource reran and passed in 1.011s; Go reported crossconformance `(cached)` from the preceding direct successful run. |
| `CURATOR_CONFORMANCE_ROOT=<pinned-root> go test -json -count=1 -run '^TestBuildSourceConformanceVectors$' ./internal/buildsource` | 0 | Final buildsource consumer produced the 8/0/0/2 tally above. |
| `CI_PLATFORM_CASES=<task-scoped-ledger> CI_GATE_GOOS=darwin bash .github/ci/platform-case-gate.sh <go-test-json> <evidence-dir>` | 0 | Both host-capability skips were recorded; no unrecognised skip remained. |
| `sh .github/ci/gate-selftest.sh` | 0 | 185 passed, 0 failed, including the unrecognised-skip refusal. |
| `sh .github/ci/ledger-consistency.sh` | 2 | The bare invocation prints usage because this script requires an evidence-directory argument. |
| `sh .github/ci/ledger-consistency.sh <task-scoped-evidence-dir>` | 0 | 241 platform-case rows checked across Linux, Darwin, and Windows. |
| `sh .github/ci/no-broad-suppression.sh` | 0 | Narrow-suppression policy passes. |
| `go test ./internal/conformancecoverage` | 0 | Count and ledger fixture tests pass. |
| `go build ./...` | 0 | Repository builds. |
| `golangci-lint run` | 0 | 0 issues; the runner emitted a non-fatal warning for an already-removed neighboring temporary worktree. |
| `git diff --check` | 0 | No whitespace errors. |
| `gofmt -l internal` | 0 | No files listed. |

The three narrowing-mutant results recorded above remain applicable: the revision changes only build-source vector routing, while `conformancecoverage.Check` and its named negative tests are unchanged. They were not rerun during this platform-specific rework. The hosted landing suite remains assigned to the handoff runtime and was not run locally. `SPEC_PIN` remains `dced9b8317e0e8af79edf2d0539b32bd22b6c85b`.

## Revision 3 — complete published-case consumer sweep

- Routed the missed `skillfile-dev-v2` directory consumer through `RunOutcomes` as `TestSkillfileDevV2PublishedCasesUseCoverageRatchet`, then proved its count pin is active with a temporary 16-to-15 narrowing mutant.
- The full-list grep also found `TestReleasedSkillBuildSchemaCases`; rc.12 publishes 13 `skill-build-v1` files, and this consumer now uses the shared harness too.
- Routed all four complete `external-repository` JSON lists in `internal/buildrepo/admission_test.go`: raw objects (12), LFS pointers (10), pack index (8), and local config/refs (15). Their counts are pinned in `.github/ci/conformance-case-counts.tsv`; `.github/ci/root-artifacts.tsv` now requires those files, `skill-build-v1`, and `skillfile-dev-v2` for the packages that consume them.
- Measured tallies at the unchanged rc.12 pin (`dced9b8317e0e8af79edf2d0539b32bd22b6c85b`):

  | Family | Pin | Driven | Known gap | Bound | Skipped | Total |
  | --- | ---: | ---: | ---: | ---: | ---: | ---: |
  | `skillfile-dev-v2/schema-cases` | 16 | 16 | 0 | 0 | 0 | 16 |
  | `skill-build-v1/schema-cases` | 13 | 13 | 0 | 0 | 0 | 13 |
  | `external-repository/raw-objects` | 12 | 12 | 0 | 0 | 0 | 12 |
  | `external-repository/lfs-pointers` | 10 | 10 | 0 | 0 | 0 | 10 |
  | `external-repository/pack-index` | 8 | 7 | 0 | 1 | 0 | 8 |
  | `external-repository/local-config-and-refs` | 15 | 13 | 0 | 2 | 0 | 15 |

- The pack-index bound is `reject-pack-without-index`: `validatePackIndex` requires a pack/index pair. The local config/ref bounds are `reject-gitfile` and `reject-link-or-special-administration-file`, which publish no config bytes for `parseLocalConfig`. Both reasons are emitted by the harness; neither is a Go test skip. The gap ledger is unchanged and still owns the five existing marker gaps; the six newly counted families have no known-gap rows.
- Consumer sweep commands used: `rg -n --glob '*_test.go' 'range .*\.(Cases|Vectors|PositiveCases|RejectionCases|SemanticCases|SchemaCases|SnapshotCases|MaterializationCases)|for .*range entries' internal` and `rg -n 'CURATOR_CONFORMANCE_ROOT' internal --glob '*_test.go'`. The buildmeta, whitelist, and skillcheck matches select named cases from broader vectors and remain targeted lookups as instructed. `TestSystemModuleSchemaSubset` is likewise a selected subset; the full manager/system config family consumers retain their family pins.
- No product behavior or `SPEC_PIN` changed. `docs/ci-gates.md` and `CHANGELOG.md` name the new consumers and retain the rule that every ledger row is owed implementation work, never an accepted deviation.

### Revision 3 validation

Commands were run directly as standalone processes. Tests that read the release root used curator-spec commit `dced9b8317e0e8af79edf2d0539b32bd22b6c85b`.

| Command | Exit | Result |
| --- | ---: | --- |
| `CURATOR_CONFORMANCE_ROOT=<pinned-root> go test -count=1 ./internal/devsub` | 0 | Full package passes with the pinned root. |
| `CURATOR_CONFORMANCE_ROOT=<pinned-root> go test -count=1 ./internal/buildrepo` | 0 | Full package passes with the pinned root (163.177s). |
| `CURATOR_CONFORMANCE_ROOT=<pinned-root> go test -v -count=1 -run '^TestSkillfileDevV2PublishedCasesUseCoverageRatchet$' ./internal/devsub` | 0 | 16/16 cases driven. |
| `CURATOR_CONFORMANCE_ROOT=<pinned-root> go test -v -count=1 -run '^(TestReleasedSkillBuildSchemaCases|TestRawObjectAndLFSPinnedConformanceFixtures|TestLocalConfigAndAdministrationAdversarialBoundaries|TestPackIndexConformanceAndExactSSHWrapper)$' ./internal/buildrepo` | 0 | The skill-build, raw-object, LFS, pack-index, and local-config/ref tallies above. |
| `CURATOR_CONFORMANCE_ROOT=<pinned-root> go test -v -count=1 -run '^TestLocalConfigAndAdministrationAdversarialBoundaries$' ./internal/buildrepo` | 0 | Rechecked the final two-bound reason; 13 driven + 2 bound = 15. |
| `go test -count=1 ./internal/conformancecoverage` | 0 | Harness, policy parser, and ratchet fixtures pass. |
| `CI_REQUIRE_FULL_ROOT=1 bash .github/ci/suite-plan.sh <pinned-root> <task-evidence-dir>` | 0 | 77 served, 0 deferred, 0 excluded. |
| `sh .github/ci/gate-selftest.sh` | 0 | 185 passed, 0 failed. |
| `sh .github/ci/no-broad-suppression.sh` | 0 | Suppression policy passes. |
| `go build ./...` | 0 | Repository build passes. |
| `golangci-lint run` | 0 | 0 issues. The runner emitted its non-fatal stale-path warning for `/tmp/rev-2f3xbf-clone`. |
| `gofmt -l internal/devsub/repository_test.go internal/buildrepo/buildrepo_test.go internal/buildrepo/admission_test.go` | 0 | No files listed. |
| `git diff --check` | 0 | No whitespace errors. |

### Revision 3 narrowing mutant

Temporarily changed the `skillfile-dev-v2/schema-cases` pin from 16 to 15, then ran `CURATOR_CONFORMANCE_ROOT=<pinned-root> go test -count=1 -run '^TestSkillfileDevV2PublishedCasesUseCoverageRatchet$' ./internal/devsub`: exit 1, killed by `publishes 16 cases, want pinned count 15`. Restored the count file from a saved copy (`cmp` exit 0) and reran the named test (`-count=1`, exit 0). This narrows a newly routed production-entry consumer's pin and proves the regression test catches it.

The revision-2 harness mutants (missing tally refusal, unlisted-failure refusal, and stale passing-gap refusal) remain the previously attached and reviewer-verified evidence; they were not rerun in revision 3 per the binding rework note. The single hosted landing gate remains for handoff and was not run locally.

The handoff checklist's implementation, architecture, and test items were checked against the evidence above. Its review-rejection item is conditional: no reviewer verdict exists at this handoff, so that branch has not been triggered. If review requests changes, the verdict evidence and explicit routing still apply before the next handoff.

### Corrected revision-3 attempts

- The first selected buildrepo run exited 1 because the new `skill-build-v1` count was entered as 14. The pinned root contains 13 cases; after setting the pin to 13, the same family run exited 0 and the full buildrepo package run exited 0.
- The first lint run exited 1 on two unused test callback parameters. They were changed to `_`; the final lint run exited 0 with 0 issues.

## Revision 4 — route the `cmd/curator` published families

Revision 3's verdict, `TASK-260922-3bbvrs_review-verdict-rev3.md`, identified F1: the earlier sweep stopped at `internal/`. This revision routes the three uncovered `cmd/curator` families through the same counted runner:

- `umbrella-provider-resolution/cases` uses `conformancecoverage.Run` at the `resolveProvider` production-entry vector test; pin: 14.
- `manager-lifecycle/bootstrap-cases` uses `conformancecoverage.Run`; pin: 3.
- `manager-lifecycle/upgrade-cases` uses `conformancecoverage.Run`; pin: 3.

The `cmd/curator` root-artifact row requires `vectors/umbrella-provider-resolution.json` and `vectors/manager-lifecycle.json`, so a root missing either artifact defers the package and fails a full-root lane. The root remains pinned at `dced9b8317e0e8af79edf2d0539b32bd22b6c85b`; no product behavior or `SPEC_PIN` changed. New families have no ledger gaps.

### Whole-module sweep and hit dispositions

The required regex as pasted in the rework note has an unmatched opening parenthesis (`range (entries|family\.`); that exact `rg` invocation exited 2 with `unclosed group`. I reran it with the group closed and `family\.Cases` explicit, across `.`. That corrected whole-module sweep exited 0. The supplemental sweep also exited 0:

```text
rg -n --glob '*_test.go' 'range [A-Za-z_.]*\.[A-Z][A-Za-z]*Cases\b|range (entries|family\.Cases)' .
rg -n --glob '*_test.go' 'range .*(Cases|Vectors|PositiveCases|RejectionCases|SemanticCases|SchemaCases|SnapshotCases|MaterializationCases|BootstrapCases|UpgradeCases)|range (entries|family\.)' .
```

The three F1 loops no longer appear as direct range loops: each now enters `conformancecoverage.Run`. The corrected sweep's remaining hits have these dispositions; each location below is from the search output.

**Whole published families and schema-list adapters already counted by the harness:**

- `internal/contextmaterialize/system_module_admission_test.go:119` — the complete `environments/materialization-cases` family enters `RunOutcomes`; non-target surfaces are explicitly `bound`. Its supplemental `:123` named selection only defines the production-entry subset.
- `internal/install/dryrun_conformance_test.go:834` and `internal/buildcache/builddriver_rejection_conformance_test.go:264` — full published families enter `Run`/`RunOutcomes` (`install/dry-run-cases` and `build-drivers/rejection-cases`). The latter was found by the supplemental pattern.
- `internal/config/environments_conformance_test.go:67`, `internal/devsub/repository_test.go:52`, `internal/envmarker/marker_env_schema_test.go:41`, `internal/envfragment/fragment_schema_test.go:301`, `internal/marker/schema_coverage_test.go:31`, `internal/buildrepo/buildrepo_test.go:33`, and `internal/skillspec/conformance_test.go:70` — schema-index or directory entries are adapted into the case slices passed to `Run`/`RunOutcomes`.
- `cmd/curator/umbrella_conformance_test.go:98` — the raw JSON case loop only pairs revision-member-presence bits with the decoded slice and checks equal lengths; production assertions and the 14-case tally run through `Run` at `:80`.
- `internal/crossconformance/draftsources_corpus_test.go:182` — corpus index integrity and file-existence check. The production schema and semantic consumers have their own pins and counted runners.
- `internal/config/system_module_schema_test.go:62,67,78` — index lookup and a fixed seven-case E2 subset check, not a whole-family consumer. The full manager/system config schema families are counted in `internal/config/environments_conformance_test.go`.

**Targeted selectors, host-specific selectors, or fixed subsets (not whole-family consumers):**

- Build-driver named-case selectors: `internal/whitelist/builddriver_context_conformance_test.go:45`, `internal/godriver/builddriver_rejection_conformance_test.go:49`, `internal/skillcheck/builddriver_context_conformance_test.go:45`, `internal/buildcache/builddriver_positive_conformance_test.go:45`, `internal/buildmeta/builddriver_policy_conformance_test.go:118,150`, `internal/buildsource/builddriver_conformance_test.go:88`, and `internal/skillspec/builddriver_conformance_test.go:51,318`. These look up a named case or owned boundary; the earlier review verdict explicitly found these selectors were not whole-list production consumers. The full positive/rejection families are covered by their counted consumer(s).
- `internal/godriver/builddriver_positive_conformance_test.go:238` — scans fixed-environment cases to select the one matching the native host; it does not drive the full list on one host.
- `internal/contextmaterialize/system_module_admission_test.go:123` — fixed names that define the selected production seam; all published materialization cases still enter the counted family at `:119`.
- `internal/config/system_module_schema_test.go:67,78` — the same seven-case E2 subset is selected from the schema index; the whole families are covered separately as listed above.

**Remaining generic `range entries` matches are local file listings, fixture builders, helper parameters, or state snapshots, not release-root published case lists:**

- `cmd/curator/draft_transport_provenance_test.go:343`, `cmd/curator/main_test.go:268`, `cmd/curator/envalias_test.go:113`, `cmd/curator/status_test.go:425,441,1049,1151`, and `cmd/curator/draft_transport_test.go:275,281,294,300` — temporary install/cache directory state or local JSON-policy fixture serialization.
- `internal/runtimestore/launcher_conformance_test.go:344`, `internal/godriver/session_test.go:523`, `internal/godriver/builddriver_positive_conformance_test.go:415`, `internal/godriver/build_test.go:275`, `internal/closure/refresh_atomicity_test.go:110`, `internal/sourcelock/restore_test.go:33`, `internal/buildrepo/local_release_test.go:31`, `internal/buildcache/compensation_test.go:103`, `internal/buildcache/cache_test.go:322`, `internal/buildcache/sourceaware_test.go:31`, `internal/buildsource/builddriver_conformance_test.go:323`, and `internal/marker/marker_test.go:161` — helper membership, temporary directory residue, host filename checks, or test-owned filesystem snapshots.
- `internal/gitcred/gitcred_test.go:524,571,586,599`, `internal/crossconformance/draftsources_semantic_selection_test.go:139`, `internal/crossconformance/draftsources_semantic_boundary_test.go:185,208`, `internal/gitops/deadlock_test.go:63,72`, `internal/gitops/byteexact_test.go:268,306`, and `internal/interop/environments/contract_test.go:46` — environment/member collections, temporary tree handling, literal test fixtures, or enumerating Go source files for an AST contract.
- `internal/install/stage_boundaries_test.go:186,504`, `internal/install/revalidation_test.go:110`, `internal/install/draftbuild_test.go:158`, `internal/install/stage_test.go:1670`, `internal/install/draftatomic_test.go:140`, `internal/closureexec/closureexec_test.go:561,662`, `internal/transaction/boundaries_test.go:346,441,551,759`, `internal/install/draftpublish_test.go:590,656`, `internal/install/commit_test.go:86`, `internal/install/private_test.go:38`, `internal/install/drafttransport_provenance_test.go:48`, and `internal/install/atomicity/fixture_test.go:267,398` — expected directory contents, copied fixtures, or cleanup assertions after install/transaction operations.
- `internal/artifactpolicy/reviewer_latest_rework_test.go:424`, `internal/artifactpolicy/test_helpers_test.go:159,183`, `internal/artifactpolicy/containers_test.go:493,513`, `internal/install/drafttransport_test.go:519`, and `internal/transaction/entry_test.go:192` — archive/transport fixture construction or temporary transaction-directory contents.
- `internal/buildrepo/buildrepo_test.go:33`, `internal/devsub/repository_test.go:52`, `internal/config/environments_conformance_test.go:67`, `internal/envmarker/marker_env_schema_test.go:41`, `internal/envfragment/fragment_schema_test.go:301`, `internal/marker/schema_coverage_test.go:31`, and `internal/skillspec/conformance_test.go:70` are the exceptions to the generic `entries` classification: they are schema-case adapters and are listed above as inputs to counted consumers.

The supplemental pattern also found local in-code matrices and shared test fixtures: `internal/snapshot/capture_test.go:151,191` normative snapshot fixtures; `internal/envprofile/envprofile_f13f14_test.go:229,253,283` install-operand table; `internal/runtimestore/conformance_test.go:45` a local expected-name set; `internal/sourcelock/bindings_test.go:181` local malformed bindings; `internal/buildmeta/buildmeta_test.go:217,242` receipt/input fixtures; `internal/crossconformance/rejection_test.go:25`, `export_test.go:102`, and `guard_test.go:121` the internal rejection vocabulary; `internal/crossconformance/suite_test.go:283,301` the code-owned artifact-policy matrix and its filtered deny subset; `internal/marker/marker_v5_git_test.go:135` local write cases; `internal/audit/sourceaudit_test.go:1556` a local object table; `internal/yarnclassicsource/conformance_test.go:434,452` local lock/config matrices; `internal/transaction/preparation_durability_test.go:11,60` local durability fixtures; `internal/nodesource/python_protocol_shared_test.go:122` the internal shared-protocol testdata corpus; and `internal/closuregraph/codec_test.go:121,151` local node/edge codec cases. These do not read the pinned release case families. `internal/buildcache/builddriver_rejection_conformance_test.go:264` is the published-family exception discovered by this broader pattern and is already routed through `RunOutcomes`.

### Revision 4 validation

All commands were standalone processes. Go tests that read the release root used the existing checkout of curator-spec at `dced9b8317e0e8af79edf2d0539b32bd22b6c85b` under `.temp/TASK-260922-3bbvrs-spec/conformance/v1`.

| Command | Exit | Result |
| --- | ---: | --- |
| `CURATOR_CONFORMANCE_ROOT=<pinned-root> go test -v -count=1 -run 'Umbrella|Lifecycle' ./cmd/curator` | 0 | Umbrella, bootstrap, and upgrade tests pass; tallies are 14/14, 3/3, and 3/3, all driven. |
| `CURATOR_CONFORMANCE_ROOT=<pinned-root> go test -v -count=1 -run '^TestUmbrellaProviderResolutionVectors$' ./cmd/curator` with a temporary `family.Cases[1:]` narrowing mutant | 1 | Killed: 13 published cases, want pinned count 14. |
| Same named umbrella test after restoring `family.Cases` | 0 | Restored 14-case family passes. |
| `go test -count=1 ./internal/conformancecoverage` | 0 | Shared harness and named ratchet tests pass. |
| `CI_REQUIRE_FULL_ROOT=1 bash .github/ci/suite-plan.sh <pinned-root> <revision4-evidence-dir>` | 0 | 77 served, 0 deferred, 0 excluded; confirms the new `cmd/curator` artifact row. |
| `sh .github/ci/gate-selftest.sh` | 0 | 185 passed, 0 failed, including suite-plan, ledger-consistency, root-artifact, and skip-gate cases. |
| `sh .github/ci/no-broad-suppression.sh` | 0 | Suppression gate passes. |
| `go build ./...` | 0 | Repository builds. |
| `golangci-lint run` | 0 | 0 issues. |
| `gofmt -l cmd/curator/umbrella_conformance_test.go cmd/curator/lifecycle_conformance_test.go` | 0 | No files listed. |
| `git diff --check` | 0 | No whitespace errors. |

The first `Umbrella|Lifecycle` run exited 0 but, because the pre-existing lifecycle test names did not contain `Lifecycle`, selected only umbrella tests. I renamed those two test functions to include `Lifecycle` and reran the same filter; the passing result above includes all three newly routed families. The supplied malformed `rg` pattern's exit 2 is recorded above; its corrected whole-module equivalent and supplemental sweep both exited 0.

The three shared-harness narrowing-mutant results from Revision 2 remain applicable and attached; this round added and killed the requested mutant on a newly routed umbrella family. No hosted landing gate was run locally. The worktree remains uncommitted for developer handoff.
