# TASK-260819-3kwd8g: specify-portable-and-verified-assurance-modes

## Description
Update authoritative curator-spec with a common assurance model that keeps portable implementable now and verified provider-backed later.

## Scope
Define closed mode identities, provider capability negotiation, policy requirements, fail-closed selection, no silent downgrade, typed capability and execution receipts, cache and checkpoint separation, compiled-artifact prohibition, platform-neutral verified provider contract, migration and compatibility rules. Add normative schemas, valid and invalid vectors, decision record, operator guidance, generators, validators, and release-gate coverage. Reuse the hardened-profile draft only as reviewed input; do not inherit Linux-only assumptions.

## Acceptance Criteria
Portable is the default CLI-only policy and records only guarantees actually established. Verified is explicit, requires a healthy compatible provider, and fails before execution otherwise. Portable evidence cannot satisfy verified requirements; identities cannot alias across modes or providers. Host provider binaries are separately installed trusted components and never skill-vendored artifacts. macOS, Linux, and Windows share one versioned provider/capability/receipt contract. All spec validation and regeneration gates pass and independent review accepts the normative change.
