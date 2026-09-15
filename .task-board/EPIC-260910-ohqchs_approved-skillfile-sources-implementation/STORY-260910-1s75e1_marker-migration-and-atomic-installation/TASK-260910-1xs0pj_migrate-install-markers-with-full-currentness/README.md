# TASK-260910-1xs0pj: migrate-install-markers-with-full-currentness

## Description
Migrate install markers with full currentness. Extend current implementation after inspecting existing outcomes; this task is not authorization to start work.

## Scope
internal/marker, audit, install status/currentness readers.

## Acceptance Criteria
Implement marker v5 and complete 25-field v4 migration. Preserve Git attestation and legacy substituted semantics where applicable; reject those fields for invalid local arms. New selectors cannot enable legacy substitution. Verify package/lock binding and all marker mismatch/evidence cases; summaries never authorize execution.
