# TASK-260916-33abdk: manager-strip-or-report-seeded-mcp-servers

## Description
curator: implement the chosen rule for the codex_cli seed (strip mcp_servers at provisioning, or record and report ungoverned entries in env status).

## Scope
curator internal/envprofile provisioning seeds, env status

## Acceptance Criteria
Conformance subset green; a native config.toml with mcp_servers provisions per the rule and env status reports accordingly
