# STORY-260819-2khabz: verified-macos-provider

## Description
Deliver the signed macOS verified provider using Apple-supported system-extension mechanisms.

## Scope
Endpoint Security for authoritative process and file authorization or observation plus Network Extension for required network-flow controls, with signed installation, entitlements, authenticated IPC, lifecycle, and platform conformance. Exact mechanisms remain gated by research and Apple entitlement availability.

## Acceptance Criteria
On supported macOS versions, verified mode establishes the common required guarantees with documented Apple primitives, starts no workload on missing entitlement or provider failure, produces provider-bound receipts, passes adversarial and lifecycle conformance, and ships as a signed separately installed Curator component.
