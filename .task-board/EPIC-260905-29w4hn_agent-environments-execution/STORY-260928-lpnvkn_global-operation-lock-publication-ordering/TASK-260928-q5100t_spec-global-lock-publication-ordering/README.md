# TASK-260928-q5100t: spec-global-lock-publication-ordering

## Description
curator-spec: state in environments.md §9.4 (and the manager transaction rules) that global add / global install publish the extended profile lock before in-place materialization and keep it when a surface write meets environment_surface_unmanaged_conflict, so the §9.4/§9.5 recovery (profile sync --takeover or profile use --takeover, then retry) is true; add conformance vector(s) for the ordering and the preserved lock; CHANGELOG under Unreleased.

## Scope
(define task scope)

## Acceptance Criteria
spec text + vectors merged to curator-spec main via a squash PR; validators green
