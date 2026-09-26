# TASK-260923-em42lw: emit-launch-env-fragment-v2

## Description
Make curator env resolve --format json emit launch-env-fragment-v2 (curator-spec ec8dc656, environments §10.1/§10.2/§12.5): the REQUIRED closed member permissions {mode: native|yolo, locked: bool, source: profile|global|default} computed for the resolved profile from the effective permissions.<profile> knob (manager-config-v2 §12.1) with the §12.2 system-file force-native lock applied (manager §1 warning). Lattice: locked iff source=global; mode=native whenever source is global or default; source=default means the profile level is silent.

## Scope
curator: internal/config (permissions knob parse + lock), internal/envfragment (v2 emission), cmd/curator env resolve, tests, docs/environment-config.md, CHANGELOG. No launcher change.

## Acceptance Criteria
1) the permissions.<profile> knob parses per manager-config-v2 (valid/invalid rows) and the system lock forces native with the manager §1 warning; 2) env resolve emits v2 with the member for: knob yolo, knob native, knob absent (source default), lock engaged (source global, locked true, mode native) — production-entry rows; 3) v2 output validates against schemas/v1/launch-env-fragment-v2.schema.json from curator-spec main; 4) one narrowing mutant per lattice rule killed; 5) CHANGELOG
