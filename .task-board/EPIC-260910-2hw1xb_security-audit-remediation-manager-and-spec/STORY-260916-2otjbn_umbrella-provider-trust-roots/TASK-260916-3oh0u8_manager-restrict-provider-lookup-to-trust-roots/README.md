# TASK-260916-3oh0u8: manager-restrict-provider-lookup-to-trust-roots

## Description
curator: implement provider lookup from trust roots in cmd/curator/umbrella.go with the ownership/writability check and the named diagnostic.

## Scope
curator cmd/curator/umbrella.go (trust-root resolution, revisions A/B with A default, disjoint diagnostics), internal/config (provider_directories knob, lockable), env status posture rows, vector-execution test for conformance/v1/vectors/umbrella-provider-resolution.json, CHANGELOG

## Acceptance Criteria
Conformance subset green; a curator-run planted through a PATH entry outside the trust roots is refused
