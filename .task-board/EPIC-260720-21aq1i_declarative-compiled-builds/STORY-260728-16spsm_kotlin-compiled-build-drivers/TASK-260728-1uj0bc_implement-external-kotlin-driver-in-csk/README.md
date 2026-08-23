# TASK-260728-1uj0bc: implement-external-kotlin-driver-in-csk

## Description
Independently implement the selected external Kotlin repository driver in Python csk/CocoaSkills.

## Scope
External schema and target models, clean source acquisition/audit, trusted compiler/runtime worker, artifact/bundle cache and lifecycle, atomic activation, rollback/status/repair/GC and platform tests.

## Acceptance Criteria
csk matches Curator for every external Kotlin valid, cache, offline, forbidden-feature and rollback case without invoking Curator internals; Python and native gates pass.
