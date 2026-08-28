# STORY-260819-cif2tp: verified-provider-contract-and-conformance

## Description
Define and implement the platform-neutral verified provider boundary, authenticated transport, capability model, receipt binding, test harness, and lifecycle contract.

## Scope
Curator core and provider SDK or harness only. No platform backend implementation. Align with the released curator-spec assurance contract and keep portable behavior independent.

## Acceptance Criteria
A versioned provider protocol supports macOS, Linux, and Windows without platform claims leaking into the common contract; authentication, capability negotiation, fail-closed health, receipt binding, upgrades, revocation, and fault injection are covered by independent conformance tests.
