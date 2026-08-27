# Consume shared schema v6 vectors

## Description
Extend the Python conformance harness as an independent consumer of the protocol rc.4 schema, build-driver, identity, cache, context, and manager-lifecycle vectors.

## Scope
Own tests/test_protocol_conformance.py plus focused vector adapters and fixtures that do not duplicate product logic. Consume the exact candidate suite through caller-supplied CURATOR_CONFORMANCE_ROOT after TASK-260720-3ag6pi; do not edit the committed curator-spec checkout ref in this task. Cover generated schema cases for both manifest names, receipt v1, marker v2, claim v2, build-drivers.json, and manager-lifecycle.json. The official manager suite pin remains on the previous released protocol until TASK-260720-25d05o qualifies the schema v6 release and TASK-260720-1utsx8 audits the promotion. Do not import or copy Curator implementation code and do not fabricate release or conformance claims.

## Acceptance Criteria
The Python implementation passes shared positive vectors for schema 6, build-root context exclusion, build-source and toolchain bytes, CCJ-1 input and key, canonical receipt and marker, fixed environment and all five Go argv forms, hit and dry-run plans, ordering, commit, rollback, recovery, status, repair, and GC. Named negative coverage includes every minimum cluster from the accepted contract: unsafe declarations and paths, dependencies and Go graph inputs, toolchain and telemetry, poisoned environment, cache corruption and forged provenance, legacy marker-embed identity, dry-run mutation, concurrent publication, multi-project isolation, and claim-version separation. Legacy rc.3 vectors remain green. Tests against the exact caller-supplied candidate suite and strict mypy pass, the candidate revision and suite digest are recorded as non-release evidence, and no committed curator-spec pin or release claim changes.
