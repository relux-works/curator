# STORY-260916-33vuzm: launch-plane-minor-residuals

## Description
Minor findings from the environments/launch plane supplement: (a) the launcher configuration family (defaults.json, ax.json under /etc/curator-run and $XDG_CONFIG_HOME/curator-run) is schema-validated but not ownership-validated, so a user-writable machine file or a symlinked operator file flips tracking policy or locks; (b) env resolve --repair on every launch re-materializes from the store under the launcher authority, so under S5 a tampered store is re-applied on every launch; (c) the --strict-mcp-config residual for claude_code and its codex inverse (E3) should be one table row in §7.8.

## Scope
curator-agent-launcher SPEC §4.3/§4.6/§4.7; curator-spec environments §7.8/§10.1 and the S5 remediation

## Acceptance Criteria
Launcher SPEC and implementation validate ownership/permission of both configuration files and refuse symlinked or foreign-writable files with a named diagnostic; the S5 remediation names repair-as-persistence; §7.8 carries the asymmetry row
