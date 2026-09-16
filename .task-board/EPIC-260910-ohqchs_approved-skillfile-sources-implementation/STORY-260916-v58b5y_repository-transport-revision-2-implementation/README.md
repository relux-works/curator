# STORY-260916-v58b5y: repository-transport-revision-2-implementation

## Description
Implement repository transport revision 2 (curator-spec main 8ba9c23, protocol/repository-transport.md §§4–7) on top of the revision 1 executor from STORY-260910-1bhj0g: source-policy schema 2 as an additive superset (port-bearing endpoint and pin URLs, mirror_of attestation, operator host-alias table), canonical host/path remains the only portable identity, bounded fail-closed resolution with the two new failure classes and attempt bounds, provenance and secrets rules of §7. Draft/opt-in; frozen v1 schema untouched.

## Scope
(define story scope)

## Acceptance Criteria
Schema 2 policies load and validate (schema 1 still accepted unchanged); ports, mirrors and aliases resolve exactly per §5–§6 with the attempt bounds and failure classes; negative tests for every refusal row; goldens and docs updated; narrow package tests green; remote gate green.
