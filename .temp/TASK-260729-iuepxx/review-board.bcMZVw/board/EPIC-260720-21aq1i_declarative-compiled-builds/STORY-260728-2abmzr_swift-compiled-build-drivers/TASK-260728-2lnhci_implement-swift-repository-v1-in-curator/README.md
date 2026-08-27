# TASK-260728-2lnhci: implement-swift-repository-v1-in-curator

## Description
Implement external swift-repository-v1 builds in Go Curator on top of exact Git snapshot audit and atomic publication.

## Scope
skill-build.json target models, external source identity and audit-before-cache/compiler, fixed offline Swift worker, protected snapshots/artifacts, mixed receipts/markers, dry-run, install/rollback/currentness and native tests.

## Acceptance Criteria
Curator accepts only locked audited Swift repositories and the closed target contract; inaccessible or forbidden source fails before mutation; valid builds, cache hits, rollback and platform behavior match shared vectors.
