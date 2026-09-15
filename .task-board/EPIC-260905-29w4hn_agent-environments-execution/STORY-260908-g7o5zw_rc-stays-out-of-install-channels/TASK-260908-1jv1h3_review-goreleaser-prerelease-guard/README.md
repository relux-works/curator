# TASK-260908-1jv1h3: review-goreleaser-prerelease-guard

## Description
Review PR #65 at d1bb0a4e: three .goreleaser.yml fields to auto. Branch chore/rc-stays-out-of-install-channels in worktree /Users/iv/Developer/ReluxWorks/.worktrees/curator-release-channels, base 04550e28.

## Scope
curator repository, branch chore/rc-stays-out-of-install-channels, base 04550e28, head d1bb0a4e. Files: .goreleaser.yml (three fields to auto).

## Acceptance Criteria
The three fields switch install-channel publishing to auto; semantics auto-equals-skip-on-prerelease is stated with docs-confidence labelled; nothing else in the release pipeline changes; verdict names the docs-confidence boundary and that first real proof is the rc1 tag itself.
