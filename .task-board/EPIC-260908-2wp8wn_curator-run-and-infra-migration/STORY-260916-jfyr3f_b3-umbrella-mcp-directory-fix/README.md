# STORY-260916-jfyr3f: b3-umbrella-mcp-directory-fix

## Description
Fix the published umbrella relux-root-context-ivan so its requires.mcp entries resolve: the relux-mcp repository hosts figma and safari under packages/<name>/agent-mcp.json, but the umbrella declares them without a directory, so curator profile install refuses with profile_source_invalid: mcp_declaration_invalid: agent-mcp.json is absent (B5 evidence on TASK-260908-yl5x3k). Add directory to both MCP requirements, bump every package in the repository to 1.0.1 (the repository tag namespace is shared; the resolver requires manifest version == chosen tag version), land through PR and cut signed tag v1.0.1. Then B5 onboarding re-runs.

## Scope
(define story scope)

## Acceptance Criteria
curator profile install git@github.com:relux-works/relux-root-context.git --directory packages/relux-root-context-ivan --range ^1.0 --use --takeover resolves all contexts, skills and MCP members on host e11-1 (verified by the B5 re-run).
