# TASK-260910-2c7s0u: service-deployment-rate-limit-docs

## Description
curator-skill-registry: document reverse-proxy rate-limit bucketing (shared network bucket), body-before-auth squatting bounds, and recommended forwarded-header trust in README/DEPLOYING.

## Scope
(define task scope)

## Acceptance Criteria
1. README (or the deployment doc the README delegates to — one place) documents the rate-limit model: which limiter keys on what, that a reverse proxy without trusted forwarded headers collapses all clients into one network bucket, and the recommended configuration citing the exact forwarded-header trust setting/flag/env the service exposes today with its default, with a worked nginx/Caddy/Traefik snippet. 2. The same section documents the body-before-auth squatting bounds (16 MiB, 15 s, 128 slots, 0.1 s acquire; what an attacker can and cannot do; 503 overloaded semantics) and the proxy-side mitigations (body size limits, per-client connection limits), plus a short operator checklist. 3. SECURITY.md carries one paragraph pointing to that section from the threat model; CHANGELOG has an Unreleased R4 entry. 4. Every documented limit/setting matches the code (file:line in the results); a mismatch is reported in the results, never patched in code. 5. The test suite is unchanged and green (transcript in TASK-260910-2c7s0u_results.md); no results/logbook files inside the worktree.
