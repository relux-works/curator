# TASK-260728-1jafds: specify-hardened-build-execution-profile

## Description
Define the normative hardened profile and its capability-gated fail-closed contract before any native implementation.

## Scope
curator-spec threat model, process graph, TCB identity, platform capability declarations, cache/receipt/marker/claim separation, stable errors and adversarial conformance vectors. Preserve the portable profile as a distinct weaker contract.

## Acceptance Criteria
Reviewed schemas and vectors express all six hard guarantees without package-controlled executable, argv, environment, paths or hooks; profile identity is bound into every reusable output; unsupported capabilities reject before compiler execution; existing portable schemas and vectors remain compatible.
