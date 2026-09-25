# TASK-260924-m28s6b results

## Implementation

Schema-2 install and upgrade now replay missing project snapshots from the lock’s declared source. Git replay requests the full locked object ID with tags disabled; when machine bindings are absent it resolves the declared endpoint policy and fetches that same object ID. Path replay reads current source bytes. Both paths verify the locked package identity and `content_sha256` before publishing a snapshot. A mismatch reports `source_snapshot_changed`; a missing or unreachable source reports `source_snapshot_unavailable`. Replay never writes `Skillfile.lock.json`.

Updated README, CLI, and troubleshooting guidance to describe the committed lock and fresh-machine behavior. `Skillfile.lock.json` remains project-owned; machine bindings remain outside the project. No `CHANGELOG.md` edit was made per campaign policy.

## Acceptance evidence

| Acceptance row | Production-entry test evidence | Result |
|---|---|---|
| Git tag, repository, and path through install and upgrade with an empty snapshot store | `TestDraftFreshMachineReplaysEverySourceThroughInstallAndUpgrade` (uses `Project` with `Fetch` false/true) | All rows pass; lock bytes remain identical |
| Git tag and logical repository replay with no machine bindings | `TestDraftFreshMachineReplaysDeclaredGitSourceWithoutBindings` (local bare fixture rewrites the declared endpoint; no network) | All install/upgrade rows pass; lock bytes remain identical |
| Moved tag does not change a locked result | `TestDraftFreshMachineMovedTagReplaysLockedCommitWithoutRefResolution` moves and pushes the tag after resolve | Upgrade uses the originally locked bytes; lock stays byte-identical |
| Exact Git object fetch does not fetch tags | `TestFetchCommitFromURLIsolatedRequestsLockedObjectWithoutTags` | Locked tree is present; tag refs remain empty |
| Changed package identity and changed content hash | `TestDraftFreshPathMismatchAndUnreachableSourcePreserveLock`, `TestDraftFreshGitReplayChecksContentHashBeforePublishingSnapshot` | `source_snapshot_changed`; mismatching snapshot is not published; lock stays unchanged |
| Unreachable sources | Path removal in `TestDraftFreshPathMismatchAndUnreachableSourcePreserveLock`; missing logical-repository endpoint in `TestDraftFreshGitSourceWithoutEndpointIsUnavailableAndPreservesLock` | `source_snapshot_unavailable`; lock stays unchanged; no network used |
| Lock is not ignored or machine-private | `TestDraftInstallDoesNotIgnoreOrPrivatizeSkillfileLock` drives `Project` with `FixGitignore`, checks the lock is in the project, bindings are outside it, and `git check-ignore` exits 1 | Pass |

## Narrowing mutants

Each mutant ran in a disposable copy. Exit 1 is the expected red result and means the mutant was killed.

| Mutant | Test command | Exit | Why it was killed |
|---|---|---:|---|
| Accept on package identity alone (disable content-hash guards) | `go test ./internal/install -run '^TestDraftFreshGitReplayChecksContentHashBeforePublishingSnapshot$' -count=1` | 1 | The forged lock installed successfully; the test required `source_snapshot_changed` before publication. |
| Re-resolve the declared ref and consume the moved tag | `go test ./internal/install -run '^TestDraftFreshMachineMovedTagReplaysLockedCommitWithoutRefResolution$' -count=1` | 1 | The moved-tag upgrade no longer replayed the locked snapshot. |
| Accept a path snapshot without the replay hash check | `go test ./internal/install -run '^TestDraftFreshPathMismatchAndUnreachableSourcePreserveLock$/^content-hash-mismatch-with-matching-package-identity$' -count=1` | 1 | The test found the mismatching snapshot had been published. |

## Final validation

All commands below were run directly. Exit codes are the actual command results.

