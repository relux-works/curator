# TASK-260910-2ohnjo: spec-bound-env-passthrough

## Description
curator-spec: change the default passable_env_names from null (unbounded) to empty/opt-in, require a loud warning when mcp_package_allowlist is empty, and specify operator surfacing of stdio command+args at profile install.

## Scope
curator-spec protocol/environments.md §2.2, §10.3, §12.1/§12.2, diagnostics tables, conformance vectors for the new defaults and surfacing output, CHANGELOG

## Acceptance Criteria
Environments revision merged with conformance vectors for the new defaults
