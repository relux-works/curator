# TASK-260910-dufdai: implement-source-aware-build-receipts-and-cache

## Description
Implement source aware build receipts and cache. Extend current implementation after inspecting existing outcomes; this task is not authorization to start work.

## Scope
internal/buildcache, buildsource, buildrepo, godriver and current supported build driver integration.

## Acceptance Criteria
Implement receipt v3 package identity wrapper/cache key for both local and external build arms. Preserve all existing declared/effective identities, locked commits, targets, substitutions and assurance. Preserve toolchain readiness/failure behavior; runtime/build input mutations invalidate required state; no prebuilt download or compiler bootstrap added. Test all external-evidence field mismatches and both arms.
