# Add build receipt v1 and install marker v2 schemas

## Description
Define the portable compiled-artifact metadata schemas and generate their conformance cases. The receipt binds the exact logical build input, target, toolchain, fixed policy, and derived artifact. Marker v2 records build roots, raw build-source identity, sorted build records, keys, receipt hashes, and artifact paths while marker v1 remains intact.

## Scope
Work in curator-spec after TASK-260720-wajgn8. Own new build-receipt-v1.schema.json and install-marker-v2.schema.json, supporting common definitions, their portions of tools/generate-vectors/main.go and main_test.go, generated schema cases and index entries, and resulting manifest hashes. Keep install-marker-v1.schema.json untouched. Physical cache and lock paths are not schema fields.

## Acceptance Criteria
Receipt v1 strictly requires schema version, cache key, full go-v1 input with curator-build-source-v1, build_root, command, source_dir, native target and tuning, curator-go-toolchain-v1, fixed policy, and one derived artifact path, hash, and size; receipt fields cannot self-assert provenance; marker v2 supports skill schema through 6, requires sorted build_roots and builds, requires build_source exactly when builds are non-empty, and records driver, key, receipt hash, artifact hash, and path per command; generated cases exercise required, conditional, unknown, and mismatched fields; marker v1 remains byte-semantically unchanged; go test ./tools/generate-vectors, make regenerate, make validate, and make regenerate-check pass.
