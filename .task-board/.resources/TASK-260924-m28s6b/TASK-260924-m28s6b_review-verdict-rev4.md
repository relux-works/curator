# TASK-260924-m28s6b review verdict — CR rev4 (tree 0b97142d) — ACCEPTED

Reviewer: claude-opus-5-5 low. Worktree tree == candidate 0b97142d (temp-index write-tree). Base a48f584c; delta carries checkpointed 1aa9wb default-on + m28s6b replay.

## Rule vs addendum
- Missing snapshot -> replay: `replayMissingDraftSnapshots` (internal/install/draftsources.go:~298) called from `draftFrozenNodes` before `closure.LoadDraftFrozenNodes` (production Project/upgrade path via install.go:394).
- git/repository: `gitops.FetchCommitIsolated` / `FetchCommitFromURLIsolated` fetch only the full locked OID (`--no-tags --no-write-fetch-head`, validateLockedCommit rejects short/non-hex); no ref resolution.
- path: `replayLocalDraftSnapshot` re-captures current bytes; package identity compare then content_sha256 compare -> `source_snapshot_changed`; unreachable -> `source_snapshot_unavailable`.
- Identity check for repository: declared identity != lock -> source_snapshot_changed.
- Lock not written by replay; tests assert lock bytes; gitignore test asserts lock never ignored (draftsources_test.go:772-795).

## Windows changes
Fixture-only: `seedDraftReplaySource` + `draftReplayFileURL` in draftruntime/draftsources tests plant a repo-local insteadOf in the curator-private replay repo. No production Windows branch added. Bound: production isolated fetch honours repo-local config of its own private repo (by design, not user-controlled).

## Independent reruns (darwin, zsh, pipefail)
- `go test -count=1 -run 'Replay|FreshMachine|FreshGit|MissingSnapshot|Gitignore|MovedTag' ./internal/install` -> ok 107.5s
- `go test -count=1 -run FetchCommit ./internal/gitops` -> ok

## Mutant (disposable git-archive copy)
M1: `verifyLockedGitMember` content check `content != member.ContentSHA256` -> `false && ...` : KILLED by TestDraftFreshGitReplayChecksContentHashBeforePublishingSnapshot.

## Hosted gate (rev4 validation log)
All lanes success: Test ubuntu/macos/windows, Race ubuntu/macos, Gate self-test x3, Lint, Naming, Interop conformance.

## Other
CHANGELOG.md untouched in delta (== base); docs state commit Skillfile.lock.json like package-lock.json (README, docs/cli.md, troubleshooting).
Residual: rev5 ("Windows redirect via repo-local config") mentioned in the note was not handed to this run; verdict binds rev4 only.
