# TASK-260916-55g9dg: manager-refuse-transitive-system-modules

## Description
curator: implement the admission rule in resolution, the waiver in machine configuration, and keep the fragment system-modules flag semantics.

## Scope
curator internal/contextresolve, internal/config, internal/envfragment

## Acceptance Criteria
Conformance subset green; a transitive system module fails resolution unless waived