- `go test ./internal/install -run '^TestDraftFreshMachineReplaysEverySourceThroughInstallAndUpgrade$' -count=1 -timeout=2m -v` — exit 0.
- `go test ./internal/install -run '^TestDraftFreshMachineReplaysDeclaredGitSourceWithoutBindings$' -count=1` — exit 0.
- `go test ./internal/install -run '^TestDraftFreshGitSourceWithoutEndpointIsUnavailableAndPreservesLock$' -count=1` — exit 0.
- `go test ./internal/install -run '^(TestDraftFreshMachineMovedTagReplaysLockedCommitWithoutRefResolution|TestDraftFreshGitReplayChecksContentHashBeforePublishingSnapshot)$' -count=1` — exit 0.
- `go test ./internal/install -run '^(TestDraftFreshPathMismatchAndUnreachableSourcePreserveLock|TestDraftInstallReplaysMissingLocalSnapshot|TestDraftLocalMissingSnapshotReplays|TestDraftInstallDoesNotIgnoreOrPrivatizeSkillfileLock)$' -count=1` — exit 0.
- `go test ./internal/gitops -run '^TestFetchCommitFromURLIsolatedRequestsLockedObjectWithoutTags$' -count=1` — exit 0.
- `go build -o "$TMPDIR/curator-m28s6b-build/curator" ./cmd/curator` — exit 0.
- `golangci-lint run ./internal/install ./internal/gitops` — exit 0, 0 issues.
- `git diff --check` — exit 0.

A broad run, `go test ./internal/install -run '^(TestDraft|TestSchema2)' -count=1`, was interrupted after about eight minutes of silence; its real exit was 1 and it is not reported as passing. A combined focused acceptance run was also interrupted (exit 1) while host test runs were active. Its acceptance cases were rerun in the bounded green commands above. During iteration, earlier focused commands exited 1 on compilation/fixture mistakes and on a Git mismatch initially being downgraded to unavailable; those issues were corrected and the affected cases passed in the final commands above.

## CHANGELOG entry (for release prep)

- Skillfile schema 2 source handling is supported by default for projects.
  Schema 1 retains its exact meaning, and no on-disk migration is implicit.
- Schema-2 installs and upgrades replay missing locked snapshots from declared Git or path sources, validate package identity and content hashes before publication, and preserve the committed `Skillfile.lock.json`.

## Findings for the task record

The Git hash regression caught a retry-classification defect: after an exact commit fetch succeeded but content validation failed, endpoint fallback changed `source_snapshot_changed` to `source_snapshot_unavailable`. Replay now returns the mismatch directly after a successful fetch. The no-bindings integration row also caught that the fetched repository had to be carried into frozen consumption; that repository handoff is now retained for the duration of the install.

## Revision 2

Revision 1 was rejected on the race lane because `TestDraftDocsPinExamples` still pinned the obsolete `source_snapshot_unavailable` remedy. Updated the CLI remediation table and troubleshooting guidance to match lock replay semantics: `source_snapshot_changed` directs operators to restore the locked package identity and `content_sha256`, with refresh reserved for intentional lock changes; `source_snapshot_unavailable` directs operators to restore access to the declared path or Git source. The docs test pins the new remedy phrases. Its unavailable CLI row now creates a lock, removes both the declared path and local snapshot, and exercises install through the CLI.

No lock replay implementation or `CHANGELOG.md` edits were made in this revision. The release-prep changelog entry and the Revision 1 acceptance/mutant evidence above remain in this resource. Mutants were not rerun in this bounded rework.

### Revision 2 validation

All commands were run directly; exit codes are the actual results.

