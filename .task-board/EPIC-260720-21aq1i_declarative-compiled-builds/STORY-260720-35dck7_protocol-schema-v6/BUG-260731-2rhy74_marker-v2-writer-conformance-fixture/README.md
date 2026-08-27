# BUG-260731-2rhy74: marker-v2-writer-conformance-fixture

## Description
The CocoaSkills PR 16 CI exposes a protocol-suite drift: current writers MUST emit install marker v2 for schema 1 through 6, but tests/test_protocol_conformance.py compares writer output to conformance/v1/expected/marker.json, which is intentionally frozen marker v1 legacy-read evidence. Add a distinct generated canonical marker-v2 writer fixture to curator-spec without changing the legacy marker-v1 fixture; update conformance manifest/release candidate inventory and CocoaSkills consumer test to use the new fixture. Preserve exact legacy-read compatibility and fail closed if the new fixture is absent.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
curator-spec retains expected/marker.json byte-identical as marker-v1 legacy evidence; a separately named canonical expected marker-v2 fixture represents the schema-5 golden skill with empty build_roots/builds and no build_source; generator, manifest hashes, validator tests, regenerate-check and rc.6 release-check pass; CocoaSkills protocol conformance reads the new fixture directly and compares byte-semantic writer output; local focused/full/mypy gates pass; PR CI passes on macOS, Linux and Windows.
