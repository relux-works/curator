# TASK-260916-1ihonr: launcher-config-family-ownership-check

## Description
curator-agent-launcher: validate ownership and permissions of defaults.json and ax.json (machine and operator files), refuse symlinked or foreign-writable files with a named diagnostic; SPEC §4.7 gains the contract.

## Scope
curator-agent-launcher SPEC §4.3/§4.6/§4.7, cmd/curator-run configuration loading

## Acceptance Criteria
SPEC revision and implementation merged; goldens for a symlinked and a group-writable configuration file
