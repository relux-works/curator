# TASK-260728-2yxdo7: implement-rust-v1-local-builds-in-csk

## Description
Independently implement local rust-v1 vendored skill builds in Python csk/CocoaSkills.

## Scope
Local build-root planning and context exclusion, trusted toolchain preflight, fixed offline worker, artifact cache/receipt/marker integration, atomic activation, rollback, status/repair/GC and macOS/Windows tests.

## Acceptance Criteria
csk matches Curator and shared vectors for valid local builds, forbidden inputs, missing toolchains, cache and rollback while preserving existing script and Go behavior; pytest, mypy and lint gates pass.
