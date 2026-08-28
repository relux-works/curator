# TASK-260819-g6lw7c: qualify-macos-security-primitives-and-entitlements

## Description
Prove the macOS mechanism composition and support matrix before implementation.

## Scope
Map each common verified capability to Endpoint Security authorization and notification events, Network Extension filtering, process containment, filesystem and resource controls, entitlement and Full Disk Access requirements, OS-version behavior, event muting and loss semantics, extension activation, and Apple distribution constraints.

## Acceptance Criteria
A reviewed feasibility matrix identifies exact APIs and minimum macOS versions for every mandatory guarantee, demonstrates how event loss and provider failure are detected before acceptance, documents entitlement acquisition risk, and marks any unprovable guarantee unsupported rather than approximated.
