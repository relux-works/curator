# TASK-260728-2uxmut: csk-external-snapshot-audit-and-cache

## Description
Treat each admitted external repository snapshot as an independent csk audit subject and protected snapshot/artifact cache input. Implement exact declared/effective identity, build-source digest, policy binding, audit-before-cache/compiler ordering, and deterministic go-repository-v1 build-root containment.

## Scope
Python audit integration, frozen snapshot store, protected cache admission/quarantine, external receipt input preparation, offline protected-snapshot reuse, Go compiler session reuse, and ordering/containment tests. Marker publication and full transaction lifecycle are downstream.

## Acceptance Criteria
The whole external snapshot is validated, hashed, and audited separately from the skill before cache lookup or compile on install, dry-run, cache-hit, repair, and coverage audit; only the selected build root is compiler-visible; inaccessible unprotected source hard-fails before mutation while an exact protected audited snapshot supports the specified offline reinstall; substitutions cannot collide with declared cache identity; corruption is detected and quarantined; existing script and local go-v1 audit/cache behavior remains unchanged; tests pass.
