# Draft source conformance corpus (vendored, pinned)

Vendored without edits from curator-spec at
`dcc7f015e2d97edf2d52928afb6fd79ec8129e8b`
(`git archive` of `conformance/draft-sources-v1` and
`schemas/draft-sources-v1`), via `DRAFT_SOURCES_PIN`.
`MANIFEST.sha256` pins every vendored byte; the pin test fails on any
drift, addition, or removal.

Measured counts at this pin: 116 schema index entries, 94 semantic
cases, 3 snapshot vectors.

The task acceptance names 102 schema cases and 73 semantic cases: that
is the accepted-contract corpus at `a4fcaf0`, and it is a subset of this pin.
Since the earlier `802caee` corpus, curator-spec added the valid
`install-marker-v5/valid-no-builds.json` case and corrected
`install-marker-v5/valid.json` so its empty `builds` value omits
`build_source`; the snapshot vectors remain byte-identical. The earlier
source-policy-v2 and transport-revision-2 additions remain included.
`ac-schema-cases.txt` (102 instance paths) and
`ac-semantic-cases.txt` (73 ids) record that subset; the harness
asserts every listed case is present and has a row.

Go tests do not execute a JSON Schema engine: the vendored schemas
document the oracle, and every schema case is driven through the
production reader the manager actually uses (see
`draftsources_schema_test.go`). Cases with no byte-reader production
entry are explicit bounds, never passes.
