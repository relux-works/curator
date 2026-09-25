# TASK-260924-1aa9wb results

## Scope and decisions

The attached `1aa9wb-brief-v2.md` and `skillfile-default-on-handoff-20260924.md` supersede the original checklist wording: `CURATOR_DRAFT_SOURCES_V1`, `EnvDraftSourcesV1`, `DraftSourcesEnabled`, and all `DraftSourcesV1` policy/options are removed. There is no opt-out. The only remaining occurrence of `CURATOR_DRAFT_SOURCES_V1` in `cmd/`, `internal/`, `docs/`, or `README.md` is the test row that sets it to `0` and proves it is ignored.

Schema 2 is the default project parser and install/status/resolution path. Schema 1 keeps its read-only project meaning, and there is no implicit on-disk migration. Global Skillfile scope still requires schema 1. Environment profile path sources remain under the existing environments profile lock/context-store contract; the project Skillfile local snapshot protocol is not applied to profile scope.

README, CLI/troubleshooting docs, workflow help and CHANGELOG now describe schema 2 as the default. The source-expansion and transport docs were retitled to describe the supported workflow; the separate external build transport gate remains documented. Added a real CLI test for canonical `repository` sources using a local bare repository and operator policy, covering resolve, install, refresh and reinstall.

## Verification

All commands below were run directly as standalone processes. Exit codes are reported as observed.

- `env -u CURATOR_DRAFT_SOURCES_V1 go test ./cmd/curator -run '^(TestProjectResolveLocalCreatesLockThroughCLI|TestProjectResolveSchema1KeepsReadOnlyMeaning|TestProjectResolveIgnoresRemovedSwitch|TestProjectResolveGitTagDiffersFromHEADThroughCLI|TestProjectResolveCanonicalRepositorySourceThroughCLI|TestDraftMachinePolicySetupThroughCLI|TestProjectRefreshGitBranchMembershipChangeThroughCLI|TestDraftProjectResolveHelp|TestDraftInstallStatusHelpIncludesSchema2Workflow)$' -count=1` — exit 0.
- `env -u CURATOR_DRAFT_SOURCES_V1 go test ./cmd/curator -run '^TestGlobalStatusRejectsSchema2SkillfileThroughCLI$' -count=1` — exit 0.
- `env -u CURATOR_DRAFT_SOURCES_V1 go test ./cmd/curator -run '^(TestDraftDocsPinExamples|TestDraftDocumentedLocalShapesThroughCLI|TestDraftDocumentedGitRevisionThroughCLI)$' -count=1` — exit 0.
- `env -u CURATOR_DRAFT_SOURCES_V1 go test ./internal/manifest` — exit 0.
- `env -u CURATOR_DRAFT_SOURCES_V1 go test ./internal/install -run '^(TestDraftInstallMissingLockFails|TestDraftInstallStaleLockFails|TestDraftInstallMissingSnapshotFails|TestDraftInstallUsesPinnedBytes|TestDraftInstallGitPinnedSubtree|TestSchema2InstallRequiresLockAndLegacyStillWorks|TestGlobalInstallRejectsSchema2Skillfile)$' -count=1` — exit 0.
- `env -u CURATOR_DRAFT_SOURCES_V1 go test ./internal/crossconformance -run 'TestDraftSources(Pin|CorpusCounts|SchemaCases|SemanticCoverage|SnapshotVectors|CLILocalSkillScriptDependencies|CLIProjectPathWithSpaceAndUnicode|CLISymlinkMemberRefused|BrokerAskpassDispatch|CrossCompile|CLIInstallRestoresPriorState)$' -count=1` — exit 0.
- `env -u CURATOR_DRAFT_SOURCES_V1 go test ./internal/crossconformance -run '^TestAcceptedCorpusSatisfiesEveryPublishedStructuralClaim$' -count=1` — exit 0.
- `env -u CURATOR_DRAFT_SOURCES_V1 go test ./internal/envprofile -run '^(TestPathInstallAdmitsOrdinaryDirectory|TestLegacyPathInstallIgnoresStoreOverlap|TestPathOverlayAdmitsOrdinaryDirectory|TestPathInstallCapturesDirtyUntrackedInsideGit|TestNestedGitIsSourceInvalid|TestSymlinkInPathIsSourceInvalid|TestUnreadablePathIsUnreadable)$' -count=1` — exit 0.
- `env -u CURATOR_DRAFT_SOURCES_V1 go test ./internal/envprofile -run '^TestReinstallEmitsSurfacingBeforePublication$' -count=1` — exit 0.
- `golangci-lint run` — exit 0 (`0 issues`).
- Base binary from archived `HEAD`: `go build -buildvcs=false -o ../output/base ./cmd/curator` (run in `.temp/TASK-260924-1aa9wb-validation/base-src`) — exit 0.
- `go build -buildvcs=false -o .temp/TASK-260924-1aa9wb-validation/output/candidate ./cmd/curator` — exit 0.
- `git diff --check` — exit 0.

### Schema-1 output identity

Built the base binary from the recorded `HEAD` source and the candidate binary from this worktree. Ran both through the real CLI against a schema-1 project with a tagged local Git skill. Each invocation started from an identical restored fixture state. Base used `CURATOR_DRAFT_SOURCES_V1=1`; candidate had the variable unset. `project resolve app`, `list`, `install app`, and `status app --check` were 4/4 byte-identical: all exits were 0, stdout and stderr matched, and stderr was empty. The per-command hashes and byte counts are in `TASK-260924-1aa9wb_v1-cli-identity.json`.

### Narrowing mutants

Both mutants were applied only to disposable copies of the candidate tree; each exact test failed with exit 1 as expected, killing the mutant:

| Mutant | Production-entry test | Result |
|---|---|---|
| Route schema-1 projects into the schema-2 resolution lane | `env -u CURATOR_DRAFT_SOURCES_V1 go test ./cmd/curator -run '^TestProjectResolveSchema1KeepsReadOnlyMeaning$' -count=1` | Exit 1; the CLI no longer produced the schema-1 path report and reached the source resolver. Killed. |
| Refuse schema-2 projects again | `env -u CURATOR_DRAFT_SOURCES_V1 go test ./cmd/curator -run '^TestProjectResolveLocalCreatesLockThroughCLI$' -count=1` | Exit 1; CLI resolve failed. Killed. |

### Broad-suite time limits

- `env -u CURATOR_DRAFT_SOURCES_V1 go test ./internal/envprofile` — exit 1 after the Go test timeout at 10 minutes in `TestReinstallEmitsSurfacingBeforePublication`. The focused source/overlay/path set and that test run alone passed as recorded above. The full package is not reported as passing.
- `env -u CURATOR_DRAFT_SOURCES_V1 go test ./internal/crossconformance` — exit 1 after the Go test timeout at 10 minutes inside `TestDraftSourcesSemanticCases/external-evidence-mismatch-declared_identity`. No assertion failure was reported before the timeout. The bounded schema, corpus, semantic coverage, local CLI and selected conformance subsets listed above passed; the full package is not reported as passing.

The earlier project-snapshot implementation applied to environment-profile scope produced profile-contract differences. It was removed in accordance with the scope rule in the superseding operator brief; profile paths again use the environments context store and retain their existing semantics.

No `LOGBOOK.md` change was made per the campaign instruction; this task-scoped result records the implementation decisions, anomalies and validation evidence.
