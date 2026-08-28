# TASK-260819-2g9yz9: integrate-verified-mixed-fleet-policy

## Description
Integrate provider selection and policy behavior across macOS, Linux, Windows, and portable-only hosts.

## Scope
Configuration precedence, organization and project minimum assurance, explicit provider identities, remote or CI fleet reporting, cache portability rules, receipt verification, capability drift, unsupported hosts, and stable diagnostics. Never silently downgrade verified to portable.

## Acceptance Criteria
The same policy produces deterministic results across mixed fleets; verified requirements reject portable or insufficient provider receipts; portable remains the default only where policy permits; caches and checkpoints never alias across assurance or provider identity; integration tests pass.
