# TASK-260822-1yz9ug: module-roots-decision-document

## Description
Write and land the first-party module roots decision in curator-spec decisions/ — pick the next free number against origin/main at landing time, numbering has collided twice before. Content per EPIC-260822-18ylpq notes: build commands gain a declared modules list; the manager validates a bijection between declared module directories and go.mod replace directives — directory form only, strictly inside the snapshot, link-free, disjoint from build and runtime roots, no versions, no module-to-module redirects; declared directories join the directive, cgo, and assembly scan surface; external dependencies stay vendor-only and versioned; snapshot-wide curator-build-source-v1 identity per core.md 8.1 keeps cache keys sound unchanged. Rejected alternatives: reading replace directives as implicit manager input (package-controlled steering of the manager), and requiring repo consolidation (a packaging shape requirement that costs third-party adoption). Land as a decision-only PR to main, squash on green — maintainer pre-authorized 2026-08-22, no human gate.

## Scope
(define task scope)

## Acceptance Criteria
Decision merged to main via squash PR with all required checks green.
