# TASK-260728-3ar1qp: implement-local-kotlin-driver-in-csk

## Description
Independently implement the selected local Kotlin vendored-skill driver in Python csk/CocoaSkills.

## Scope
Local build roots and context exclusion, trusted compiler/runtime preflight, fixed offline worker, artifact/bundle cache, receipt/marker/launcher integration, atomic activation, rollback/currentness and platform tests.

## Acceptance Criteria
csk matches Curator and shared local Kotlin vectors without arbitrary build scripts, preserves existing drivers, and passes pytest, mypy, lint and supported native gates.
