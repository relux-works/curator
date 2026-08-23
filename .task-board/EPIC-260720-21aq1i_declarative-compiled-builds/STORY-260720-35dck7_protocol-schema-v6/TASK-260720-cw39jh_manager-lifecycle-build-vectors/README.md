# Generate compiled-build manager lifecycle vectors

## Description
Extend the generated manager lifecycle suite with shared installation semantics for compiled commands. The vectors must make audit-before-build, provider and command order, dry-run non-mutation, private build staging, protected publication, cross-project isolation, rollback, recovery, currentness, repair, and garbage collection independently testable.

## Scope
Work in curator-spec after TASK-260720-1s1vr6 so edits to tools/generate-vectors/main.go are serialized. Own the compiled-build additions to conformance/v1/vectors/manager-lifecycle.json and corresponding generator tests. Reuse logical keys and receipt fixtures from build-drivers.json rather than defining incompatible identities. Do not edit normative prose or release docs.

## Acceptance Criteria
Vectors require full snapshot and manifest validation, closure and collision planning, source audit, registry and attestation gates before any build; active build order is provider-first with lexical command order; dry-run executes no go list or go build and mutates no audit, registry, probe, cache, lock, journal, runtime, context, marker, shim, adapter, consumer, or GC state; build two failure after build one leaves all persistent state unchanged; cache race, corruption, untrusted boundary, identical winner, and determinism mismatch outcomes are explicit; two-project success preserves both consumers and success-versus-rollback preserves the successful project; deterministic lock and target order, interrupted global-journal recovery, consumer-last commit, reverse rollback, status/currentness, repair rebuild, and locked GC are covered; go test ./tools/generate-vectors, make regenerate, and make regenerate-check pass deterministically.
