# TASK-260728-3j60e3: implement-swift-v1-local-builds-in-csk

## Description
Independently implement local swift-v1 vendored skill builds in Python csk/CocoaSkills.

## Scope
Local build-root/context exclusion, trusted Swift toolchain and SDK preflight, fixed offline worker, cache/receipt/marker integration, atomic activation, rollback/currentness and native tests.

## Acceptance Criteria
csk matches Curator and shared local Swift vectors without arbitrary package commands, preserves existing drivers, and passes pytest, mypy, lint and supported native gates.
