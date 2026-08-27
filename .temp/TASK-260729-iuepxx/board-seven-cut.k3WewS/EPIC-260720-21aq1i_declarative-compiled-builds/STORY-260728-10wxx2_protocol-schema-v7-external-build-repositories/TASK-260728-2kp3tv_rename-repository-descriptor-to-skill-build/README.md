# TASK-260728-2kp3tv: rename-repository-descriptor-to-skill-build

## Description
Replace the implementation-branded curator-build.json repository descriptor with the manager-neutral skill-build.json name throughout the unreleased schema-7 and rc.5 candidate contract.

## Scope
Rename the descriptor file, schema artifact, registry entries, normative paths, receipt-v2 and marker-v3 bindings, generated cases, vectors, documentation, release metadata and candidate digest. Remove the old name entirely; do not add aliases or compatibility behavior because schema 7 is unreleased. Preserve schemas 1-6 byte-for-byte and all closed driver semantics.

## Acceptance Criteria
Every schema-7 producer and consumer recognizes only repository-root skill-build.json schema 1; curator-build.json is rejected and absent from normative/generated/release surfaces; command, target, output and driver ownership remain unchanged; deterministic regeneration, compatibility guards, Python and Go tests, release gates and independent digest recomputation pass; no release, pin, commit or platform claim is fabricated.
