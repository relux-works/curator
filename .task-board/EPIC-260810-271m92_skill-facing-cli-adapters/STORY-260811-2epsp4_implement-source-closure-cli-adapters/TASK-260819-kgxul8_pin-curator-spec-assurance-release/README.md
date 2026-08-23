# TASK-260819-kgxul8: pin-curator-spec-assurance-release

## Description
Advance Curator to the exact published curator-spec assurance release after portable implementation and release publication.

## Scope
Update only authoritative protocol version and conformance-suite pins, generated fixtures or compatibility mappings required by the published release, and release-facing documentation. Reject mutable or mismatched pins and preserve older supported release semantics.

## Acceptance Criteria
Curator consumes the exact released version and manifest digest, rejects mismatches, keeps legacy compatibility intentional, passes protocol and full repository gates, and independent review confirms there is no unreviewed semantic drift.
