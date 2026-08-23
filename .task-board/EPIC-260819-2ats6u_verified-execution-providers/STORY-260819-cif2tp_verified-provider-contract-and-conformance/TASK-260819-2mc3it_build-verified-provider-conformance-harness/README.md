# TASK-260819-2mc3it: build-verified-provider-conformance-harness

## Description
Build a reusable black-box and fault-injection suite every platform provider must pass.

## Scope
Exercise undeclared processes, reads, writes, environment, network, resource escape, TOCTOU races, provider crash and restart, event loss and reordering, partial enforcement, IPC spoofing, receipt tampering, reboot, upgrade, rollback, and revocation. Produce machine-verifiable evidence bundles.

## Acceptance Criteria
The harness has deterministic positive and negative vectors, detects intentionally faulty reference providers, runs platform adapters through one required suite plus platform extensions, and emits immutable evidence consumable by the release gate.
