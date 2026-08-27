# TASK-260728-2u5u14: shared-rc5-external-repository-cases

## Description
Publish the shared rc.5 external-build-repository interop corpus after both managers have accepted implementations. Define byte-exact fixtures and expected typed outcomes for canonical source identity, raw snapshots, declared/effective substitutions, receipts, markers, mixed commands, failures, caches, lifecycle, and platform claims.

## Scope
Shared fixture repositories and bundles, canonical JSON/vector files, expected hashes and bytes, adversarial Git stores, case manifest, and reproducible fixture generation. Manager-specific harness adapters and release promotion are downstream.

## Acceptance Criteria
Cases cover SHA-1/SHA-256, tagged/untagged match/moved/missing, HTTPS/SSH/local identities, monorepo targets, substitutions, raw object and LFS/submodule/link/special/alternate/replace/graft/promisor/helper/filter negatives, audit ordering, cache hit/miss/corruption/offline reuse, mixed receipt-v1/v2 and marker-v3, status/repair/GC, PATH/shim/collision/rollback, signing boundaries, and truthful platform claims; fixture generation is deterministic and independently reviewed; both manager teams can consume the corpus without implementation-specific assumptions.
