# TASK-260728-gjxj1v: implement-rust-repository-v1-in-csk

## Description
Independently implement the accepted rust-repository-v1 contract in Python csk/CocoaSkills with byte-compatible wire and security behavior.

## Scope
Python schema and plan models, trusted Rust toolchain and worker orchestration, offline source/dependency validation, artifact cache and receipts, atomic activation, rollback, status/repair/GC, macOS/Windows tests and regression coverage for existing scripts and Go drivers.

## Acceptance Criteria
csk matches Curator and the shared vectors for every valid, cache, offline, forbidden-build and rollback case without invoking Curator internals or arbitrary Cargo commands; pytest, mypy, lint and native macOS/Windows gates pass.
