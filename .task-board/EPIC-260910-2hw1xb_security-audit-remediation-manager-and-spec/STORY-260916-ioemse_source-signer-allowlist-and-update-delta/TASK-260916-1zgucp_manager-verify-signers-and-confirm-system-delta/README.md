# TASK-260916-1zgucp: manager-verify-signers-and-confirm-system-delta

## Description
curator: enforce the signer allowlist during resolution (contextresolve/contextlock), print the resolved-version delta on profile update and refuse without --confirm-system-delta (or equivalent) when the delta touches class: system modules or MCP declarations; report the posture in env status.

## Scope
curator internal/contextresolve, internal/contextlock, cmd/curator profile update, env status

## Acceptance Criteria
Conformance subset green; an unsigned or wrong-signer candidate is refused with the specified diagnostic; the delta confirmation is exercised by a golden
