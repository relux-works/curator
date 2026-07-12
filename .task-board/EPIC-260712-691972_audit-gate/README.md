# EPIC-260712-691972: audit-gate

## Description
Phase 8 of docs/implementation-plan.md. Minimal conforming audit layer: decision semantics per Spec 12.2 (allow, warn, block, require_pin), local revocations by hash and source glob, operator pins, blocking canary, a small deterministic detector set, verdict cache keyed per Spec 12.1.

## Scope
See docs/implementation-plan.md, Phase 8. Stories and tasks are created under this epic as work starts.

## Acceptance Criteria
- Decision table tests over mode and fail_on
- Canary failure blocks
- Cache hit skips analysis but re-decides under current policy
