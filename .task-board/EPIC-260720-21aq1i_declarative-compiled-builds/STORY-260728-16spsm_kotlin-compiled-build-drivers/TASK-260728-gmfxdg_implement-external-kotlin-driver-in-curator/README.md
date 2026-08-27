# TASK-260728-gmfxdg: implement-external-kotlin-driver-in-curator

## Description
Implement the selected external Kotlin repository driver in Go Curator on top of exact source audit and atomic publication.

## Scope
skill-build.json target models, locked repository source identity, audit-before-cache/compiler, trusted compiler/runtime worker, protected snapshot/artifact or bundle cache, mixed receipts/markers, dry-run, rollback/currentness and native tests.

## Acceptance Criteria
Curator accepts only the closed selected Kotlin target and exact audited source; inaccessible or forbidden inputs fail before mutation; valid build, cache, launch and rollback behavior matches shared vectors.