- `go test ./cmd/curator -run 'TestDraftDocsPinExamples|Diagnostic'` — first run exit 1: the initial changed-remedy pin crossed a doc line break. After narrowing the pin to contiguous remedy text, rerun exit 0.
- `go test ./cmd/curator -run 'TestWithDraftRemediationTable|TestDraftRemediationThroughCLI'` — first run exit 1: the unavailable fixture retained its local snapshot, so install did not attempt source replay. After removing the snapshot store as well as the source path, rerun exit 0.
- `go test ./internal/install -run 'TestDraftInstallReplaysMissingLocalSnapshot|TestDraftFreshMachineReplaysEverySourceThroughInstallAndUpgrade|TestDraftFreshMachineReplaysDeclaredGitSourceWithoutBindings|TestDraftFreshGitSourceWithoutEndpointIsUnavailableAndPreservesLock|TestDraftFreshMachineMovedTagReplaysLockedCommitWithoutRefResolution|TestDraftFreshGitReplayChecksContentHashBeforePublishingSnapshot|TestDraftFreshPathMismatchAndUnreachableSourcePreserveLock|TestDraftInstallDoesNotIgnoreOrPrivatizeSkillfileLock'` — exit 0.
- `go test ./cmd/curator -run 'TestWithDraftRemediationTable|TestDraftRemediationThroughCLI'` — exit 0.
- `go test ./cmd/curator -run 'TestDraftDocsPinExamples|Diagnostic'` — exit 0.
- `go build -o /tmp/curator-m28s6b-revision2 ./cmd/curator` — exit 0.
- `golangci-lint run ./cmd/curator` — exit 0, 0 issues.
- `git diff --check` — exit 0.

### Revision 2 CHANGELOG entry (for release prep)

- Lock replay diagnostic guidance now distinguishes a source mismatch from an unreachable declared source; install retries from the locked source and refresh remains explicit for intentional lock changes.


## Revision 2 — refresh

The refreshed Story checkpoint is `c4ced3a9d75af3bc8b4964387c08e35d38e59387`, parented on trunk `5b326aa384d9e3f3a7c047e7b228107e84c90df9`. The managed workspace reports base `main`. The candidate `CHANGELOG.md` bytes equal trunk; `git diff --exit-code 5b326aa3 -- CHANGELOG.md` exited 0. The default-on and lock-replay release entries are recorded above for release prep. No root-level TASK/BUG markdown files or nested `.task-board` changes remain, and nothing is staged.

This refresh exposed that `TestDraftDocsPinExamples` still required the default-on release entry in `CHANGELOG.md`. The leaf policy requires trunk's changelog bytes, so the test no longer pins release-prep changelog content; schema-2 default-on guidance remains pinned in README and CLI docs. The first post-refresh run failed for the stale changelog assertion, then passed after this adjustment.

### Refresh acceptance reruns

All commands were run directly as standalone processes. Exit codes are the actual command results.

- `go test ./internal/install -run '^(TestDraftFreshMachineReplaysEverySourceThroughInstallAndUpgrade|TestDraftFreshMachineReplaysDeclaredGitSourceWithoutBindings|TestDraftFreshGitSourceWithoutEndpointIsUnavailableAndPreservesLock|TestDraftFreshMachineMovedTagReplaysLockedCommitWithoutRefResolution|TestDraftFreshGitReplayChecksContentHashBeforePublishingSnapshot|TestDraftFreshPathMismatchAndUnreachableSourcePreserveLock|TestDraftInstallReplaysMissingLocalSnapshot|TestDraftLocalMissingSnapshotReplays|TestDraftInstallDoesNotIgnoreOrPrivatizeSkillfileLock)$' -count=1 -timeout=3m` — exit 0. The six install/upgrade source-kind rows, moved-tag row, content-hash mismatch rows, unavailable-source rows, and lock byte/ignore checks passed.
- `go test ./internal/gitops -run '^TestFetchCommitFromURLIsolatedRequestsLockedObjectWithoutTags$' -count=1` — exit 0.
- `go test ./cmd/curator -run 'TestDraftDocsPinExamples|Diagnostic' -count=1` — initial run exit 1 on the obsolete CHANGELOG pin; after updating the test, exit 0.
- `go test ./cmd/curator -run 'TestWithDraftRemediationTable|TestDraftRemediationThroughCLI' -count=1` — exit 0.

### Default-on regression rows

