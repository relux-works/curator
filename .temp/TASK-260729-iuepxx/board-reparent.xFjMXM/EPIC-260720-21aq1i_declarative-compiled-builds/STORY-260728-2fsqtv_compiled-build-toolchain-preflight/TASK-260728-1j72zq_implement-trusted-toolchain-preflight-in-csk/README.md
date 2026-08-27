# TASK-260728-1j72zq: implement-trusted-toolchain-preflight-in-csk

## Description
Independently implement the accepted toolchain requirement and manager-owned guidance contract in Python csk/CocoaSkills.

## Scope
Python models and canonical version comparison, trusted locators and probes, metadata cross-checks, cache identity, dry-run/status behavior, guidance catalog, macOS/Windows tests and regression coverage.

## Acceptance Criteria
csk matches Curator and shared vectors for all resolution and diagnostic cases, never executes an installer or package-provided guidance, preserves existing flows, and passes pytest, mypy and lint gates.
