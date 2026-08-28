# EPIC-260819-2ats6u: verified-execution-providers

## Description
Design, implement, package, sign, and validate verified Curator execution providers for macOS, Linux, and Windows behind one platform-neutral assurance contract.

## Scope
Future delivery after portable adapters. Build separately installed privileged host providers and the unprivileged Curator client integration. Cover provider discovery, authenticated IPC, capability negotiation, enforcement and lossless observation semantics, health and lifecycle, signing and updates, receipts and attestation, fail-closed behavior, and cross-platform conformance. Provider binaries are never vendored inside skills. Verified admission of third-party precompiled skill artifacts remains a separate future capability.

## Acceptance Criteria
All three operating systems have independently reviewed verified providers satisfying the same normative capability contract and platform-specific threat model. Explicit verified mode fails before process start on missing, unhealthy, incompatible, downgraded, or unverifiable providers; emits non-forgeable provider-bound receipts; prevents cross-provider and cross-mode cache aliasing; survives lifecycle and fault-injection tests; and passes a shared cross-platform conformance and release gate.
