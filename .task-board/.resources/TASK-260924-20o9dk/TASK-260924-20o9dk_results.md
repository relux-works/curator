# TASK-260924-20o9dk results — developer handoff

## Candidate and coverage

Candidate base is converged `9f0da708350c15e78b7c900022629aa8e42e5e12`. The corpus is vendored under `internal/crossconformance/testdata/skillfile-sources-v1` from curator-spec commit `574636785c9da22757095ca279e8a9da801156ec`; `SKILLFILE_SOURCES_PIN` and `MANIFEST.sha256` are checked by the passing pin test. The old `draft-sources-v1` vendoring and pin were removed. Published counts are 121 schema cases, 105 semantic cases, and 3 snapshot vectors.

| Family | Passing production-entry rows | Explicit bounds | Gaps |
|---|---:|---:|---:|
| Schema cases | 118/121 | 3/3 | 0 |
| Semantic cases | 105/105 | 0 | 0 |
| Snapshot vectors | 3/3 | 0 | 0 |

The three `local-snapshot-v1` schema documents remain explicit bounds under the attached `TASK-260924-20o9dk-decision-1.md` decision. The exact reason is present in the test classification: “Curator never reads serialized inventory JSON; inventory is recomputed from the stored tree (internal/snapshot Capture/OpenLocal)”. The matrix checks the pinned corpus against driver IDs in both directions. Semantic cases are split into five bounded tests; the count pin is 21, 28, 14, 19, and 23 cases.

Production changes cover current resolved endpoint use for refresh of an existing checkout; pre-I/O `repository_policy_invalid` for an SCP-like endpoint with an alias port; SSH URI alias-port rendering; deny-wins repository+commit revocation under advisory policy; machine-global schema-2 refusal and project schema-2 acceptance; and path/Git lock replay, including locked-object fetch, moved tags, listed mirrors, and unreachable sources. Replay verifies fetched repository object format against the locked package before reading/installing content, still verifies commit, identity, and `content_sha256`, leaves lock bytes unchanged, and does not resolve a declared tag. `driveV2DeclaredMirror` now compares its observed outcome to `c.Expected`. The converged `draftsources.go` still calls `declaredDependencyReplaySources` while replaying transitive dependencies; the C1 object-format check is retained alongside that recovery.

## Validation run on this host

Host: `GOOS=darwin GOARCH=amd64`. Each command below was run directly, without a pipe.

- `go test ./internal/crossconformance -run '^(TestSkillfileSourcesPin|TestSkillfileSourcesCorpusCounts|TestDraftSourcesSchemaCases|TestDraftSourcesSnapshotVectors|TestDraftSourcesSemanticCoverage)$'` — exit 0.
- `go test ./internal/crossconformance -run '^TestDraftSourcesSemanticCasesBatch0$'` — exit 0, 81.250s.
- `go test ./internal/crossconformance -run '^TestDraftSourcesSemanticCasesBatch1$'` — exit 0, 68.022s.
- `go test ./internal/crossconformance -run '^TestDraftSourcesSemanticCasesBatch2$'` — exit 0, 59.928s.
- `go test ./internal/crossconformance -run '^TestDraftSourcesSemanticCasesBatch3$'` — exit 0, 71.958s.
- `go test ./internal/crossconformance -run '^TestDraftSourcesSemanticCasesBatch4$'` — exit 0, 84.808s.
- `go test ./internal/crossconformance -run '^TestDraftSourcesReplayRejectsLockedObjectFormatMismatch$'` — exit 0.
- `go test ./internal/crossconformance -run '^TestDraftSourcesPlaybookCollectionAcceptanceThroughProductionCLI$'` — exit 0, 60.139s.
- `go test ./internal/config -run '^(TestParseSourcePolicyV2Refusals|TestAttemptConnectionURL|TestAttemptConnectionURLRefusesMistranslations|TestAttemptConnectionURLFollowsLoaderPlans)$'` — exit 0.
- `go test ./internal/registry -run '^TestResolveExact$'` — exit 0.
- `go test ./internal/gitops` — exit 0, 14.210s.
- `go test ./internal/install -run '^(TestDraftInstallReplaysMissingLocalSnapshot|TestDraftFreshMachineMovedTagReplaysLockedCommitWithoutRefResolution|TestDraftFreshGitReplayChecksContentHashBeforePublishingSnapshot|TestDraftFreshGitSourceWithoutEndpointIsUnavailableAndPreservesLock)$'` — exit 0.
- `go test ./internal/snapshot -run '^(TestCaptureDetectsLiveContentChange|TestReviewIdentityReplacementDuringCapture)$'` — exit 0.
- `git diff --check HEAD` — exit 0.
- `go build -o "$TMPDIR/curator-20o9dk-build" ./cmd/curator` — exit 0. An earlier identical build directed to `/tmp` also exited 0; that binary was removed and the build was repeated under `$TMPDIR`.
- `make lint` — exit 0, 0 issues.

