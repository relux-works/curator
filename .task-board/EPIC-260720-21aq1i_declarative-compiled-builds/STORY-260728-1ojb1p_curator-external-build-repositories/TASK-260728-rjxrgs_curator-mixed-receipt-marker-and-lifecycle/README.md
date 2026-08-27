# TASK-260728-rjxrgs: curator-mixed-receipt-marker-and-lifecycle

## Description
Extend Curator planning and transactional lifecycle for mixed schema-7 installations. Produce receipt v2 for external commands, preserve receipt v1 for local go-v1 commands, publish marker v3, derive and verify shims/PATH entries, and implement currentness, repair, rollback, recovery, deduplication, and GC.

## Scope
Side-effect-free planner, receipt/cache key canonicalization, marker-v3 serialization, project/global transaction staging and commit, command collision checks, manager-derived bin paths and shims, signer-policy fail-closed gate, status/repair/GC and crash recovery. Release notarization and Authenticode are out of scope.

## Acceptance Criteria
Local and external commands coexist without receipt interpretation aliasing; receipt v2 binds declared/effective source, substitution state, exact commit, digests, target, toolchain, and policy; marker v3 structurally represents local-only, external-only, and mixed installs; publication occurs only after exact-tag/source proof, audit, build, and artifact validation; command names and output paths are manager-derived, executable/hash/link-count/shim/PATH checks are enforced, collisions fail before mutation, and rollback restores prior live state; status is read-only, repair reacquires/audits safely, and GC keeps only journal/marker roots; signing credentials never enter package inputs; tests pass.
