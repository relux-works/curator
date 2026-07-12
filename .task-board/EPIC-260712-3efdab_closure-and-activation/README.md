# EPIC-260712-3efdab: closure-and-activation

## Description
Phase 3 of docs/implementation-plan.md. Closure resolution: expansion from direct declarations, unification to one commit and one identity, conflict errors carrying requirement chains, cycle detection, deterministic provider-first order, edge-based activation with command narrowing, requirement-command validation, active-command collisions (Spec 8.3, 8.4).

## Scope
See docs/implementation-plan.md, Phase 3. Stories and tasks are created under this epic as work starts.

## Acceptance Criteria
- Graph fixtures: diamond, cycle, version conflict, source conflict, narrowing
- Errors name both requirement chains
- Ordering deterministic across runs and platforms
