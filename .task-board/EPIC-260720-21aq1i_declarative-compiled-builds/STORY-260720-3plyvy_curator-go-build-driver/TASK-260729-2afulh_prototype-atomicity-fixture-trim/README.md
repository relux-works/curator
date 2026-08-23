# TASK-260729-2afulh: prototype-atomicity-fixture-trim

## Description
Prototype the separately scoped 14-file atomicity fixture reduction on top of the accepted TASK-260729-rfrdfo test-only state by removing the unasserted references/info.md fixture subtree and measuring the real staging-entry/save reduction.

## Scope
Use a task-owned copy of TASK-260729-rfrdfo worktree. Preserve its exact 13-file Patch A/B edits and add only internal/install/atomicity/fixture_test.go as the 14th patch path. Instrument tests only as needed to count before/after StagingEntries and saveJournal calls, without product edits, timeout changes, skips, coverage cuts, or assertion weakening. Focused gates only; use the shared process barrier so no heavy Go run overlaps TASK-260729-365r5r.

## Acceptance Criteria
Literal 14-file allowlist and manifests; references/ directory and references/info.md removal proven assertion-neutral; measured before/after StagingEntries, non-empty file chunks, and saveJournal count for project/global scenarios; focused atomicity non-race and count-one race results with real exits demonstrate <=480 seconds with defensible margin or reject the trim; internal/install Patch A remains green by inherited immutable evidence; independent review before integration.
