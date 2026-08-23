# Portable and verified assurance split

## Decision

Curator exposes two non-aliasing execution assurance modes.

- portable is the default and requires only the Curator CLI. It admits immutable source inputs, selects immutable toolchain identities, applies every portable control available by contract, verifies declared outputs, and emits a typed receipt containing only capabilities actually established. It does not claim lossless observation of every process, file, or network attempt.
- verified is explicit and requires a separately installed, signed, authenticated, healthy host provider satisfying the requested capability version. Missing, partial, stale, downgraded, or unhealthy providers reject before workload start. Curator never silently falls back to portable.

Policies may require a minimum assurance mode and capabilities. A portable receipt can never satisfy a verified requirement. Cache, checkpoint, permit, execution, observation, and publication identities include assurance mode and provider-contract identity so evidence cannot alias across modes.

Provider binaries are host trusted components, not skill dependencies. The global prohibition on vendored compiled artifacts remains in force for Go, Rust, Node/TypeScript, Swift, C, C++, Objective-C, and Objective-C++ closures. Third-party precompiled artifact admission is out of scope.

Current delivery implements portable plus Rust, Node/TypeScript, and SwiftPM C-family adapters. Kotlin, Dart, and .NET remain deferred. Future verified delivery is EPIC-260819-2ats6u.
