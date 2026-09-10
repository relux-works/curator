# TASK-260910-gocke2: manager-mcp-install-surfacing

## Description
curator: surface the resolved MCP set (command, args, env_names) at profile install/update and warn when the allowlist is empty or when env_names intersects operator-secret-looking variables.

## Scope
(define task scope)

## Acceptance Criteria
Install output lists stdio commands; warnings implemented; tests cover both
