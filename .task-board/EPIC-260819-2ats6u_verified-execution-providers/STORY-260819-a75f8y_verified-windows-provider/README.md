# STORY-260819-a75f8y: verified-windows-provider

## Description
Deliver a signed Windows verified provider using supported kernel and service isolation mechanisms.

## Scope
Select and implement the required composition of Windows service, process containment, filesystem and registry filtering, network filtering, telemetry, code signing, authenticated IPC, installation, updates, and platform conformance. Candidate primitives include Job Objects, AppContainer or sandbox boundaries, minifilter, WFP, ETW, and WDAC where justified by the threat model.

## Acceptance Criteria
On the declared Windows support matrix, verified mode establishes every required capability, rejects missing drivers, policy, signing, or partial enforcement before workload start, produces provider-bound receipts, passes adversarial and lifecycle conformance, and ships as a signed separately installed Curator component.
