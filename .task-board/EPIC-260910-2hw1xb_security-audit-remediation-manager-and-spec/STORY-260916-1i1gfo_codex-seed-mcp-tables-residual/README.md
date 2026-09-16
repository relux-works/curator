# STORY-260916-1i1gfo: codex-seed-mcp-tables-residual

## Description
Finding E3 (Medium): environments §7.4 seeds codex_cli with the native config.toml copied whole, MCP tables included, and §7.8 layers the profile set over it with -p curator-mcp. A managed codex home therefore runs every native MCP server the operator ever configured, outside the profile lock and outside the §2.2 allowlist, while claude_code runs only the profile set under --strict-mcp-config. The asymmetry is not stated.

## Scope
curator-spec environments §7.4/§7.8; curator envprofile provisioning seeds, env status

## Acceptance Criteria
Either the codex seed strips mcp_servers (seeding only trust/model/tui members) or §7.4 and §7.8 state the residual and env status reports the ungoverned seeded entries as §7.6 does for Xcode targets; §7.8 carries the per-adapter asymmetry as a table row; vectors updated
