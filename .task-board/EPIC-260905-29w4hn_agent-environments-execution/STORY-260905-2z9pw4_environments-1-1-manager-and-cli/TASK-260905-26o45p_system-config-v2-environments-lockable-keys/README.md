# TASK-260905-26o45p: system-config-v2-environments-lockable-keys

## Description
Deferred from the batch-3 inconsistencies: environments §12.2 extends the manager §1 locked set by the environments keys (overlays_allowed, precedence, mcp_package_allowlist, passable_env_names, require_current_profile, isolation), but system-config-v1.schema.json closes locked to the four schema-1 keys and has no environments member, so no system file can lock them today. Deliver system-config schema 2 (additive; environments member with exactly the §12.2 keys), schema cases and vectors through the generator, manager §1 text, COMPATIBILITY and CHANGELOG notes. Schedule with implementation stage (c) config work; keep any consumed-by-pin vector file byte-identical (new family file if needed).

## Scope
(define task scope)

## Acceptance Criteria
system-config-v2 schema, cases and vectors land after an independent ACCEPT; manager §1 names the environments lockable keys; make validate and regenerate-check green; Implementations lane green.
