# TASK-260728-17sclp: external-repository-wire-schemas-and-generated-cases

## Description
Add the schema-7 wire model and generated cases for external build repositories: manifest build_repositories and go-repository-v1 commands, curator-build.json schema 1, Skillfile.dev schema 2 substitutions, receipt schema 2, marker schema 3, and conformance claim schema 3.

## Scope
JSON schemas, canonical examples, generated valid/invalid cases, schema registry/version selection, and legacy compatibility guards in curator-spec. No manager implementation or release promotion.

## Acceptance Criteria
All new schemas are closed with conditional branches and reject unknown or package-controlled execution fields; immutable commit/object-format, URL/ref/tag grammar, logical target/build_root/source_dir containment, declared/effective substitution state, mixed receipt-v1/v2 marker-v3 entries, and claim-v3 platform/language assertions are represented exactly; schema 1-6 fixtures remain byte-valid and reject schema-7-only fields; generators are deterministic and all schema tests pass on clean regeneration.
