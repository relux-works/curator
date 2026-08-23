# Enforce schema v6 validation and release gates

## Description
Teach the specification validator and release gate to require the full protocol rc.4 artifact set and to fail closed when any schema, generated case, vector, decision, claim transition, or manifest entry is absent, stale, or inconsistent.

## Scope
Work in curator-spec after all schema and vector generation tasks. Own tools/validate.py, tools/release_gate.py, tools/test_release_gate.py, any directly related validator tests, and generator manifest inventory assertions. Validate both new manifest names, receipt v1, marker v2, claim v2, build-drivers and manager-lifecycle vectors, decision 0004, release version identity, and current suite SHA semantics. Do not write release prose in this task.

## Acceptance Criteria
Validation discovers and resolves every new schema and validates all indexed positive and negative cases; it asserts conformance/v1/manifest.json protocol_version 1.0.0-rc.4 and complete deterministic file hashes; release gate requires decision 0004, both v6 manifest schemas, receipt v1, marker v2, claim v2, build-driver vectors, lifecycle coverage, and claim v1 frozen at rc.3; tests fail when each required artifact or claim version is removed, renamed, stale, or mismatched and reject an rc.3 suite hash presented as rc.4 evidence; python3 -B -m unittest discover -s tools -p test_*.py, go test ./tools/..., make validate, and make regenerate-check pass.
