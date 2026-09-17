# TASK-260916-55g9dg: manager-refuse-transitive-system-modules

## Description
curator: implement the admission rule in resolution, the waiver in machine configuration, and keep the fragment system-modules flag semantics.

## Scope
curator internal/config (knobs transitive_system_modules, system_module_waivers, lock direction), internal/contextresolve + internal/contextmaterialize (direct-only admission, drop warning, error refusal), internal/envfragment (admitted-set flag), env status posture, vector and schema-case tests from CURATOR_CONFORMANCE_ROOT, CHANGELOG

## Acceptance Criteria
Conformance subset green; a transitive system module fails resolution unless waived
