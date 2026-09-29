# STORY-260928-7eowfl: path-boundary-walk-vanishing-entry

## Description
Flake found on 3i6vod gate run 36467710280 (macOS): internal/envprofile TestPathInstallCapturesDirtyUntrackedInsideGit fails with profile_source_invalid regular_types boundary check failed at root/.git/objects/maintenance.lock: lstat … no such file or directory. The E6 path-source boundary walk (internal/pathboundary) readdir-then-lstat races with an entry that disappears (git background maintenance).

## Scope
(define story scope)

## Acceptance Criteria
the walk is deterministic under concurrent removal (decided semantics per environments §4 with a row that removes an entry between readdir and lstat); the test fixture disables git auto-maintenance; no flake in 20 local repetitions