- `env -u CURATOR_DRAFT_SOURCES_V1 go test ./cmd/curator -run '^(TestProjectResolveLocalCreatesLockThroughCLI|TestProjectResolveSchema1KeepsReadOnlyMeaning|TestProjectResolveIgnoresRemovedSwitch|TestProjectResolveGitTagDiffersFromHEADThroughCLI|TestProjectResolveCanonicalRepositorySourceThroughCLI|TestDraftMachinePolicySetupThroughCLI|TestProjectRefreshGitBranchMembershipChangeThroughCLI|TestDraftProjectResolveHelp|TestDraftInstallStatusHelpIncludesSchema2Workflow)$' -count=1` — exit 0.
- `env -u CURATOR_DRAFT_SOURCES_V1 go test ./cmd/curator -run '^TestGlobalStatusRejectsSchema2SkillfileThroughCLI$' -count=1` — exit 0.
- `env -u CURATOR_DRAFT_SOURCES_V1 go test ./cmd/curator -run '^(TestDraftDocsPinExamples|TestDraftDocumentedLocalShapesThroughCLI|TestDraftDocumentedGitRevisionThroughCLI)$' -count=1` — exit 0.
- `env -u CURATOR_DRAFT_SOURCES_V1 go test ./internal/manifest` — exit 0 (Go test cache hit).
- `env -u CURATOR_DRAFT_SOURCES_V1 go test ./internal/install -run '^(TestDraftInstallMissingLockFails|TestDraftInstallStaleLockFails|TestDraftInstallMissingSnapshotFails|TestDraftInstallUsesPinnedBytes|TestDraftInstallGitPinnedSubtree|TestSchema2InstallRequiresLockAndLegacyStillWorks|TestGlobalInstallRejectsSchema2Skillfile)$' -count=1` — exit 0.
- `env -u CURATOR_DRAFT_SOURCES_V1 go test ./internal/crossconformance -run 'TestDraftSources(Pin|CorpusCounts|SchemaCases|SemanticCoverage|SnapshotVectors|CLILocalSkillScriptDependencies|CLIProjectPathWithSpaceAndUnicode|CLISymlinkMemberRefused|BrokerAskpassDispatch|CrossCompile|CLIInstallRestoresPriorState)$' -count=1 -timeout=3m` — exit 0.
- `env -u CURATOR_DRAFT_SOURCES_V1 go test ./internal/crossconformance -run '^TestAcceptedCorpusSatisfiesEveryPublishedStructuralClaim$' -count=1 -timeout=3m` — exit 0.
- `env -u CURATOR_DRAFT_SOURCES_V1 go test ./internal/envprofile -run '^(TestPathInstallAdmitsOrdinaryDirectory|TestLegacyPathInstallIgnoresStoreOverlap|TestPathOverlayAdmitsOrdinaryDirectory|TestPathInstallCapturesDirtyUntrackedInsideGit|TestNestedGitIsSourceInvalid|TestSymlinkInPathIsSourceInvalid|TestUnreadablePathIsUnreadable)$' -count=1 -timeout=3m` — exit 0.
- `env -u CURATOR_DRAFT_SOURCES_V1 go test ./internal/envprofile -run '^TestReinstallEmitsSurfacingBeforePublication$' -count=1 -timeout=3m` — exit 0.

### Build and lint

- `go build -o /tmp/curator-m28s6b-refresh ./cmd/curator` — exit 0.
- `golangci-lint run ./cmd/curator ./internal/install ./internal/gitops ./internal/manifest ./internal/envprofile` — exit 0, 0 issues.
- `golangci-lint run` — exit 0, 0 issues; the runner emitted a warning that a source file in another Story worktree had been removed before generated-file filtering. The focused lint run above completed without warnings.
- `git diff --check` — exit 0.

The narrowing-mutant table and its killed-mutant evidence from the earlier result are retained above; mutants were not rerun during this bounded refresh.

## Revision 3 — Windows no-bindings replay

Hosted run 35990838316 failed only the four no-bindings subtests of `TestDraftFreshMachineReplaysDeclaredGitSourceWithoutBindings` on Windows. The attached hosted artifact was downloaded with `gh run download 35990838316 -R relux-works/curator --dir "$TMPDIR/m28"` (exit 0) and inspected from that temporary directory.

