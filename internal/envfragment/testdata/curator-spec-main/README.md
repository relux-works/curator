# Curator spec main fixture

Copied byte-for-byte from the local curator-spec checkout at commit
`eadb1c06480f4438f01b2b0a973caf775a188f41` (main, after `ec8dc656`).

- Schema source: `schemas/v1/launch-env-fragment-v2.schema.json`
- Fixture path: `internal/envfragment/testdata/curator-spec-main/schemas/v1/launch-env-fragment-v2.schema.json`
- SHA-256: `4d26b5e2a89452eb3c7fe6945f19dd16dda557382be15972ab9c57cb02d38d40`
- Conformance cases: `conformance/v1/schema-cases/launch-env-fragment-v2/`

The test registers the copied external schema dependencies locally and validates
both the copied valid/invalid cases and `curator env resolve --format json`
output. It does not use curator's pinned rc.12 conformance checkout for v2.
