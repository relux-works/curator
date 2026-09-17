# TASK-260910-gocke2: manager-mcp-install-surfacing

## Description
curator: surface the resolved MCP set (command, args, env_names) at profile install/update and warn when the allowlist is empty or when env_names intersects operator-secret-looking variables.

## Scope
curator internal/envfragment + internal/envprofile (s4-warn/s4-enforce passthrough profiles, default s4-warn), internal/config (empty-allowlist warning), cmd/curator profile install/update surfacing rows (§2.3) and env status posture, vector-execution test for environments-env-passthrough.json, CHANGELOG

## Acceptance Criteria
Install output lists stdio commands; warnings implemented; tests cover both