Root cause was fixture-only. `draftReplayGitPath` wrote the temporary Git URL rewrite as `file://` followed by the raw local path. On Windows that produced a quoted Git config URL containing drive-letter backslashes, which does not represent the intended local file URL. The source declaration in the test remains HTTPS. Production resolves its declared endpoint and passes that URL together with the locked full object ID to `FetchCommitFromURLIsolated`; it does not convert a Windows filesystem path into a URL. The existing-binding Windows row uses a local repository remote directly, bypassing the test-only URL rewrite, and passed. No production defect was found.

Changed the test fixture to serialize local paths as portable file URLs (`file:///C:/...` for drive paths), including URL escaping and UNC paths. Added `TestDraftReplayFileURLUsesPortableGitConfigSyntax` to pin POSIX, Windows drive, and UNC forms. The real install/upgrade rows remain unchanged in intent and continue to drive `Project` without machine bindings.

All commands were run directly; exit codes are the observed command results.

- `gh run download 35990838316 -R relux-works/curator --dir "$TMPDIR/m28"` — exit 0; diagnostic artifacts stayed outside the worktree.
- `go test ./internal/install -run '^(TestDraftReplayFileURLUsesPortableGitConfigSyntax|TestDraftFreshMachineReplaysEverySourceThroughInstallAndUpgrade|TestDraftFreshMachineReplaysDeclaredGitSourceWithoutBindings|TestDraftFreshGitSourceWithoutEndpointIsUnavailableAndPreservesLock|TestDraftFreshMachineMovedTagReplaysLockedCommitWithoutRefResolution|TestDraftFreshGitReplayChecksContentHashBeforePublishingSnapshot|TestDraftFreshPathMismatchAndUnreachableSourcePreserveLock|TestDraftInstallReplaysMissingLocalSnapshot|TestDraftLocalMissingSnapshotReplays|TestDraftInstallDoesNotIgnoreOrPrivatizeSkillfileLock)$' -count=1 -timeout=3m` — exit 0 on Darwin (48.427s).
- `env GOOS=windows go vet ./internal/install` — exit 0.
- `golangci-lint run ./internal/install` — exit 0, 0 issues.
- `git diff --check` — exit 0.

The narrowing-mutant table and killed results from the prior revision remain applicable: this revision changes the test fixture only and does not alter the replay gate. No new mutant run was needed for this Windows fixture correction. The release-prep CHANGELOG entry above remains the entry for this work; `CHANGELOG.md` was not changed in this revision per campaign policy.

## Revision 4 — refresh onto a48f584c

Trunk is now `a48f584c` (1kpw4w manifest dependency directory landed). Combined with `git diff 5b326aa3 a48f584c -- . ':!.task-board' ':!CHANGELOG.md' | git apply --3way`: 40 files applied cleanly; 2 conflicted paths kept with both sides.

- `internal/gitops/gitops.go`: trunk added `UnsupportedLinkError` plus a typed symlink refusal in `listTree`; this leaf added `FetchCommitIsolated` / `FetchCommitFromURLIsolated` / `validateLockedCommit` in a non-overlapping region. Kept both via targeted edits. Single definition verified (`grep -c UnsupportedLinkError` = 4: type, method, use + comment).
- `internal/install/draftsources.go`: trunk added directory-aware transitive mapping plus `lockedNetworkRepository`; this leaf added fresh-machine replay (`draftFrozenInput` with sources map, `replayMissingDraftSnapshots`, git/path replay, hash gates). Replaced the transitive block with trunk's directory-aware version adapted to the replay signature and added the helper verbatim. Single definition verified (`lockedNetworkRepository` = 1 def + 1 call); replay entry points still present (`replayMissingDraftSnapshots` = 3).
- `git reset` left nothing staged; `git checkout HEAD -- .task-board` synced the board checkout artifact. Worktree diff is the 12 Revision 3 files (1007 insertions, 84 deletions; +1 vs Revision 3 from the trunk error-message suffix `; capture it with an explicit attempt`).
- `CHANGELOG.md` is byte-identical to trunk (`git show a48f584c:CHANGELOG.md` cmp exit 0). No leaf CHANGELOG edit per campaign policy; release-prep entries above stand.
- `task-board worktree refresh-candidate TASK-260924-m28s6b` → `refresh_advanced`, TrunkOID `a48f584c`, BranchOID `66bc92aa`.