The five cross-conformance semantic batches and focused replay/CLI rows were run on this Darwin host. Hosted Linux, macOS, and Windows lanes were not run by this producer; the orchestrator still needs to verify hosted evidence.

## Mutant evidence

Each valid mutant run used `-count=1` to prevent Go's test cache from substituting an earlier result. Temporary mutations were restored byte-for-byte before handoff.

- Fetch stored origin instead of the current endpoint plan: `go test -count=1 ./internal/crossconformance -run '^TestDraftSourcesSemanticCasesBatch3$'` — exit 1 at `v2-refresh-current-endpoint-existing-checkout`; the stale origin did not satisfy the current endpoint fixture.
- Render/connect an SCP-like alias port: `go test -count=1 ./internal/crossconformance -run '^TestDraftSourcesSemanticCasesBatch4$'` — exit 1 at `v2-scp-alias-port-refused`; the mutant reached `repository_endpoint_unavailable` instead of refusing as `repository_policy_invalid` before I/O.
- Ignore repository+commit revocation under advisory: `go test -count=1 ./internal/registry -run '^TestResolveExact$'` — exit 1; the broad revocation returned `unknown` rather than denial.
- Skip locked object-format verification: `go test -count=1 ./internal/crossconformance -run '^TestDraftSourcesReplayRejectsLockedObjectFormatMismatch$'` — exit 1; the forged SHA-1-width lock installed from the SHA-256 repository.

Refresh-mutant iteration notes: an initial run reported `(cached)` and exited 0, so it was not counted as mutant evidence. The first uncached mutation exited 1 at compile time because its loop variable became unused; that was also not counted. The compile-preserving mutation then failed at the intended refresh row with exit 1.

## Scope notes

No `CHANGELOG.md` or `LOGBOOK.md` edit was made. The task-scoped result resource records the relevant findings and the accepted local-snapshot bounds. The three hosted OS lanes remain an orchestrator/CI verification item.

## Addendum — convergence and bounded rerun (2026-09-26)

The post-convergence worktree is based on `9f0da708`. `internal/install/draftsources.go` retains the transitive replay recovery through `declaredDependencyReplaySources` and the C1 repository-object-format comparison. The vendored corpus was compared file-by-file with curator-spec commit `574636785c9da22757095ca279e8a9da801156ec`: 135/135 source files matched, exit 0. The tracked and untracked worktree paths were within this task's corpus, driver, implementation, test, and CI-count paths; there are no `CHANGELOG.md` or `LOGBOOK.md` edits.

### Fresh validation on this host

Host: `GOOS=darwin GOARCH=amd64`.

- `go test -count=1 ./internal/crossconformance -run '^(TestSkillfileSourcesPin|TestSkillfileSourcesCorpusCounts|TestDraftSourcesSchemaCases|TestDraftSourcesSnapshotVectors|TestDraftSourcesSemanticCoverage)$'` — exit 0.
- `go test -count=1 ./internal/crossconformance -run '^TestDraftSourcesSemanticCasesBatch0$'` — exit 0 (21 rows).
- `go test -count=1 ./internal/crossconformance -run '^TestDraftSourcesSemanticCasesBatch1$'` — exit 0 (28 rows).
- `go test -count=1 ./internal/crossconformance -run '^TestDraftSourcesSemanticCasesBatch2$'` — exit 0 (14 rows).
- `go test -count=1 ./internal/crossconformance -run '^TestDraftSourcesSemanticCasesBatch3$'` — exit 0 (19 rows).
- `go test -count=1 ./internal/crossconformance -run '^TestDraftSourcesSemanticCasesBatch4$'` — exit 0 (23 rows).

