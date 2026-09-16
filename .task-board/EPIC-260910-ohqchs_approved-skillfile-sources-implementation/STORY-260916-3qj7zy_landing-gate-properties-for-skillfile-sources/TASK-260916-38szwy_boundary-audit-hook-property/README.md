# TASK-260916-38szwy: boundary-audit-hook-property

## Description
Install an audit hook once per test process, generate source trees from the traversal families, run the expansion, assert no filesystem event names a path outside the resolved root. The oracle must not trust the walker.

## Scope
gate property suite for source expansion

## Acceptance Criteria
Property green in the gate over the eight traversal families; a planted out-of-root read fails it
