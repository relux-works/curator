# TASK-260728-12pnm1: design-rust-repository-v1-security-contract

## Description
Design the closed local rust-v1 and external rust-repository-v1 driver pair and prove which Cargo/rustc behaviors can be admitted under the portable manager-worker policy.

## Scope
Shared trusted toolchain, manifest requirements and Cargo metadata; local context-excluded build roots versus external skill-build.json targets; Cargo.lock/vendor/config grammar; workspaces/features; build.rs, proc macros, plugins, target config, flags, linker, FFI, bindgen, native libraries, network and registries; fixed process graph and deterministic artifacts.

## Acceptance Criteria
The reviewed contract defines both source modes and their distinct identities with one exhaustive allow/reject matrix and manager-owned pipeline, rejects or controls every package-selected process/native input without deferred hardened guarantees, and is implementation-ready on macOS/Windows with Linux qualification rules.
