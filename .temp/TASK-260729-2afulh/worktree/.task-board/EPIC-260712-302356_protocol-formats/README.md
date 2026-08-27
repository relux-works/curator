# EPIC-260712-302356: protocol-formats

## Description
Phase 1 of docs/implementation-plan.md. Parsing and validation of every protocol format: skill manifest schemas 1 through 5 (Spec 5), project manifest schema 1 (Spec 6.1), dev substitutions (Spec 6.2), user and enforced system configuration with locked keys (Spec 7). Pure functions, typed errors with field paths.

## Scope
See docs/implementation-plan.md, Phase 1. Stories and tasks are created under this epic as work starts.

## Acceptance Criteria
- One table-driven test per MUST clause of Spec 5, 6, 7
- Unknown-field rejection exactly where the spec requires
- Fixture corpus of valid and invalid manifests under testdata/
