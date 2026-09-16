# TASK-260916-yvxbs1: manager-validate-path-source-boundary

## Description
curator: refuse MCP declarations and system modules from path-kind packages per the rule and validate path overlay directories with the S5 boundary contract.

## Scope
curator internal/contextpkg path sources, internal/contextresolve

## Acceptance Criteria
Conformance subset green; a path-kind package with an MCP declaration or a system module is refused with the specified diagnostic
