# TASK-260927-31gaka integration preconditions

Date: 2026-09-28. Assignment: confirm landing preconditions for accepted CR revision 2 and attach evidence. This run did not invoke checkpoint or integrate.

## Current state
- Board query confirms TASK-260927-31gaka status integrating (task-board q, exit 0).
- Read-only task-board worktree integrating (exit 0) reports protected refs/heads/main at d8e87bacda3bb4cd9010801156646f75de9354ce; revision 2 is awaiting_landing, candidate_tree_on_trunk=no, repository_delta=present, reason: candidate tree is not carried by any post-base ancestor of refs/heads/main.
- Worktree is clean on task-board/story/STORY-260927-3qf8er at HEAD d8e87bacda3bb4cd9010801156646f75de9354ce (git status and rev-parse, exit 0). The accepted revision 2 snapshot tree is 31e7a36f7b7263f824533e2ae028a74efd5d821b.

## Prior landing evidence
The previously attached TASK-260927-31gaka_story-integrate-run.md records the integration command refusing with exit 1 and code integration_base_moved: trunk advanced with a change to .github/ci/conformance-gaps.tsv, which this Change Request also changes; no one has looked at the combination. That attempt reports no files edited and no status changed. The current read-only integrating view still classifies revision 2 as awaiting_landing.

## Handoff note
The accepted revision is not carried by current protected main, and the earlier landing attempt identified an unreviewed overlap in conformance-gaps.tsv. This evidence is attached for the integration runner and orchestrator. No code changed and no tests or build were run in this precondition-only assignment.