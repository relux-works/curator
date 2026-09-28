# STORY-260928-lpnvkn: global-operation-lock-publication-ordering

## Description
Operator decision 2026-09-28 on the TASK-260906-1xbrz6 decision packet: alternative 1. For global add / global install, the extended profile lock is published before in-place materialization, and on an environment_surface_unmanaged_conflict the published lock is kept, so profile sync --takeover / profile use --takeover can materialize it before a retry. The five-operation takeover carrier set stays closed.

## Scope
(define story scope)

## Acceptance Criteria
spec §9.4 + manager transaction rules state the ordering and conflict behaviour; curator conforms with a production-entry row