### Revision 4 validation (all run directly as standalone processes; real exit codes)

- `go test ./internal/install -run '^(TestDraftReplayFileURLUsesPortableGitConfigSyntax|TestDraftFreshMachineReplaysEverySourceThroughInstallAndUpgrade|TestDraftFreshMachineReplaysDeclaredGitSourceWithoutBindings|TestDraftFreshGitSourceWithoutEndpointIsUnavailableAndPreservesLock|TestDraftFreshMachineMovedTagReplaysLockedCommitWithoutRefResolution|TestDraftFreshGitReplayChecksContentHashBeforePublishingSnapshot|TestDraftFreshPathMismatchAndUnreachableSourcePreserveLock|TestDraftInstallReplaysMissingLocalSnapshot|TestDraftLocalMissingSnapshotReplays|TestDraftInstallDoesNotIgnoreOrPrivatizeSkillfileLock)$' -count=1 -timeout=3m` — exit 0 (171.115s).
- `go test ./internal/gitops -run '^TestFetchCommitFromURLIsolatedRequestsLockedObjectWithoutTags$' -count=1` — exit 0 (3.489s).
- `go test ./internal/manifest ./internal/closure -count=1` — exit 0 (manifest 4.388s, closure 155.581s).
- `GOOS=windows go vet ./internal/install` — exit 0.
- `go build -o /tmp/curator-m28s6b-rev4 ./cmd/curator` — exit 0.
- `golangci-lint run ./internal/install ./internal/gitops` — exit 0, 0 issues.
- `git diff --check` — exit 0.

The narrowing-mutant table and killed results above remain applicable: this refresh merges trunk's directory support without altering the replay gate, and the replay rows above all drive the production entry points. Mutants were not rerun in this bounded refresh.

## Revision 5 — Windows no-bindings redirect via repo-local config (republish)

The rev3 Change Request gate (run 36020763072) failed only the Windows lane, on exactly the four no-bindings subtests Revision 3 claimed to fix:

- `TestDraftFreshMachineReplaysDeclaredGitSourceWithoutBindings/git-tag/install` + `/upgrade`
- `TestDraftFreshMachineReplaysDeclaredGitSourceWithoutBindings/repository/install` + `/upgrade`

Evidence (from `test-evidence-windows-latest` in `$TMPDIR`, never in the worktree): every subtest failed with `source_snapshot_unavailable: declared Git source for review cannot provide the locked revision`. Root cause was NOT the file-URL syntax Revision 3 fixed. The no-bindings rows redirect the declared `https://example.org/kit.git` endpoint to a local fixture repository through a `#!/bin/sh` git wrapper planted on PATH. On Windows, Go's `os/exec` resolves `git` to the shim but `CreateProcess` cannot execute a script (the toolchain ships no `cmd.exe` wrap), so the rewrite never applied and the declared HTTPS host stayed unreachable. `GIT_CONFIG_GLOBAL`/`GIT_CONFIG_COUNT` redirection cannot substitute: the isolated fetch lane (`isolatedGitEnv`) pins both config files to empty and drops every `GIT_*` name by allow-list (repository-transport §5 isolation, proven by `TestUserConfigIgnoredByResolvedLane`) — weakening that boundary was rejected.

