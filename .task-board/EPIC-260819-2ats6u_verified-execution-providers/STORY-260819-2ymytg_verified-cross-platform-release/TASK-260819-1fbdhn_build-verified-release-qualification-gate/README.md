# TASK-260819-1fbdhn: build-verified-release-qualification-gate

## Description
Build the release gate that consumes common and platform conformance evidence.

## Scope
Require exact provider versions and signatures, common capability vectors, platform matrices, threat-model traceability, lifecycle tests, upgrade and rollback evidence, revocation status, reproducible artifacts, known limitations, and independent reviews. Reject missing, stale, mixed-version, or self-attested evidence.

## Acceptance Criteria
The gate rejects intentionally incomplete or tampered evidence and accepts only a coherent macOS, Linux, and Windows release set satisfying the same mandatory verified capability version. Gate output is deterministic and auditable.
