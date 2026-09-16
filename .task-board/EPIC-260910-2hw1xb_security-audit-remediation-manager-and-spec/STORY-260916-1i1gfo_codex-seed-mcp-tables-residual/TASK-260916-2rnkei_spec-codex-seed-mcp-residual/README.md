# TASK-260916-2rnkei: spec-codex-seed-mcp-residual

## Description
curator-spec: decide between stripping mcp_servers from the codex_cli provisioning seed and stating the residual; record the per-adapter MCP-channel asymmetry (claude --strict-mcp-config disables home servers, codex -p layers over the seeded base) as a §7.8 table row; require env status to report ungoverned seeded servers if they are kept.

## Scope
protocol/environments.md §7.4/§7.8, vectors

## Acceptance Criteria
Environments revision merged with the chosen rule, the asymmetry row and vectors
