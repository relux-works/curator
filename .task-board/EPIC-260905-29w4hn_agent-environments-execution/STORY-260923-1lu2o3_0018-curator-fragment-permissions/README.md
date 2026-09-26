# STORY-260923-1lu2o3: 0018-curator-fragment-permissions

## Description
Curator side of Decision 0018: emit launch-env-fragment-v2 with the permissions member fixed by F-S2 (curator-spec ec8dc656).

## Scope
curator: manager-config-v2 permissions knob, system-config lock direction, env resolve fragment emission, tests, docs.

## Acceptance Criteria
env resolve emits launch-env-fragment-v2 with a correct permissions member for every knob/lock combination; landed.
