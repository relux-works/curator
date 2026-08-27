# STORY-260712-524369: identifiers-and-errors

## Description
Identifier and source-path validation (Spec 5.2, 6.1) and the typed error foundation: errors carry the offending field path and a stable code.

## Acceptance Criteria
- Identifier regexp matches Spec 5.2 exactly
- Source path rule: every segment a valid identifier
- Errors expose field path programmatically
