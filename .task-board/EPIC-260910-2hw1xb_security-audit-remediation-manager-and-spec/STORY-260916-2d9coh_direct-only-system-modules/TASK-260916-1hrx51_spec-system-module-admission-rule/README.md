# TASK-260916-1hrx51: spec-system-module-admission-rule

## Description
curator-spec: restrict class: system modules to packages named directly by the root or by a machine-config allowlist; add the resolution error context_system_module_transitive and its scoped waiver; keep context-system-module-present as the always-warn finding; vectors for direct, transitive and waived cases.

## Scope
protocol/environments.md §2/§12/§13, schemas, conformance vectors

## Acceptance Criteria
Environments revision merged with the admission rule, diagnostic, waiver and vectors
