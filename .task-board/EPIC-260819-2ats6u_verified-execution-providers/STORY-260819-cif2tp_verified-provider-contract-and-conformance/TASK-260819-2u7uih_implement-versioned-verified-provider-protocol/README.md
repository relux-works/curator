# TASK-260819-2u7uih: implement-versioned-verified-provider-protocol

## Description
Implement the common Curator to provider protocol and provider SDK boundary.

## Scope
Provider discovery, version and platform identity, mutual authentication, session challenge, capability negotiation, workload permit, atomic start, event and observation stream, final receipt, cancellation, health, revocation, upgrade compatibility, and stable diagnostics. Bind provider binary identity and policy to cache and receipt identities.

## Acceptance Criteria
Fake, stale, downgraded, replayed, cross-host, cross-provider, capability-mismatched, disconnected, or unhealthy providers start no workload or invalidate it fail closed. Protocol fuzz, replay, concurrency, lifecycle, and compatibility tests pass.
