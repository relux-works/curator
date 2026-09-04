# STORY-260901-2zdg81: csk-identifier-cleanup

## Description
Surface-level csk naming uniformity across curator and curator-spec — operator-narrowed scope (2026-09-01): prose, documentation, and diagnostics stop spelling csk except where they name a frozen §1.1 wire identifier (.csk-install.json, .csk-managed.json, CSK_PROJECT_ROOT — these stay exactly as they are; NO retirement, alias migration, or deprecation work). Every new identifier joins the agent-* family (.agent-environment.json is the precedent from Decision 0010).

## Scope
(define story scope)

## Acceptance Criteria
Inventory of csk occurrences in curator and curator-spec prose/docs/diagnostics (wire identifiers excluded); non-wire occurrences rewritten to uniform agent-* / neutral wording; no schema, vector, or wire-identifier change; CLI human output checked for stray csk spellings.
