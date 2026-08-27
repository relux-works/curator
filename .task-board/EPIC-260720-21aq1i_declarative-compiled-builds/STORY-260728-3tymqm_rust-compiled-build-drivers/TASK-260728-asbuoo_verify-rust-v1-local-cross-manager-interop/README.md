# TASK-260728-asbuoo: verify-rust-v1-local-cross-manager-interop

## Description
Verify local vendored rust-v1 behavior across Curator and csk on macOS and Windows.

## Scope
Skill-contained fixture source and vendor tree, context-exclusion proof, deterministic artifacts/receipts, cache hit/miss, toolchain diagnostics, forbidden Cargo cases, rollback injection, concurrent install and PATH launch.

## Acceptance Criteria
Both managers produce matching bytes and semantic results for the complete local Rust corpus; build roots never enter runtime/context; native macOS/Windows evidence is reproducible and claims remain empty on any red gate.
