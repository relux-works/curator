# EPIC-260712-978324: scopes-and-gc

## Description
Phase 5 of docs/implementation-plan.md. Global scope with home-level adapters (Spec 9.2), hybrid scope with targets, shadowing, machine-locale store rendering and reachability split (Spec 9.3), consumers registry and runtime garbage collection (Spec 8.7).

## Scope
See docs/implementation-plan.md, Phase 5. Stories and tasks are created under this epic as work starts.

## Acceptance Criteria
- Shadowing order project over hybrid over global proven by tests
- GC never deletes runtime referenced by any registered consumer
- Hybrid store renders once per machine
