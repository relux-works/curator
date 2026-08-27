# TASK-260728-3lqm4z: verify-swift-v1-local-cross-manager-interop

## Description
Verify local vendored swift-v1 behavior across Curator and csk on supported macOS and Windows toolchains.

## Scope
Skill-contained package/vendor fixtures, context exclusion, artifact and receipt equality, toolchain/SDK diagnostics, rejected plugin/macro/native cases, rollback/concurrency and PATH launch.

## Acceptance Criteria
Both managers match the complete local Swift corpus and produce reproducible native evidence; unsupported tuples remain honest errors with no claims.
