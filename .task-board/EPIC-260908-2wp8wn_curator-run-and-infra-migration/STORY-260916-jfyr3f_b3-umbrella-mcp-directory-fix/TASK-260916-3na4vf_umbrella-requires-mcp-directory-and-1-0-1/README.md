# TASK-260916-3na4vf: umbrella-requires-mcp-directory-and-1-0-1

## Description
In relux-root-context: (1) packages/relux-root-context-ivan/agent-context.json requires.mcp.figma gets directory packages/figma and requires.mcp.safari gets directory packages/safari (git and range unchanged); (2) bump version to 1.0.1 in all six packages/*/agent-context.json (core, workflow, style, claude, attachments, ivan) and update any README/test fixture that pins 1.0.0; (3) scripts/validate.sh and tests pass. No other changes.

## Scope
(define task scope)

## Acceptance Criteria
agent-context.json of the umbrella declares directory for both MCP members; all six package versions are 1.0.1; bash scripts/validate.sh exits 0; python tests under tests/ pass; the diff touches only manifests, README/tests that pin the version.
