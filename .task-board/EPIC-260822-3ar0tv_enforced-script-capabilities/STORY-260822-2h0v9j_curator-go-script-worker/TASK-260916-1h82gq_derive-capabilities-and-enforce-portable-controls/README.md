# TASK-260916-1h82gq: derive-capabilities-and-enforce-portable-controls

## Description
R2: connect declaration-derived environment, PATH, working directory/private roots, exec grants, offline configuration and secret identifier handling to the R1 worker path; drive all four derivation cases and all eleven mandatory portable controls from the script-worker-v1 vectors. Deny by default.

## Scope
(define task scope)

## Acceptance Criteria
All four capability-derivation cases and all eleven mandatory controls pass at the production entry with narrowing mutants per control.
