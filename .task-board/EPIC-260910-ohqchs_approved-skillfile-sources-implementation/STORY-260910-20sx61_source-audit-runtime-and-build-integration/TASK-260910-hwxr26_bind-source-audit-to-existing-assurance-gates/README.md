# TASK-260910-hwxr26: bind-source-audit-to-existing-assurance-gates

## Description
Bind source audit to existing assurance gates. Extend current implementation after inspecting existing outcomes; this task is not authorization to start work.

## Scope
internal/audit, registry, artifactpolicy, scriptpolicy.

## Acceptance Criteria
Validate source-audit-v1 identity/context/policy/evidence/time/decision bindings; keep it distinct from registry attestation. Required network attestation for local packages fails; preserve authorized pins, revocation and assurance checks before cache/compiler. Validate absent, malformed, stale and wrong evidence without weakening currentness.
