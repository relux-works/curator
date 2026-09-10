# STORY-260910-1lf0m5: bound-mcp-declaration-exposure

## Description
Finding S4 (High): MCP declaration packages control launch-time code execution (stdio command+args) and by default receive every operator env var they name (passable_env_names=null, empty mcp_package_allowlist permits all sources). Bound the exposure.

## Scope
curator-spec environments + envregistry/contextpkg + profile install

## Acceptance Criteria
Default passable_env_names is empty or explicitly bounded; empty MCP allowlist produces a loud install-time warning; stdio command+args of the resolved set are surfaced to the operator at profile install/update