Together the five batches exercise 105/105 released semantic rows.
- After restoring all temporary mutants, batches 3 and 4 were rerun — exit 0 for both.
- After restoring all temporary mutants, `go test -count=1 ./internal/crossconformance -run '^(TestDraftSourcesReplayRejectsLockedObjectFormatMismatch|TestDraftSourcesPlaybookCollectionAcceptanceThroughProductionCLI)$'` — exit 0.
- `go build -o "$TMPDIR/curator-20o9dk-final-build" ./cmd/curator` — exit 0; output stayed under `$TMPDIR`.
- `golangci-lint run ./cmd/curator ./internal/config ./internal/gitops ./internal/install ./internal/registry ./internal/snapshot ./internal/crossconformance` — exit 0, 0 issues.
- `go vet ./cmd/curator ./internal/config ./internal/gitops ./internal/install ./internal/registry ./internal/snapshot ./internal/crossconformance` — exit 0.
- `gofmt -l` on changed Go files — exit 0, no files listed; `git diff --check HEAD` — exit 0.

### Fresh mutant evidence

Each behavior mutant was temporary and restored. The listed nonzero exit is the expected red result that killed the mutant.

| Mutant | Production-entry gate | Result |
|---|---|---|
| Refresh from stored origin instead of current resolved endpoint | Semantic batch 3, `v2-refresh-current-endpoint-existing-checkout` | exit 1; current endpoint row failed because the old origin was used |
| Allow and render SCP-like alias port | Semantic batch 4, `v2-scp-alias-port-refused` | exit 1; row observed `repository_endpoint_unavailable`, not the required pre-I/O `repository_policy_invalid` |
| Ignore revocation when registry policy is advisory | Semantic batch 3, `attestation-evidence-revoked-identity-commit-advisory` | exit 1; fresh production install returned `ok` instead of refusing |
| Skip the locked object-format comparison | `TestDraftSourcesReplayRejectsLockedObjectFormatMismatch` | exit 1; the forged SHA-1-width lock installed from the SHA-256 repository |

The first refresh-mutant attempt in this turn also exited 1, but only because the temporary mutation left an unused loop variable; it was excluded. A compile-preserving behavior mutant then failed at the intended row with exit 1. All final positive tests above ran after restoring the production sources.

The local Darwin gates passed. Hosted Linux, macOS, and Windows lanes were not run in this producer turn and remain unverified here; the orchestrator's hosted checks are still the arbiter.

## CHANGELOG entry (for release prep)

Update Curator's Skillfile sources v1 conformance to the released corpus, verify locked Git object formats during replay, use current resolved endpoints on refresh, reject SCP-like alias ports before I/O, and enforce repository+commit revocation under advisory policy.

## Revision 4 — gate fix (2026-09-26)

The hosted gate’s three repeated failures reproduced locally before this revision: the two refresh fixtures exited with `repository_endpoint_unavailable` because they had no machine source policy for the now-current endpoint plan, and the process-boundary test rejected `draftsources_gitshim_test.go` for starting processes directly (command exit 1). The production refresh rule stayed unchanged.

Both refresh fixtures now install a current machine `source-policy.json` endpoint plan, route that endpoint to their local bare repository, and assert the checkout’s stored `remote.origin.url` did not select the endpoint. The agent/askpass fixture still observes both variables at Git and successfully refreshes; the hostile user-config fixture still proves the config payload is live and verifies refresh fetches through the current endpoint. The Git shim is now compiled through `internal/testcli.Run`, with its portable helper source embedded from a non-Go fixture file, so the shared process-boundary guard passes without an allow-list addition.

### Revision 4 validation

Host: `GOOS=darwin GOARCH=amd64`. Each command ran directly as a standalone process.

- Before the fix, `go test ./internal/crossconformance -run 'TestDraftLiteralKeepsAgentAndAskpass|TestDraftLiteralRefreshIgnoresUserConfig|TestIntegrationSurfaceStartsNoProcessOutsideTheSharedSeams' -count=1` — exit 1, reproduced all three failures.
- After the fix, the same command — exit 0 (7.037s).
- `go test -race ./internal/crossconformance -run 'TestDraftLiteralKeepsAgentAndAskpass|TestDraftLiteralRefreshIgnoresUserConfig' -count=1` — exit 0 (9.102s).
- `go test ./internal/crossconformance -run '^TestDraftSourcesSemanticCasesBatch2$' -count=1` — exit 0 (31.777s), including `v2-ssh-uri-alias-port`.
- `go test ./internal/crossconformance -run '^TestDraftSourcesSemanticCasesBatch3$' -count=1` — exit 0 (47.976s), including `attestation-evidence-revoked-identity-commit-advisory` and `v2-refresh-current-endpoint-existing-checkout`.
- `go test ./internal/crossconformance -run '^TestDraftSourcesSemanticCasesBatch4$' -count=1` — exit 0 (54.640s), including `v2-scp-alias-port-refused`.
- `git diff --check` — exit 0.

