# TASK-260822-96m5pj: config-build-ssh-field

## Description
Add the build_ssh top-level field to the global config (internal/config): map of scope -> {agent?, identity?, known_hosts?}. Scope grammar: lowercase host, optional path segments [A-Za-z0-9._-]+, matching only on whole '/' boundaries; at least one of agent/identity per entry; fail closed on anything else. Provide a match helper (longest segment-prefix wins) and serialization. Extend managerKeys. Reference: canonical repository identity of spec section 6.3. Do not name other manager implementations anywhere.

## Scope
(define task scope)

## Acceptance Criteria
Parse/serialize roundtrip; grammar rejections tested (empty segment, uppercase host, boundary bleed 'portals' vs 'portals-evil'); longest-prefix match tested; unknown entry fields rejected; go test ./internal/config green.
