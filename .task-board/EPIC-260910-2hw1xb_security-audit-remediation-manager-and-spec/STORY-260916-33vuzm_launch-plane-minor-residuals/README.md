# STORY-260916-33vuzm: launch-plane-minor-residuals

## Description
Minor findings from the environments/launch plane supplement: (a) the launcher configuration family (defaults.json, ax.json under /etc/curator-run and $XDG_CONFIG_HOME/curator-run) is schema-validated but not ownership-validated, so a user-writable machine file or a symlinked operator file flips tracking policy or locks; (b) env resolve --repair on every launch re-materializes from the store under the launcher authority, so under S5 a tampered store is re-applied on every launch; (c) the --strict-mcp-config residual for claude_code and its codex inverse (E3) should be one table row in §7.8.

Implementation verification (TASK-260916-dv7xv5 rev2, curator main 80483355, launcher main b34e1e27, static): confirmed for all three minors. Both launcher loaders follow a symlinked configuration file instead of refusing it (axconfig/config.go:58-74, defaults/defaults.go:108-118) and nothing checks uid/gid/permissions (only execution.go:197 checks the executable bit); --repair is unconditional (fragment/resolve.go:75-80) so repair-as-persistence is S5's residual; the strict-MCP asymmetry is real: --mcp-config with --strict-mcp-config for claude_code and -p curator-mcp for codex_cli (launcher fragment.go:172-173, curator envregistry.go:192/:215) combined with the seeded native mcp_servers (E3).

## Scope
curator-agent-launcher SPEC §4.3/§4.6/§4.7; curator-spec environments §7.8/§10.1 and the S5 remediation

## Acceptance Criteria
Launcher SPEC and implementation validate ownership/permission of both configuration files and refuse symlinked or foreign-writable files with a named diagnostic; the S5 remediation names repair-as-persistence; §7.8 carries the asymmetry row
