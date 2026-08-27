# TASK-260728-3jaa57: csk-mixed-receipt-marker-and-lifecycle

## Description
Extend csk project/global planning and transactions to publish receipt v2 for external commands, receipt v1 for local go-v1 commands, marker v3 for mixed installs, manager-derived activation shims, and complete currentness, repair, rollback, recovery, deduplication, and GC behavior.

## Scope
Python planner and canonical serialization, protected cache/journal integration, project/global install and uninstall, command collision/PATH activation, signer-policy gate, status/repair/GC, crash recovery, and tests. Platform release notarization and Authenticode remain outside installation.

## Acceptance Criteria
Canonical receipt-v2 and marker-v3 bytes match spec/shared vectors; local and external receipt schemas never alias; successful publication is gated by exact source proof, audit, deterministic build, and artifact verification; manager-derived names/paths, executable/hash/link-count/shim checks, collision prevention, atomic rollback, status, repair, and GC behave identically to normative outcomes while retaining csk-specific home layout; credentials/signing never come from package data; project/global and failure-injection tests pass.
