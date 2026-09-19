# Draft source conformance corpus (vendored, pinned)

Vendored without edits from curator-spec at
`802caee548ddc8b19408746d26c7972d39b39cc2`
(`git archive` of `conformance/draft-sources-v1` and
`schemas/draft-sources-v1`), via `DRAFT_SOURCES_PIN`.
`MANIFEST.sha256` pins every vendored byte; the pin test fails on any
drift, addition, or removal.

Measured counts at this pin: 115 schema index entries, 94 semantic
cases, 3 snapshot vectors.

The task acceptance names 102 schema cases and 73 semantic cases: that
is the accepted-contract corpus at `a4fcaf0`, and it is an unchanged
subset of this pin (the delta `a4fcaf0..802caee` is purely additive:
13 `source-policy-v2` schema cases, the `source-policy-v2` schema, and
21 `v2-*` transport-revision-2 semantic cases; `snapshot-cases.json` is
byte-identical). `ac-schema-cases.txt` (102 instance paths) and
`ac-semantic-cases.txt` (73 ids) record that subset; the harness
asserts every listed case is present and has a row.

Go tests do not execute a JSON Schema engine: the vendored schemas
document the oracle, and every schema case is driven through the
production reader the manager actually uses (see
`draftsources_schema_test.go`). Cases with no byte-reader production
entry are explicit bounds, never passes.
