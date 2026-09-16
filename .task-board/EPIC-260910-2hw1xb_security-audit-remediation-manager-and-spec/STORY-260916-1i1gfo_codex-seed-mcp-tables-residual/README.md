# STORY-260916-1i1gfo: codex-seed-mcp-tables-residual

## Description
Finding E3 (Medium): environments §7.4 seeds codex_cli with the native config.toml copied whole, MCP tables included, and §7.8 layers the profile set over it with -p curator-mcp. A managed codex home therefore runs every native MCP server the operator ever configured, outside the profile lock and outside the §2.2 allowlist, while claude_code runs only the profile set under --strict-mcp-config. The asymmetry is not stated.

Implementation verification (TASK-260916-dv7xv5 rev3, curator main 80483355, launcher main b34e1e27, static): confirmed. Seeds: []string{"config.toml"} (envregistry.go:219); gatherSeeds reads the native file whole (managed.go:570-584) and the plan writes the bytes unchanged at provisioning (:887-893); the other mcp_servers paths are the profile layer renderer (contextmaterialize/mcp.go:231-240) and the read-only home verifier (mcp/mcp.go:253), neither touches the seed. Strip or report belongs in gatherSeeds and the registry seed row.

## Scope
curator-spec environments §7.4/§7.8; curator envprofile provisioning seeds, env status

## Acceptance Criteria
The codex seed strips mcp_servers at provisioning (seeding only trust, model and tui members); §7.4 and §7.8 state the rule and the per-adapter MCP-channel asymmetry as a table row; env status lists the native entries that were not inherited; the manager README/CHANGELOG document that a managed codex home runs only the profile MCP set and native ~/.codex/config.toml servers are not inherited; vectors updated
