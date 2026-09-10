# TASK-260910-3p2rbh: service-health-cached-verdict

## Description
curator-skill-registry: /health returns a cached integrity verdict refreshed by a background full verifier; a failed refresh makes /health non-ready.

## Scope
(define task scope)

## Acceptance Criteria
Health no longer runs the full chain per probe; corruption still flips readiness
