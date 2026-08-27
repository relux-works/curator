# Add conformance claim v2 for protocol rc.4

## Description
Create and generate the explicit conformance-claim transition required by schema 6. Claim v1 remains immutable rc.3 evidence. Protocol rc.4 writers use claim v2 and readers dispatch by schema version. This task also moves the generated suite identity and validator version assertion to rc.4 while keeping the repository green.

## Scope
Work in curator-spec after TASK-260720-12iigs. Own new conformance-claim-v2.schema.json, claim-related and version constants in tools/generate-vectors/main.go and main_test.go, generated v2 cases and index entries, the frozen v1 case constant, generated manifest version and hashes, and only the protocol-version assertion in tools/validate.py. Do not edit claim-v1 schema, release prose, broader validation inventory, or release-gate logic.

## Acceptance Criteria
Claim v2 strictly fixes schema_version 2 and protocol_version 1.0.0-rc.4 with the existing required claim fields and allowed sets; generated v2 cases reject rc.3, schema version 1, duplicate classes, fail result, missing and unknown fields; claim-v1 schema and generated valid and invalid instances remain rc.3 and byte-semantically unchanged; generator separates protocolVersion rc.4 from conformanceClaimV1ProtocolVersion rc.3; generated manifest and validator both identify rc.4; go test ./tools/generate-vectors, make regenerate, make validate, and make regenerate-check pass.
