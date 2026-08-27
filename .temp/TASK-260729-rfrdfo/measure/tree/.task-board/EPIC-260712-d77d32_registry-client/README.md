# EPIC-260712-d77d32: registry-client

## Description
Phase 7 of docs/implementation-plan.md. The audit registry client: canonical bytes, Ed25519 verification against pinned keys, artifact matching, deny-wins federation with the warning taxonomy, snapshot verification with persisted monotonic versions and staleness bounds, cache TTL and offline grace, install wiring, attest re-check, publish (Spec 13).

## Scope
See docs/implementation-plan.md, Phase 7. Stories and tasks are created under this epic as work starts.

## Acceptance Criteria
- httptest registry with signed fixtures: audited, revoked, deprecated, tampered, rollback, freeze, offline
- Verified revocation always denies; strict policy fails unknown
- Attestation lands in the install marker
