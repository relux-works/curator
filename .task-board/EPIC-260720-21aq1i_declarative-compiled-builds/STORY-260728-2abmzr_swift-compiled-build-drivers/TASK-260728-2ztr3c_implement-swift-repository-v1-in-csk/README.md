# TASK-260728-2ztr3c: implement-swift-repository-v1-in-csk

## Description
Independently implement external swift-repository-v1 in Python csk/CocoaSkills.

## Scope
External schema and target models, clean snapshot acquisition/audit, trusted Swift worker, cache/receipt/marker lifecycle, atomic activation, rollback/status/repair/GC and native tests.

## Acceptance Criteria
csk matches Curator for every external Swift valid, cache, offline, forbidden-feature and rollback case without invoking Curator internals; Python and native gates pass.
