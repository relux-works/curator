# TASK-260728-16kefa: verify-rust-cross-manager-and-native-interop

## Description
Run the black-box rust-repository-v1 suite against Curator and csk and qualify matching macOS and Windows behavior.

## Scope
Valid fixture builds, exact artifacts and receipts, cache hits, inaccessible sources, forbidden build.rs/proc-macro/plugin/native-link cases, toolchain drift, rollback injection, concurrent installs, PATH launch and claim evidence on ssh relux and ssh win.

## Acceptance Criteria
Both managers return matching bytes, warnings, errors and installed command behavior for the complete Rust corpus; native macOS and Windows evidence is immutable and reproducible; no tuple is claimed when a gate is missing or red.
