# BUG-260920-2eg8nv: marker-reader-admits-external-record-without-substituted

## Description
From the 1xs0pj and 1xya7x reviews (schema bound invalid-external-missing-substituted): marker.Read admits an external go-repository-v1 build record without the substituted field although the v5 schema requires it; also valid.json from the spec corpus is refused (spec-corpus quirk to be reported upstream rather than worked around).

## Scope
internal/marker v5 external build record validation; conformance schema rows

## Acceptance Criteria
invalid-external-missing-substituted refused by marker.Read (schema row flips to driven-pass); the valid.json quirk reported to curator-spec with the exact reason (bound documented if the corpus is wrong); narrowing mutant killed.
