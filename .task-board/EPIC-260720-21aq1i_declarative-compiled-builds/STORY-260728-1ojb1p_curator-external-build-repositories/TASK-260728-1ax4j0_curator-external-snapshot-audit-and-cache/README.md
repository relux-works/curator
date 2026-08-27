# TASK-260728-1ax4j0: curator-external-snapshot-audit-and-cache

## Description
Integrate admitted external snapshots as independent audit subjects and protected snapshot/artifact cache inputs. Enforce validation, hashing, policy, audit, and cache ordering for install, dry-run, cache-hit, repair, and coverage-claiming audit operations.

## Scope
External snapshot materialization, build-source digest and declared/effective identity, audit request/result binding, protected cache admission and corruption handling, go-repository-v1 compiler-visible build root containment, deterministic Go build reuse, and tests. Transaction publication and marker lifecycle are downstream.

## Acceptance Criteria
The whole external repository snapshot is validated, hashed, and audited independently from the consuming skill before artifact-cache lookup or compiler execution; only descriptor-selected build_root/source_dir is compiler-visible; inaccessible exact source fails mutating and coverage operations before mutation while an already protected exact audited snapshot supports specified offline reinstall; substitutions cannot alias declared cache keys; cache hits repeat source validation/audit and corrupt entries quarantine/fail safely; go-repository-v1 reuses the closed go-v1 toolchain session without changing go-v1 receipts; tests prove ordering and containment.
