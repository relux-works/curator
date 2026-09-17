# TASK-260916-hxr6qv: revision-2-bounded-resolution-and-provenance

## Description
Resolution over ports/mirrors/aliases with §6 attempt bounds and the two new failure classes; §7 secrets, provenance and compatibility rules; end-to-end tests through the production entry point.

## Scope
(define task scope)

## Acceptance Criteria
Resolution over ports, mirrors and host aliases follows repository-transport.md §6 exactly (attempt bounds, the two new failure classes) and §7 (secrets, provenance, compatibility) through the production entry point with end-to-end tests and negative rows for each refusal; legacy and revision 1 behaviour unchanged (goldens); narrow tests and remote gate green.