The targeted tests compiled the modified crossconformance package and the fixture shim, and the CLI was built through the shared test seam. I did not run the full module suite locally. The hosted Linux, macOS, and Windows lanes remain for the configured handoff validation; only this Darwin host is verified in this revision.


## Revision 5 (carry-forward republish, 2026-09-26)

Trunk moved to `3bdcfe07`; the orchestrator converged this task's accepted revision-4 delta uncommitted into the Story worktree. This revision changes no content: it verifies the carry-forward and republishes.

### Per-path verification against `TASK-260924-20o9dk_change-request_rev4.patch` (166 paths)

- Reconstructed rev4 post-images by applying the patch to a pristine `git archive HEAD` copy (`/tmp/base-check`, scratch only, outside the repo): 165/165 non-intersecting paths applied cleanly, proving trunk did not touch them.
- Byte-compared (`cmp`) every reconstructed post-image with the worktree file, including the 12 new files now untracked (`draftsources_gitshim_main.go.txt`, `draftsources_gitshim_test.go`, `draftsources_semantic_replay_test.go`, `draftsources_semantic_scope_test.go`, and 8 `testdata/skillfile-sources-v1` pin/corpus/schema files) and every `testdata/draft-sources-v1` deletion: 165/165 byte-identical, 0 mismatches.
- The single intersecting path, `internal/snapshot/capture.go`, fails `git apply --check` only on a one-line context shift (trunk added the `stateread` import above the hunk). The worktree file contains both sides: trunk's `stateread` import and call sites plus rev4's `sync` import, mutex-guarded `captureAfterCopyHook func(*LocalAcquisition)`, and `SetCaptureAfterCopyHookForTesting`. No conflict markers anywhere (`grep` for `<<<<<<<`/`>>>>>>>` over `internal/`, `cmd/` empty).
- `git diff --name-only HEAD -- . ':!.task-board'` lists only rev4-patch paths: 154 tracked + 12 untracked = 166/166. No other file (no trunk revert). The rev4 patch itself contains no `CHANGELOG.md`/`LOGBOOK.md`, root `TASK-*`/`BUG-*`, `test/`, or `ledger/` path, so the "minus CHANGELOG.md" clause is vacuous.

### CHANGELOG policy

No `CHANGELOG.md` or `LOGBOOK.md` edit in the worktree (`git diff HEAD --` clean for both; no stray root `TASK-*.md`/`BUG-*.md`). The release-prep entry text remains verbatim in the "## CHANGELOG entry (for release prep)" section above.

### Fresh bounded validation on the converged tree (this host, `GOOS=darwin GOARCH=amd64`, every command standalone, real exit codes)

- `go test -count=1 ./internal/crossconformance` split runs — exit 0 each: non-semantic Draft/Literal/Integration selection (schema, snapshot, coverage, literal, SSH, transport, broker, CLI, digest, surrogate, integration; 67.6s); `TestDraftSourcesReplayRejectsLockedObjectFormatMismatch` + `TestDraftSourcesCLIInstallRestoresPriorState` (7.9s); semantic batches 0–4 (45.9s, 47.1s, 36.7s, 53.1s, 63.2s); `TestDraftSourcesPlaybookCollectionAcceptanceThroughProductionCLI` (50.5s).
- `go test -count=1 ./internal/install -run 'Draft|Replay'` in three bounded chunks — exit 0 each: audit/build/evidence (Failure|Audit 204.4s, Build|Evidence 51.8s); git/marker/repair/refresh/local (196.0s); runtime/install/replay/fresh/acquire/transport/plan/ssh/manager (50.1s).
- `go build ./cmd/curator ./internal/snapshot ./internal/install ./internal/crossconformance` — exit 0. `go vet` over the seven changed Go packages — exit 0. `gofmt -l` over changed trees — clean. `git diff --check HEAD` — clean. `make lint` (`golangci-lint run`) — exit 0, 0 issues.
- Not run here: hosted Linux/macOS/Windows lanes (orchestrator/CI item, as in prior revisions).
