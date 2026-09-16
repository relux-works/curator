# Unreleased draft Skillfile parser fixtures

Copied without edits from curator-spec at 3535d63ea80f97bba2fcb6e1f06996cfc25cf7df
(the source contract landed in a4fcaf02):

- `conformance/draft-sources-v1/schema-cases/skillfile-v2`: all 41 cases.
- `schemas/draft-sources-v1/skillfile-v2.schema.json` and its frozen v1 references.

`TestDraftPublishedSchemaCases` drives every published case through
`manifest.LoadWithOptions`. Filename labels are the published schema oracle.
The local schemas document that oracle; Go tests do not execute a JSON Schema
engine. Semantic cases (unknown aliases, canonical spelling, duplicates and
ref bounds) additionally drive the byte parser in sources_test.go.

This evidence covers parsing only. It does not establish source containment,
collection enumeration, lock/snapshot integrity, transport policy, installation,
or cross-platform filesystem behavior. Default reader APIs still reject v2.