Fix (test fixture only, no production change): `draftReplayGitPath` is replaced by `seedDraftReplaySource`, which pre-creates the exact private replay repository production derives (`sha256(projectAbs + NUL + alias + NUL + identity)[:20]` under `<home>/source-replay`, resolved through the same `loadReplaySourcePolicy` + `ResolveRepositoryEndpoints` calls) and appends a repo-local `[url "<portable file URL>"] insteadOf = <attempt URL>` section to its `.git/config`. Repo-local configuration is honoured by the isolated exact-SHA fetch on every platform; the portable file-URL serialization from Revision 3 is reused unchanged. If the derivation ever drifts, the seed lands elsewhere and the rows fail closed with `source_snapshot_unavailable` — no silent pass. Added `TestDraftReplaySeedSurvivesIsolatedFetch`, which pins the channel directly: seed, then fetch the locked SHA from the declared URL through `gitops.FetchCommitFromURLIsolated` alone.

Production code is untouched in this revision, so the narrowing-mutant table and killed-mutant evidence above stand unchanged (the gates they narrow are byte-identical). No new CHANGELOG entry: the user-facing replay behavior and diagnostics are unchanged since Revision 3; the standing release-prep entries above still apply, and `CHANGELOG.md` remains byte-identical to trunk `a48f584c` per leaf policy.

Refresh state (unchanged): HEAD `66bc92aa` already descends from trunk `a48f584c` (1kpw4w manifest dependency directory); the Revision 3 work sits uncommitted on top, nothing staged, no untracked files. The `git diff 5b326aa3 a48f584c | git apply --3way` step is a no-op against this HEAD (trunk content already present, verified: `lockedNetworkRepository`, `UnsupportedLinkError`, closure `Directory` all present alongside replay entry points); re-applying it refuses with `does not match index`.

### Revision 5 validation (all run directly as standalone processes; real exit codes)

- `go test ./internal/install -run '^(TestDraftReplayFileURLUsesPortableGitConfigSyntax|TestDraftReplaySeedSurvivesIsolatedFetch|TestDraftFreshMachineReplaysEverySourceThroughInstallAndUpgrade|TestDraftFreshMachineReplaysDeclaredGitSourceWithoutBindings|TestDraftFreshGitSourceWithoutEndpointIsUnavailableAndPreservesLock|TestDraftFreshMachineMovedTagReplaysLockedCommitWithoutRefResolution|TestDraftFreshGitReplayChecksContentHashBeforePublishingSnapshot|TestDraftFreshPathMismatchAndUnreachableSourcePreserveLock|TestDraftInstallReplaysMissingLocalSnapshot|TestDraftLocalMissingSnapshotReplays|TestDraftInstallDoesNotIgnoreOrPrivatizeSkillfileLock)$' -count=1 -timeout=5m` — exit 0 (22.621s). All fresh-machine rows per source kind through install and upgrade, moved-tag row, mismatch/unavailable rows, lock-bytes/ignore rows, plus the new channel pin.
- `go test ./internal/gitops -run '^TestFetchCommitFromURLIsolatedRequestsLockedObjectWithoutTags$' -count=1` — exit 0.
- `go test ./internal/manifest ./internal/closure -count=1` — exit 0 (manifest 0.545s, closure 30.740s).
- `go vet ./internal/install` — exit 0.
- `GOOS=windows go vet ./internal/install` — exit 0 (the fixed fixture compiles for the Windows gate).
- `golangci-lint run ./internal/install` — exit 0, 0 issues.
- `go build -o /tmp/curator-m28s6b-rev5 ./cmd/curator` — exit 0.
- `gofmt -l internal/install` — clean; `git diff --check` — exit 0.
- Full `go test ./internal/install -count=1` locally — NOT green: timeout exit 1 at both `-timeout=8m` (480s) and `-timeout=15m` (900s). Zero `--- FAIL` lines in the full log; at panic time the only running test was `TestProjectWriteBlobsSpawnFailureIsWrapped` (0s, unrelated sequential test) with parallel tests queued behind it — suite throughput exceeds the local timeout, not an assertion failure, and no draft/replay test is implicated. The hosted gate runs the same suite with a 120m timeout, so the Change Request gate is the deciding suite for full-package proof. This item stays unchecked.
