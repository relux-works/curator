# TASK-260917-8vfgxf: manager-interim-nul-opaque-rule

## Description
curator: until v2 lands, treat any regular file containing 0x00 in a skill or context snapshot as a blocking opaque audit finding regardless of directory, with the two colliding trees as tests.

## Scope
internal/contextaudit, internal/audit detectors, vectors

## Acceptance Criteria
Both colliding trees produce a blocking finding; conformance subset green
