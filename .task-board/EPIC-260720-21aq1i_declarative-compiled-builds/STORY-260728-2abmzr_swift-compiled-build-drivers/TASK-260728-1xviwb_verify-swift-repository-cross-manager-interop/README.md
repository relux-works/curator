# TASK-260728-1xviwb: verify-swift-repository-cross-manager-interop

## Description
Verify external swift-repository-v1 behavior across Curator and csk on supported macOS and Windows toolchains.

## Scope
Locked repositories, exact-tag and offline cases, audit-before-cache/compiler, forbidden descriptor/package features, artifact/receipt equality, rollback/concurrency and PATH launch.

## Acceptance Criteria
Both managers match all shared external Swift results and native evidence is reproducible; no unsupported or red tuple is claimed.
