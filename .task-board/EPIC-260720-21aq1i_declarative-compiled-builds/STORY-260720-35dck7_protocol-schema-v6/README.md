# Protocol schema v6

## Description
Evolve the canonical agent-skill manifest with a strictly declarative compiled-artifact model. The first normative driver is Go. The schema and prose must prevent package-provided shell, arbitrary argv, hooks, plugins, and output-path escapes while preserving schemas 1-5.

## Scope
curator-spec origin/main: protocol core, manager profile, schemas for canonical and legacy manifest names, install marker implications, decision record, changelog/version metadata, positive and negative conformance vectors, vector generator and documentation.

## Acceptance Criteria
A new schema version validates the agreed build declarations; Go driver semantics and install ordering are normative; build sources are excluded from agent context; dry-run and audit-before-build are explicit; compatibility and security impact are recorded; vectors cover valid builds and all key rejection cases; spec validation and deterministic vector regeneration pass.
