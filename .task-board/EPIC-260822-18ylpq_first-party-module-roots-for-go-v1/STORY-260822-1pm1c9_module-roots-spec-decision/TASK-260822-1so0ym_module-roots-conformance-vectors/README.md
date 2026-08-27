# TASK-260822-1so0ym: module-roots-conformance-vectors

## Description
Positive and negative conformance vectors for declared module roots: acceptance of a valid declaration; rejection of escape paths, module-to-module redirects, undeclared replace directives, unused declarations, nested modules, runtime-root overlap, and Windows path collisions. Regenerate twice and prove an identical second run. Fully autonomous per the maintainer pre-authorization of 2026-08-22.

## Scope
(define task scope)

## Acceptance Criteria
Vectors committed; double regeneration proven clean; spec CI green on the branch.
