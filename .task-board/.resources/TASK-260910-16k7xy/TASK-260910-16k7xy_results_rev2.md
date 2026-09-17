# TASK-260910-16k7xy results rev2 — capture and store local package snapshots

Worktree branch task-board/story/STORY-260910-24nyb1, checkpoint 3dbefcd (14hsti boundaries leaf). Work left UNCOMMITTED for handoff snapshot.

## What was delivered
- internal/snapshot/capture.go (new): BuildInventory (local-snapshot-v1 byte inventory: portable path, raw-byte sha256, executable bit; UTF-8 sorted; duplicates and case-fold collisions fail), Inventory.Digest (SHA-256 of CCJ-1 of {schema_version:1, algorithm:curator-local-snapshot-v1, files}), Capture (copy admitted dirty/untracked bytes to private staging, then re-enumerate membership + re-read live bytes + rehash frozen copy; any drift fails source_snapshot_changed), PublishLocal (rehash frozen copy before publication, content-addressed immutable store below home/local-snapshots, concurrent winner reused only after authentication), OpenLocal (locked digest resolves to authenticated frozen tree; absent fails source_snapshot_unavailable, tampered fails source_snapshot_changed; never recreates from live bytes, never synthesizes a Git commit).
- internal/snapshot/boundaries.go: LocalAcquisition carries Outputs + RootEntries so Capture re-enumerates with the exact admission inputs.
- internal/envprofile/envprofile.go: draft path pipeline admitPathSource now runs PrepareLocalAcquisition -> Capture -> PublishLocal -> OpenLocal and stateForPath installs the returned frozen snapshot, never the live directory; snapshot store added to managed draftPathOutputs. Install operand path and overlay path declarations both route through stateForPath. Frozen v1 path (DraftSourcesV1=false) untouched.
- Tests: internal/snapshot/capture_test.go (3 normative vectors pinned from conformance/draft-sources-v1/snapshot-cases.json, dirty/untracked-inside-Git incl. no-commit-synthesized check, added/removed/live-change races, frozen-mutation, publish/open round-trip, missing + tampered opens), internal/envprofile/draft_capture_test.go (Install consumes frozen snapshot across live mutation, dirty+untracked install, store-inside-store refusal), 1-line admitPathSource signature fix in boundaries_test.go.

## Verification (real exit codes, narrow runs only)
- go test -count=1 ./internal/snapshot/ -> ok, EXIT 0
- go test -count=1 -run TestDraftInstall|TestAdmitPathSource|TestDraftPath ./internal/envprofile/ -> ok, EXIT 0 (279s, host stall on fresh test binary as warned)
- go vet ./internal/snapshot/ -> VET_EXIT 0; go vet ./internal/envprofile/ -> VET_EXIT 0; gofmt -l on both packages -> clean
- Independent oracle: python hashlib+json recomputed all 3 conformance vectors (per-file sha256 + inventory digest) -> all OK, matching the pinned test values
- Mutant 1 (PublishLocal pre-publication rehash removed) -> TestPublishDetectsFrozenMutation FAILS (exit 1), killed; bytes restored, package green again
- Mutant 2 (Capture membership re-enumeration removed) -> TestCaptureDetectsAddedMember FAILS (exit 1), killed (RemovedMember still caught by live re-read gate); bytes restored, package green again

## Acceptance mapping
- Dirty/untracked bytes inside Git captured, .git excluded, no commit synthesized: TestCaptureAdmitsDirtyAndUntrackedInsideGit + TestDraftInstallAdmitsDirtyUntrackedInsideGit
- Exact byte inventory + executable flags, all 3 normative vectors: TestCaptureMatchesNormativeVectors + independent python check
- Runtime/build/dependency inputs frozen: TestVectorsFreezeContextAcrossRuntimeAndBuildEdits (identical SKILL.md, 3 distinct digests)
- Capture races fail: AddedMember/RemovedMember/LiveContentChange/FrozenMutation rows, all source_snapshot_changed
- Missing snapshots fail: TestOpenMissingSnapshot (absent + removed-after-publish), source_snapshot_unavailable
- Live directory never used after capture: stateForPath consumes OpenLocal result; TestDraftInstallFreezesLiveMutations
- Draft/opt-in only behind DraftSourcesV1; frozen v1 untouched

## Notes for reviewer
- Case-fold collision rule in BuildInventory is fail-closed on all platforms (Linux case-only-distinct pairs are refused to keep the digest portable); no survivor.
- Lock-identity {kind:local-snapshot} in Skillfile.lock.json is NOT this leaf (transport/lock leaf owns it); install-time consumption of the frozen tree is.
