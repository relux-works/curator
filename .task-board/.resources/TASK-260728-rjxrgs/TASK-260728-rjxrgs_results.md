# TASK-260728-rjxrgs results

## Implementation

- Added published-case consumers for marker v3 schemas and external-repository lifecycle tables. The marker cases reach `marker.Read`; the mixed-build, shim/PATH, signing, and transaction cases reach the project install entry.
- Preserved receipt v1 for local `go-v1` commands and receipt v2 for external commands. The receipt-v2 test reads the published expected input bytes and cache key; receipt-v3 retains its assurance binding.
- The schema-7 mixed case reads and compares `expected_marker` from the published mixed-marker fixture and compares staging/publication against the mixed-plan fixture.
- Project and global staging order external commands before local commands. Command collisions fail before acquisition; output/PATH selection and package signing requests fail closed; transaction tests exercise publication rollback and uncertain-journal recovery.
- The exact marker-v3 schema-case count is 27. The count and lifecycle-consumer rows are present in the CI ratchet/manifests.
- No Windows-reserved command names were added. No repository CHANGELOG or LOGBOOK was edited.

## Measured conformance coverage

| Table | Driven | Bound | Total |
|---|---:|---:|---:|
| `marker/install-marker-v3/schema-cases` | 22 | 5 | 27 |
| `external-repository-lifecycle/mixed_build_cases` | 6 | 0 | 6 |
| `external-repository-lifecycle/path_shim_cases` | 3 | 0 | 3 |
| `external-repository-lifecycle/signing_cases` | 3 | 1 | 4 |
| `external-repository-lifecycle/transaction_cases` | 4 | 0 | 4 |
| `external-repository-lifecycle/status-repair-gc-cases` | 1 | 4 | 5 |

These ratios were recorded in the previously attached results artifact. This run reran the corresponding consumers against the pinned corpus; the test selections passed. The three mutant results below are also accepted from that attached artifact, rather than rerun in this session.

### Explicit bounds

- Marker-v3 cases bound by the shared v3/v4 validators and owned by `BUG-260923-2afgyq` by reference: `invalid-external-declared-effective-mismatch.json`, `invalid-marker-local-identity-kind-mismatch.json`, `invalid-marker-network-identity-kind-mismatch.json`, `invalid-marker-sha1-effective-revision-width.json`, and `invalid-marker-sha256-effective-revision-width.json`. No marker-v4 behavior was changed here.
- `release-pipeline-signing` is bound because the install entry has no post-signing operation; release signing is outside this manager surface.
- `status-current` is bound because external status planning acquires exact source before cache inspection and emits no external build-currentness fact, so status cannot report this external command current without remote contact.
- `status-missing-snapshot` is bound because the manager status classifier consumes local build-plan facts and has no external non-current result for a missing protected snapshot.
- `status-unreadable-protected-state` is bound because the manager status classifier has no external protected-state fact for unreadable entries; current external planning acquires and audits source before reading the cache, contrary to the vector's `remote_contacted=false` premise.
- `repair-reacquires-exact-source` is bound because manager install/update/status entry points expose no repair operation wired through reacquisition, audit, publication, and marker transaction. No repair shim was added to force this surface.
- `gc-retains-roots` is driven for artifact receipts, in-flight journals, install markers, protected snapshots, and uncertain entries; eligible orphans are collected.

## Mutant evidence accepted from the previously attached artifact

- Receipt interpretation aliasing: the weakened production version check was killed by `TestReadAuthoritativeMarkerV3SchemaCases`, exit 1.
- Mixed-plan order: local-before-external staging was killed by `TestAuthoritativeMixedBuildCasesUseProjectInstallEntry`, exit 1.
- Skipped transaction step: omitting the external-cache commit step was killed by `TestAuthoritativeTransactionCasesUseProjectInstallEntry`, exit 1.

The prior artifact reports all mutations reverted. I did not rerun these mutants in this session.

## Verification in this session

The conformance root was extracted from the workflow-pinned release commit `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`; its `manifest.json` SHA-256 is `be11bb1e4c46f21fb5684d586f9c2a8b0d59f3b437bc7ea7aa5aa530fe4d47ca`.

- `go test ./internal/marker -run '^TestReadAuthoritativeMarkerV3SchemaCases$' -count=1` with `CURATOR_CONFORMANCE_ROOT` set to the pinned root — exit 0.
- `go test ./internal/buildrepo -run '^(TestExternalReceiptV2CacheKeyVector|TestExternalReceipt3.*|TestExternalProtectedCache.*|TestSubstitutionCannotAliasDeclaredCacheKey)$' -count=1` with the pinned root — exit 0.
- `go test ./internal/install -run '^Test(LegacyMixedBuildProjectInstallProducesMarkerV3|ExternalCommandNameCollisionFailsBeforeMutation|GlobalMixedBuildStagesExternalBeforeLocal|AuthoritativeMixedBuildCasesUseProjectInstallEntry|AuthoritativePathShimCasesUseProjectInstallEntry|AuthoritativeSigningCasesUseProjectInstallEntry|AuthoritativeTransactionCasesUseProjectInstallEntry|LegacyBuildsKeepReceipt1|DraftBuildsPublishReceipt3OnBothArms)$' -count=1` with the pinned root — exit 0 (198.113s).
- `go test ./internal/scopes -run '^TestAuthoritativeGarbageCollectionRootsAreRetained$' -count=1` with the pinned root — exit 0.
- `go test ./internal/envprofile -run '^TestManagerOwnedAbsenceReadsAreGuarded$' -count=1` — exit 0.
- `go test ./internal/skillspec -count=1` with the pinned root — exit 0.
- `go test ./internal/marker -count=1` with the pinned root — exit 0.
- `golangci-lint run` — exit 0 (0 issues).
- `go build -o /tmp/curator-task-260728-rjxrgs ./cmd/curator` — exit 0.
- `git diff --check` — exit 0.

The initial `go test ./internal/marker -run '^TestReadAuthoritativeMarkerV3SchemaCases$' -count=1` used the spec checkout's newer `main` root (`add50233`) instead of the workflow pin and exited 1: its manifest/case hash did not match the five rows in the pinned gap ledger. This was an input-root mismatch, not counted as a passing gate; the exact pinned-root rerun above exited 0.

The previously attached artifact reports `go test ./internal/install -count=1` exiting 1 after the single-command timeout at 10 minutes. I did not rerun that unbounded package command; this session split and reran the task-specific install selection in 198.113 seconds. No Windows or Linux hosted lane was run locally; those remain for the post-handoff hosted gate.

## CHANGELOG entry (for release prep)

Consume the pinned mixed external-build lifecycle corpus through production install and marker paths, preserving receipt namespaces and rollback guarantees.

## Board note

The required initial `task-board m 'set_status(TASK-260728-rjxrgs, status=development)'` command exited 0 and reported the task already in `development`. The earlier attached artifact's dependency-rejection note is stale; the current board query reports `development` and the dependency was removed by the orchestrator.
