# TASK-260728-wy3dsw: external-repository-manager-profile-and-lifecycle-contract

## Description
Specify the rc.5 manager profile and CLI lifecycle for resolving, validating, auditing, compiling, installing, reporting, repairing, and collecting external build repositories. Pin the exact clean Git process, raw-object proof, LFS and special-file rejection, snapshot/cache identity, mixed planning, transaction, PATH/shim verification, offline semantics, and status/repair/GC behavior.

## Scope
curator-spec manager-profile prose, CLI contracts, lifecycle state machine, stable diagnostics, and normative command examples. Reuse the accepted go-v1 trusted toolchain/build session without changing its schema-6 meaning.

## Acceptance Criteria
The profile fixes supported transports and canonical identity, exact init/fetch/ref flows, SSH isolation, local-substitution admission, pack/index and cat-file grammar, commit/tag/tree/blob recomputation, complete-object and all-blob LFS proof, audit ordering, cache and receipt keys, marker currentness, rollback, repair and GC roots; tagged declarations always exact-fetch the tag as the sole path; syntax-only offline warning and install/audit failure are unambiguous; no repository-controlled hooks, helpers, filters, argv, env, output, credentials, signing, or lazy network reads remain possible; profile examples and validation pass.
